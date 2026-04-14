**Necessary**

1. Have config service keep track of heartbeats to know when to remove cache from list
2. Make sure cache client is properly adapting to the client list updating and ring is rehashing necessary values
   - make UI so that when you add can choose what cache to add after deleting one so it doesn't readd the one you just deleted
3. Test and benchmark multiple clients sending requests very quickly (need to write script, interactive UI is too slow/unrealistic)

**Stretchables**

1. Have servers forward data to each other
2. Handle puts
3. Add more debug info to a different tab in the UI (cache ring diagram, caches contents, hash values, etc)
4. Virtual nodes for more evenly spread out load distribution
