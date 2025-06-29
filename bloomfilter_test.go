package main

import (
	"fmt"
	"testing"
)

func TestBloomFilter_Creation(t *testing.T) {
	tests := []struct {
		name      string
		size      uint64
		hashCount uint64
	}{
		{"Small filter", 100, 1},
		{"Medium filter", 1000, 3},
		{"Large filter", 10000, 5},
		{"Single hash", 500, 1},
		{"Many hashes", 1000, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := NewBloomFilter(tt.size, tt.hashCount)
			if bf.size != tt.size {
				t.Errorf("Expected size %d, got %d", tt.size, bf.size)
			}
			if bf.hashCount != tt.hashCount {
				t.Errorf("Expected hash count %d, got %d", tt.hashCount, bf.hashCount)
			}
		})
	}
}

func TestBloomFilter_AddAndCheck(t *testing.T) {
	bf := NewBloomFilter(10000, 3)

	testItems := []string{"apple", "banana", "cherry", "date", "elderberry"}

	for _, item := range testItems {
		bf.Add(item)
		if !bf.Check(item) {
			t.Errorf("Item %s should be in bloom filter after adding", item)
		}
	}
}

func TestBloomFilter_LargeDataSet(t *testing.T) {
	bf := NewBloomFilter(10000, 3)
	itemCount := 998

	for i := 1; i <= itemCount; i++ {
		key := fmt.Sprintf("key%d", i)
		bf.Add(key)
	}

	for i := 1; i <= itemCount; i++ {
		key := fmt.Sprintf("key%d", i)
		if !bf.Check(key) {
			t.Errorf("Expected %s to be in bloom filter", key)
		}
	}

	nonExistingKeys := []string{"key999", "key1000", "unknown", "notadded"}
	falsePositives := 0

	for _, key := range nonExistingKeys {
		if bf.Check(key) {
			falsePositives++
			t.Logf("False positive detected for key: %s", key)
		}
	}

	t.Logf("False positive rate: %d/%d = %.2f%%",
		falsePositives, len(nonExistingKeys),
		float64(falsePositives)/float64(len(nonExistingKeys))*100)
}

func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	bf := NewBloomFilter(1000, 3)

	for i := range 100 {
		bf.Add(fmt.Sprintf("item%d", i))
	}

	falsePositives := 0
	testCount := 1000

	for i := 100; i < 100+testCount; i++ {
		if bf.Check(fmt.Sprintf("item%d", i)) {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / float64(testCount) * 100
	t.Logf("False positive rate: %.2f%% (%d/%d)", falsePositiveRate, falsePositives, testCount)

	if falsePositiveRate > 10.0 {
		t.Errorf("False positive rate too high: %.2f%%", falsePositiveRate)
	}
}

func TestBloomFilter_EdgeCases(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		bf := NewBloomFilter(100, 3)
		bf.Add("")
		if !bf.Check("") {
			t.Error("Empty string should be found after adding")
		}
	})

	t.Run("Very long string", func(t *testing.T) {
		bf := NewBloomFilter(1000, 3)
		longString := string(make([]byte, 10000))
		for i := range longString {
			longString = longString[:i] + "a" + longString[i+1:]
		}
		bf.Add(longString)
		if !bf.Check(longString) {
			t.Error("Long string should be found after adding")
		}
	})

	t.Run("Unicode strings", func(t *testing.T) {
		bf := NewBloomFilter(1000, 3)
		unicodeStrings := []string{"🚀", "café", "naïve", "北京", "москва"}

		for _, str := range unicodeStrings {
			bf.Add(str)
			if !bf.Check(str) {
				t.Errorf("Unicode string %s should be found after adding", str)
			}
		}
	})
}

func BenchmarkBloomFilter_Add(b *testing.B) {
	bf := NewBloomFilter(10000, 3)
	b.ResetTimer()

	i := 0
	for b.Loop() {
		bf.Add(fmt.Sprintf("item%d", i))
		i++
	}
}

func BenchmarkBloomFilter_Check(b *testing.B) {
	bf := NewBloomFilter(10000, 3)

	// Pre-populate with some items
	for i := range 1000 {
		bf.Add(fmt.Sprintf("item%d", i))
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		bf.Check(fmt.Sprintf("item%d", i%1000))
		i++
	}
}

func BenchmarkBloomFilter_AddAndCheck(b *testing.B) {
	bf := NewBloomFilter(10000, 3)
	b.ResetTimer()

	i := 0
	for b.Loop() {
		key := fmt.Sprintf("item%d", i)
		bf.Add(key)
		bf.Check(key)
		i++
	}
}
