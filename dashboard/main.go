package main

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

//go:embed index.html
var htmlContent embed.FS

// Process represents a managed subprocess
type Process struct {
	Name    string         `json:"name"`
	Type    string         `json:"type"` // "config", "cache", "client"
	Port    string         `json:"port,omitempty"`
	Cmd     *exec.Cmd      `json:"-"`
	Stdin   io.WriteCloser `json:"-"`
	LogCh   chan string    `json:"-"`
	Done    chan struct{}  `json:"-"`
	clients []chan string  // SSE subscribers
	mu      sync.Mutex
}

func (p *Process) addClient(ch chan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients = append(p.clients, ch)
}

func (p *Process) removeClient(ch chan string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, c := range p.clients {
		if c == ch {
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
			break
		}
	}
}

func (p *Process) broadcast(line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, ch := range p.clients {
		select {
		case ch <- line:
		default: // drop if client is slow
		}
	}
}

// ProcessManager manages all subprocesses
type ProcessManager struct {
	processes map[string]*Process
	mu        sync.Mutex
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

func (pm *ProcessManager) Start(procType, port string) (*Process, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var name string
	var args []string

	switch procType {
	case "config":
		name = "config-service"
		args = []string{"run", "cmd/config-service/main.go"}
	case "cache":
		if port == "" {
			return nil, fmt.Errorf("port required for cache")
		}
		name = "cache-" + port
		args = []string{"run", "cmd/cache/main.go", port}
	case "client":
		if port == "" {
			return nil, fmt.Errorf("port required for client")
		}
		name = "client-" + port
		args = []string{"run", "cmd/client/main.go", port}
	default:
		return nil, fmt.Errorf("unknown process type: %s", procType)
	}

	if _, exists := pm.processes[name]; exists {
		return nil, fmt.Errorf("%s is already running", name)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = projectRoot
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Capture stderr (where log output goes)
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	proc := &Process{
		Name:  name,
		Type:  procType,
		Port:  port,
		Cmd:   cmd,
		LogCh: make(chan string, 100),
		Done:  make(chan struct{}),
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Read stderr lines and broadcast
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			proc.broadcast(scanner.Text())
		}
	}()

	// Read stdout lines and broadcast
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			proc.broadcast(scanner.Text())
		}
	}()

	// Wait for process to exit
	go func() {
		cmd.Wait()
		proc.broadcast("[process exited]")
		close(proc.Done)
		pm.mu.Lock()
		delete(pm.processes, name)
		pm.mu.Unlock()
	}()

	pm.processes[name] = proc
	logger.Printf("Started %s (PID %d)\n", name, cmd.Process.Pid)
	return proc, nil
}

func (pm *ProcessManager) Stop(name string) error {
	pm.mu.Lock()
	proc, exists := pm.processes[name]
	pm.mu.Unlock()

	if !exists {
		return fmt.Errorf("%s is not running", name)
	}

	// Kill the process group so child processes (go run spawns a subprocess) also die
	pgid, err := syscall.Getpgid(proc.Cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		proc.Cmd.Process.Kill()
	}

	return nil
}

func (pm *ProcessManager) StopAll() {
	pm.mu.Lock()
	names := make([]string, 0, len(pm.processes))
	for name := range pm.processes {
		names = append(names, name)
	}
	pm.mu.Unlock()

	for _, name := range names {
		pm.Stop(name)
	}
}

func (pm *ProcessManager) List() []map[string]string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	result := []map[string]string{}
	for _, p := range pm.processes {
		result = append(result, map[string]string{
			"name": p.Name,
			"type": p.Type,
			"port": p.Port,
		})
	}
	return result
}

func (pm *ProcessManager) Get(name string) *Process {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.processes[name]
}

// NextAvailableCachePort finds the next free port in the cache range 8081-8099
func (pm *ProcessManager) NextAvailableCachePort() (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.nextPortInRange("cache", 8081, 8099)
}

