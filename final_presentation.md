## Background Information (Introductory Section)

We're creating a distributed cache in Go. Caching is important when you want to retrieve data for clients quickly and DB queries are too slow. Distributed caches are necessary when a single cache is being overloaded and you want to spread a large amount of requests over multiple cache nodes. Our goal was to simulate a distributed cache through multiple processes on one computer.

## Differences From Midterm Presentation

- Added UI visualizer to see all logging in one place
- Used pipes to display components' stdout in UI
- Fixed hashing in the cache rings to be more evenly distributed
- Added heartbeating so that client knows when a cache has crashed

## How we used the Raspberry Pi

(we didn't)

## Concepts From CS 341

1. Synchronization -- made cache and cache ring data strutures thread safe to support multiple threads reading/writing to/from them
2. Networking -- handled TCP connections between cache clients and servers; wrote distributed applications communicating across a network
3. Processes -- simulating cache clients and cache nodes through processes on one computer (for now)
4. Pipes -- used pipes to display logging in the UI

## Challenges We Faced

- Learning Go
- Figuring out the design for the cache and how to make it robust
- Setting up Supabase
- Making the caching evenly distributed among the cahces

## Tools Used

- Go
- Relevant Go networking packages (net, http)
- Supabase
- HTML/CSS/Javascript for frontend

## Addressing Questions

**What would CS341 be like if we used them for all the assignments?**
It would be easier because Go is easier than C.

**If you were in a job interview what would you say about this project?**
I would talk about how we tried to make it handle multiple clients and cache servers (scalable). I'd also talk about
handling concurrency.

## Reflection and Future Work

We liked learning Go a lot and felt like it was a pretty useful language to pick up. We also learned more about
how to design a system properly before actually coding.

Potential improvements to this include adding PUT and DELETE requests and adding data forwarding between the caches
to improve robustness.
