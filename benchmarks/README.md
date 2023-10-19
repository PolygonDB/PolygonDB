
# Benchmarks

## Index

Goal:
* Read and Parse 5000+ Character JSON File
* With Index, get the index of all users whose gender are male.
* Use Search, get the index of the user whose GUID is ``147bd43a-338c-450b-a293-4999dba1f367``

Challenge:
* Use NodeJS/Python and Compare Results with PolygonDB

### Results (Python)
![PolygonDB vs Python (Seconds, Lower is Better)](https://cdn.discordapp.com/attachments/1077973116149563543/1164690319531581440/image.png?ex=65442163&is=6531ac63&hm=540d6f7dd7498506369315799cd8163199ea696dffd3b8bad3bfbc43f665c507&)

### Results (NodeJS)
```js
======================================
  Polygon vs NodeJS via File Reading  
======================================

Platform info:
==============
   Windows_NT 10.0.22621 x64
   Node.JS: 18.15.0
   V8: 10.2.154.26-node.25
   CPU: Intel(R) Core(TM) i5-8400 CPU @ 2.80GHz × 6
   Memory: 16 GB

Suite: Polygon vs NodeJS
========================

- Running 'Using PolygonDB'...
√ Using PolygonDB         4,999 ops/sec
- Running 'Using NodeJS'...
√ Using NodeJS               14 ops/sec

   Using PolygonDB      +36,758.81%      (4,999 ops/sec)   (avg: 200μs)
   Using NodeJS (#)           0%         (14 ops/sec)   (avg: 73ms)

┌─────────────────┬────────────────────────────────────────────────────┐
│ Using PolygonDB │ ██████████████████████████████████████████████████ │
├─────────────────┼────────────────────────────────────────────────────┤
│ Using NodeJS    │                                                    │
└─────────────────┴────────────────────────────────────────────────────┘
-----------------------------------------------------------------------
```
