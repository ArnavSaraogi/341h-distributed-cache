## Requirements

### Functional

- PUT(key, value)
- GET(key)

### Non-Functional

- Scales out easily with increasing number of requests and data
- Fault tolerant (survives hardware/network failures)
- High performance (fast puts and fast gets)

## Design

https://excalidraw.com/#json=beHIaW9eBQzF5A9KRSq1H,5HtlFnZ_RGS1sRn6Lp0lMw

### Cache Client

Server has a cache client responsible for communicating with the caches

- Uses TCP to communicate with cache
- All cache clients should have the same list of cache servers
- Cache client grabs list of clients from configuration service, which discovers cache hosts and monitors their health

### Cache Nodes

- Consistent hashing for sharding
  - Cache nodes will hold different data but there will be overlap
  - Client stores list of servers in sorted order by hash value
  - Binary search by hash is used to identify the cache server to store data in
  - Replication: cache node that got sent the PUT will forward to the next $r-1$ nodes, where $r$ is number of times data should be replicated
  - Virtual nodes: to ensure amount of keys each node handles is about the same, create virtual nodes
    - Have to make sure that on replication something like "A-58" isn't replicating to "A-30", has to be a virtual node that maps to a different real node
- LRU eviction policy
- Write back on eviction -- need to handle case where replicated caches have old data
- Read through on cache misses

### Configuration Service

The cache client gets it's list of active clients from the configuration service, which it polls every couple of seconds. Configuration service keeps track of health of cache nodes by pinging caches every couple of seconds (\health endpoint). Admin can manually add/kill nodes from here as well.

## Useful Videos

[System Design Interview - Distributed Cache](https://www.youtube.com/watch?v=iuqZvajTOyA)

- LRU eviction explanation/algorithm
- Consistent hashing

[What is Distributed Caching?](https://www.youtube.com/watch?v=C8eIaEBPmw8)

- Overview of distributed caching

### Project Files

```
distributed-cache/
├── go.mod
├── cmd/
│   ├── cache/
│   │   └── main.go
│   ├── config/
│   │   └── main.go
│   └── client/
│       └── main.go
├── lru/
│   └── lru.go            // LRU map, eviction; cache data structure (no networking stuff); need to make thread-safe?
├── node/
│   └── node.go           // TCP server, replication, wraps store and handles the networking stuff
├── ring/
│   └── ring.go           // consistent hashing, virtual nodes, replica selection
└── config_service
    └── config_service.go
```
