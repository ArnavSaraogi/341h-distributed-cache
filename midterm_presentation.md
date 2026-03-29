## Background Information

We're creating a distributed cache in Go. Caching is important when you want to retrieve data for clients quickly and DB queries are too slow. Distributed caches are necessary when a single cache is being overloaded and you want to spread a large amount of requests over multiple cache nodes.

## What We've Done So Far

We've implemented the basic caching and retrival from a distributed cache for GET requests.

(Diagram)[https://excalidraw.com/#json=beHIaW9eBQzF5A9KRSq1H,5HtlFnZ_RGS1sRn6Lp0lMw]

- Created the LRU cache and a wrapper for networking capability
- Created configuration service that maintains a list of active cache IPs
- Created hash ring in order to determine which cache holds which data
- Connected to Supabase DB for caches to retrieve values from based on key

## Concepts From CS 341

1. Synchronization -- made cache and cache ring data strutures thread safe to support multiple threads reading/writing to/from them
2. Networking -- handled TCP connections between cache clients and servers; wrote distributed applications communicating across a network
3. Processes -- simulating cacche clients and cache nodes through processes on one computer (for now)

## Challenges We Faced

- Learning Go
- Figuring out the design for the cache and how to make it robust
- Figuring out how TCP works
- Setting up Supabase

## Tools Used

- Go
- Relevant Go networking packages (net, http)
- Supabase
