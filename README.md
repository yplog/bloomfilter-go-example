# Bloom Filter Go Example

This repository contains a simple yet efficient **Bloom filter implementation in Go**, using xxHash for fast hashing and a byte slice (`[]byte`) for memory-efficient bit storage.

---

## **What is a Bloom Filter?**

A Bloom filter is a **probabilistic data structure** that allows you to check whether an item is in a set, with:

- **No false negatives** (if it says no, it's definitely not there)
- **Possible false positives** (it may say yes even if the item isn't there)

Bloom filters are widely used in:

- Web crawlers (checking visited URLs)
- Databases and caches (avoiding unnecessary lookups)
- URL shorteners (checking if short codes exist)
- Network systems (fast membership checks)

---

## **Features**

Uses [xxHash](https://github.com/cespare/xxhash) for fast hashing  
Memory-efficient bit storage with `[]byte`  
Multiple hash functions via salted hashing  
Extensive tests covering:

- Creation with different sizes and hash counts
- Adding and checking items
- False positive rate measurement
- Edge cases (empty strings, long strings, Unicode)
- Benchmarks for Add, Check, and combined operations

---

## **Run tests**

```bash
go test -v bloomfilter_test.go bloomfilter.go

# Run benchmarks

go test -bench=. -benchmem bloomfilter_test.go bloomfilter.go
```

---

## **Redis Bloom Filter**

This repository also includes a **Redis-based Bloom Filter implementation** for distributed, persistent Bloom filters.

### **Setup Redis with Bloom Filter Module**

Use Docker to quickly set up Redis with the Bloom Filter module:

```bash
docker run -d -p 6379:6379 redis/redis-stack:latest
```

### **Run Redis tests**

```bash
# Run Redis Bloom Filter tests (requires Redis running)
go test -v bloomfilter_redis_test.go bloomfilter_redis.go

# Run Redis benchmarks
go test -bench=. -benchmem bloomfilter_redis_test.go bloomfilter_redis.go
```