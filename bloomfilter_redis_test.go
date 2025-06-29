package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

const (
	testRedisAddr = "localhost:6379"
	testTimeout   = 5 * time.Second
)

func setupRedisTest(t *testing.T) {
	rbf, err := NewRedisBloomFilter(testRedisAddr, "test-connectivity", 0.01, 1000)
	if err != nil {
		t.Skipf("Redis not available at %s: %v", testRedisAddr, err)
		return
	}

	ctx := context.Background()
	rbf.client.Del(ctx, "test-connectivity")
	rbf.client.Close()
}

func TestRedisBloomFilter_Creation(t *testing.T) {
	setupRedisTest(t)

	tests := []struct {
		name      string
		key       string
		errorRate float64
		capacity  int64
	}{
		{"Small filter", "test-small", 0.01, 1000},
		{"Medium filter", "test-medium", 0.001, 10000},
		{"Large filter", "test-large", 0.0001, 100000},
		{"High error rate", "test-high-error", 0.1, 5000},
		{"Low error rate", "test-low-error", 0.00001, 50000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rbf, err := NewRedisBloomFilter(testRedisAddr, tt.key, tt.errorRate, tt.capacity)
			if err != nil {
				t.Fatalf("Failed to create Redis bloom filter: %v", err)
			}
			defer func() {
				ctx := context.Background()
				rbf.client.Del(ctx, tt.key)
				rbf.client.Close()
			}()

			if rbf.client == nil {
				t.Error("Redis client should not be nil")
			}
			if rbf.key != tt.key {
				t.Errorf("Expected key %s, got %s", tt.key, rbf.key)
			}
		})
	}
}

func TestRedisBloomFilter_AddAndCheck(t *testing.T) {
	setupRedisTest(t)

	rbf, err := NewRedisBloomFilter(testRedisAddr, "test-add-check", 0.01, 10000)
	if err != nil {
		t.Fatalf("Failed to create Redis bloom filter: %v", err)
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "test-add-check")
		rbf.client.Close()
	}()

	testItems := []string{"apple", "banana", "cherry", "date", "elderberry"}

	for _, item := range testItems {
		added, err := rbf.Add(item)
		if err != nil {
			t.Errorf("Failed to add item %s: %v", item, err)
			continue
		}

		if !added {
			t.Logf("Item %s was already in filter (expected for first add)", item)
		}

		exists, err := rbf.Check(item)
		if err != nil {
			t.Errorf("Failed to check item %s: %v", item, err)
			continue
		}

		if !exists {
			t.Errorf("Item %s should exist in bloom filter after adding", item)
		}
	}
}

func TestRedisBloomFilter_LargeDataSet(t *testing.T) {
	setupRedisTest(t)

	rbf, err := NewRedisBloomFilter(testRedisAddr, "test-large-dataset", 0.01, 50000)
	if err != nil {
		t.Fatalf("Failed to create Redis bloom filter: %v", err)
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "test-large-dataset")
		rbf.client.Close()
	}()

	itemCount := 998

	// Add items
	t.Logf("Adding %d items to Redis bloom filter...", itemCount)
	for i := 1; i <= itemCount; i++ {
		key := fmt.Sprintf("key%d", i)
		_, err := rbf.Add(key)
		if err != nil {
			t.Errorf("Failed to add key %s: %v", key, err)
		}
	}

	// Check all added items exist
	t.Logf("Checking %d items in Redis bloom filter...", itemCount)
	for i := 1; i <= itemCount; i++ {
		key := fmt.Sprintf("key%d", i)
		exists, err := rbf.Check(key)
		if err != nil {
			t.Errorf("Failed to check key %s: %v", key, err)
			continue
		}
		if !exists {
			t.Errorf("Expected %s to be in bloom filter", key)
		}
	}

	// Check non-existing items
	nonExistingKeys := []string{"key999", "key1000", "unknown", "notadded"}
	falsePositives := 0

	for _, key := range nonExistingKeys {
		exists, err := rbf.Check(key)
		if err != nil {
			t.Errorf("Failed to check non-existing key %s: %v", key, err)
			continue
		}
		if exists {
			falsePositives++
			t.Logf("False positive detected for key: %s", key)
		}
	}

	t.Logf("False positive rate: %d/%d = %.2f%%",
		falsePositives, len(nonExistingKeys),
		float64(falsePositives)/float64(len(nonExistingKeys))*100)
}

