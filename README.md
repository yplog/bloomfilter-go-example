# Bloom Filter Go Example

This repository contains a simple yet efficient **Bloom filter implementation in Go**, using xxHash for fast hashing and a byte slice (`[]byte`) for memory-efficient bit storage.

---

## **What is a Bloom Filter?**

A Bloom filter is a **probabilistic data structure** that allows you to check whether an item is in a set, with:

- **No false negatives** (if it says no, it’s definitely not there)
- **Possible false positives** (it may say yes even if the item isn’t there)

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
go test -v

# Run benchmarks

go test -bench=. -benchmem
```