package main

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
)

type BloomFilter struct {
	size      uint64
	hashCount uint64
	bitArray  []byte
}

func NewBloomFilter(size, hashCount uint64) *BloomFilter {
	byteSize := (size + 7) / 8
	return &BloomFilter{
		size:      size,
		hashCount: hashCount,
		bitArray:  make([]byte, byteSize),
	}
}

func (bf *BloomFilter) getHashes(data string) []uint64 {
	hashes := make([]uint64, bf.hashCount)
	for i := uint64(0); i < bf.hashCount; i++ {
		h := xxhash.New()
		h.Write(fmt.Appendf(nil, "%s%d", data, i))
		hashes[i] = h.Sum64() % bf.size
	}
	return hashes
}

func (bf *BloomFilter) setBit(pos uint64) {
	byteIndex := pos / 8
	bitIndex := pos % 8
	bf.bitArray[byteIndex] |= 1 << bitIndex
}

func (bf *BloomFilter) getBit(pos uint64) bool {
	byteIndex := pos / 8
	bitIndex := pos % 8
	return (bf.bitArray[byteIndex] & (1 << bitIndex)) != 0
}

func (bf *BloomFilter) Add(data string) {
	for _, idx := range bf.getHashes(data) {
		bf.setBit(idx)
	}
}

func (bf *BloomFilter) Check(data string) bool {
	for _, idx := range bf.getHashes(data) {
		if !bf.getBit(idx) {
			return false
		}
	}
	return true
}