func TestRedisBloomFilter_FalsePositiveRate(t *testing.T) {
	setupRedisTest(t)

	rbf, err := NewRedisBloomFilter(testRedisAddr, "test-false-positive", 0.01, 10000)
	if err != nil {
		t.Fatalf("Failed to create Redis bloom filter: %v", err)
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "test-false-positive")
		rbf.client.Close()
	}()

	// Add 100 items
	for i := 0; i < 100; i++ {
		_, err := rbf.Add(fmt.Sprintf("item%d", i))
		if err != nil {
			t.Errorf("Failed to add item%d: %v", i, err)
		}
	}

	// Test 1000 non-existing items
	falsePositives := 0
	testCount := 1000

	for i := 100; i < 100+testCount; i++ {
		exists, err := rbf.Check(fmt.Sprintf("item%d", i))
		if err != nil {
			t.Errorf("Failed to check item%d: %v", i, err)
			continue
		}
		if exists {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / float64(testCount) * 100
	t.Logf("False positive rate: %.2f%% (%d/%d)", falsePositiveRate, falsePositives, testCount)

	if falsePositiveRate > 10.0 {
		t.Errorf("False positive rate too high: %.2f%%", falsePositiveRate)
	}
}

func TestRedisBloomFilter_EdgeCases(t *testing.T) {
	setupRedisTest(t)

	t.Run("Empty string", func(t *testing.T) {
		rbf, err := NewRedisBloomFilter(testRedisAddr, "test-empty-string", 0.01, 1000)
		if err != nil {
			t.Fatalf("Failed to create Redis bloom filter: %v", err)
		}
		defer func() {
			ctx := context.Background()
			rbf.client.Del(ctx, "test-empty-string")
			rbf.client.Close()
		}()

		_, err = rbf.Add("")
		if err != nil {
			t.Errorf("Failed to add empty string: %v", err)
		}

		exists, err := rbf.Check("")
		if err != nil {
			t.Errorf("Failed to check empty string: %v", err)
		}
		if !exists {
			t.Error("Empty string should be found after adding")
		}
	})

	t.Run("Very long string", func(t *testing.T) {
		rbf, err := NewRedisBloomFilter(testRedisAddr, "test-long-string", 0.01, 1000)
		if err != nil {
			t.Fatalf("Failed to create Redis bloom filter: %v", err)
		}
		defer func() {
			ctx := context.Background()
			rbf.client.Del(ctx, "test-long-string")
			rbf.client.Close()
		}()

		longString := string(make([]byte, 1024))
		for i := range longString {
			longString = longString[:i] + "a" + longString[i+1:]
		}

		_, err = rbf.Add(longString)
		if err != nil {
			t.Errorf("Failed to add long string: %v", err)
		}

		exists, err := rbf.Check(longString)
		if err != nil {
			t.Errorf("Failed to check long string: %v", err)
		}
		if !exists {
			t.Error("Long string should be found after adding")
		}
	})

	t.Run("Unicode strings", func(t *testing.T) {
		rbf, err := NewRedisBloomFilter(testRedisAddr, "test-unicode", 0.01, 1000)
		if err != nil {
			t.Fatalf("Failed to create Redis bloom filter: %v", err)
		}
		defer func() {
			ctx := context.Background()
			rbf.client.Del(ctx, "test-unicode")
			rbf.client.Close()
		}()

		unicodeStrings := []string{"🚀", "café", "naïve", "北京", "москва", "🌸🎯🏆"}

		for _, str := range unicodeStrings {
			_, err := rbf.Add(str)
			if err != nil {
				t.Errorf("Failed to add unicode string %s: %v", str, err)
				continue
			}

			exists, err := rbf.Check(str)
			if err != nil {
				t.Errorf("Failed to check unicode string %s: %v", str, err)
				continue
			}
			if !exists {
				t.Errorf("Unicode string %s should be found after adding", str)
			}
		}
	})

	t.Run("Special characters", func(t *testing.T) {
		rbf, err := NewRedisBloomFilter(testRedisAddr, "test-special-chars", 0.01, 1000)
		if err != nil {
			t.Fatalf("Failed to create Redis bloom filter: %v", err)
		}
		defer func() {
			ctx := context.Background()
			rbf.client.Del(ctx, "test-special-chars")
			rbf.client.Close()
		}()

		specialStrings := []string{
			"key with spaces",
			"key\nwith\nnewlines",
			"key\twith\ttabs",
			"key\"with\"quotes",
			"key'with'apostrophes",
			"key\\with\\backslashes",
			"key/with/slashes",
		}

		for _, str := range specialStrings {
			_, err := rbf.Add(str)
			if err != nil {
				t.Errorf("Failed to add special string %q: %v", str, err)
				continue
			}

			exists, err := rbf.Check(str)
			if err != nil {
				t.Errorf("Failed to check special string %q: %v", str, err)
				continue
			}
			if !exists {
				t.Errorf("Special string %q should be found after adding", str)
			}
		}
	})
}

func TestRedisBloomFilter_Concurrent(t *testing.T) {
	setupRedisTest(t)

	rbf, err := NewRedisBloomFilter(testRedisAddr, "test-concurrent", 0.01, 50000)
	if err != nil {
		t.Fatalf("Failed to create Redis bloom filter: %v", err)
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "test-concurrent")
		rbf.client.Close()
	}()

	done := make(chan bool, 2)

	go func() {
		for i := range 1000 {
			_, err := rbf.Add(fmt.Sprintf("concurrent-item-%d", i))
			if err != nil {
				t.Errorf("Failed to add concurrent item %d: %v", i, err)
			}
		}
		done <- true
	}()

	go func() {
		for i := range 1000 {
			_, err := rbf.Check(fmt.Sprintf("concurrent-item-%d", i))
			if err != nil {
				t.Errorf("Failed to check concurrent item %d: %v", i, err)
			}
		}
		done <- true
	}()

	<-done
	<-done
}