// NextAvailableClientPort finds the next free port in the client range 7001-7019
func (pm *ProcessManager) NextAvailableClientPort() (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.nextPortInRange("client", 7001, 7019)
}

// nextPortInRange finds the next unused port for the given process type. Must be called with pm.mu held.
func (pm *ProcessManager) nextPortInRange(procType string, low, high int) (string, error) {
	used := map[int]bool{}
	for _, p := range pm.processes {
		if p.Type == procType {
			port, _ := strconv.Atoi(p.Port)
			used[port] = true
		}
	}
	for port := low; port <= high; port++ {
		if !used[port] {
			return strconv.Itoa(port), nil
		}
	}
	return "", fmt.Errorf("no available %s ports in range %d-%d", procType, low, high)
}

var (
	logger      = log.New(os.Stderr, "[DASHBOARD]: ", log.Ltime)
	pm          = NewProcessManager()
	projectRoot string
)

func main() {
	// Set project root to parent of dashboard/
	wd, _ := os.Getwd()
	if strings.HasSuffix(wd, "/dashboard") {
		projectRoot = strings.TrimSuffix(wd, "/dashboard")
	} else {
		projectRoot = wd
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/start", handleStart)
	http.HandleFunc("/api/stop", handleStop)
	http.HandleFunc("/api/stop-all", handleStopAll)
	http.HandleFunc("/api/processes", handleProcesses)
	http.HandleFunc("/api/logs", handleLogs)
	http.HandleFunc("/api/send", handleSend)
	http.HandleFunc("/api/next-port", handleNextPort)

	logger.Printf("Dashboard running at http://localhost:9000\n")
	logger.Printf("Project root: %s\n", projectRoot)

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data, _ := htmlContent.ReadFile("index.html")
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Type string `json:"type"`
		Port string `json:"port"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Auto-assign port if not specified
	if req.Port == "" {
		var err error
		switch req.Type {
		case "cache":
			req.Port, err = pm.NextAvailableCachePort()
		case "client":
			req.Port, err = pm.NextAvailableClientPort()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	proc, err := pm.Start(req.Type, req.Port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"name": proc.Name, "port": proc.Port})
}

func handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := pm.Stop(req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleStopAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	pm.StopAll()
	w.WriteHeader(http.StatusOK)
}

func handleProcesses(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(pm.List())
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	proc := pm.Get(name)
	if proc == nil {
		http.Error(w, "process not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 50)
	proc.addClient(ch)
	defer proc.removeClient(ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-proc.Done:
			return
		case line := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", line)
			flusher.Flush()
		}
	}
}

// handleSend forwards a command to a running client's HTTP endpoint
func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Command string `json:"command"`
		Client  string `json:"client"` // client name, e.g. "client-7001"
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Command == "" {
		http.Error(w, "command required", http.StatusBadRequest)
		return
	}

	// Find the target client
	if req.Client == "" {
		// Default to first available client
		pm.mu.Lock()
		for _, p := range pm.processes {
			if p.Type == "client" {
				req.Client = p.Name
				break
			}
		}
		pm.mu.Unlock()
	}

	if req.Client == "" {
		http.Error(w, "no client running — start a client first", http.StatusBadRequest)
		return
	}

	proc := pm.Get(req.Client)
	if proc == nil {
		http.Error(w, req.Client+" is not running", http.StatusBadRequest)
		return
	}

	// Forward command to the client's HTTP endpoint
	clientURL := fmt.Sprintf("http://localhost:%s/command", proc.Port)
	body, _ := json.Marshal(map[string]string{"command": req.Command})
	resp, err := http.Post(clientURL, "application/json", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "failed to reach client: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		http.Error(w, string(respBody), resp.StatusCode)
		return
	}

	// Pass through the client's response
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func handleNextPort(w http.ResponseWriter, r *http.Request) {
	procType := r.URL.Query().Get("type")
	var port string
	var err error
	switch procType {
	case "client":
		port, err = pm.NextAvailableClientPort()
	default:
		port, err = pm.NextAvailableCachePort()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"port": port})
}