func BenchmarkRedisBloomFilter_Add(b *testing.B) {
	rbf, err := NewRedisBloomFilter(testRedisAddr, "bench-add", 0.01, 100000)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "bench-add")
		rbf.client.Close()
	}()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		_, err := rbf.Add(fmt.Sprintf("bench-item%d", i))
		if err != nil {
			b.Fatalf("Failed to add item: %v", err)
		}
		i++
	}
}

func BenchmarkRedisBloomFilter_Check(b *testing.B) {
	rbf, err := NewRedisBloomFilter(testRedisAddr, "bench-check", 0.01, 100000)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "bench-check")
		rbf.client.Close()
	}()

	for i := range 1000 {
		rbf.Add(fmt.Sprintf("bench-item%d", i))
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		_, err := rbf.Check(fmt.Sprintf("bench-item%d", i%1000))
		if err != nil {
			b.Fatalf("Failed to check item: %v", err)
		}
		i++
	}
}

func BenchmarkRedisBloomFilter_AddAndCheck(b *testing.B) {
	rbf, err := NewRedisBloomFilter(testRedisAddr, "bench-add-check", 0.01, 100000)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}
	defer func() {
		ctx := context.Background()
		rbf.client.Del(ctx, "bench-add-check")
		rbf.client.Close()
	}()

	b.ResetTimer()
	i := 0
	for b.Loop() {
		key := fmt.Sprintf("bench-item%d", i)
		_, err := rbf.Add(key)
		if err != nil {
			b.Fatalf("Failed to add item: %v", err)
		}

		_, err = rbf.Check(key)
		if err != nil {
			b.Fatalf("Failed to check item: %v", err)
		}
		i++
	}
}
