package intcache

import (
	"encoding/binary"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/stretchr/testify/assert"
)

func TestSimpleSet(t *testing.T) {
	var b uint8 = 1
	m := New(b)
	m.Set(1, 1)
	value, _ := m.Get(1)
	assert.Equal(t, uint32(1), value)
	m.Set(3, 3)
	value, _ = m.Get(3)
	assert.Equal(t, uint32(3), value)
	m.Set(5, 5)
	value, _ = m.Get(3)
	assert.Equal(t, uint32(3), value)
	m.Set(3, 30)
	value, _ = m.Get(3)
	assert.Equal(t, uint32(30), value)
}

func TestSet(t *testing.T) {
	var b uint8 = 21
	m := New(b)
	m.Set(1, 1)
	m.Set(1+1<<b, 11)
	m.Set(2, 2)
	m.Set(2+1<<b, 21)
	m.Set(3, 3)
	value1, _ := m.Get(1)
	value2, _ := m.Get(2)
	value3, _ := m.Get(3)
	value1_1, _ := m.Get(1 + 1<<b)
	value2_1, _ := m.Get(2 + 1<<b)
	assert.Equal(t, uint32(1), value1)
	assert.Equal(t, uint32(2), value2)
	assert.Equal(t, uint32(3), value3)
	assert.Equal(t, uint32(11), value1_1)
	assert.Equal(t, uint32(21), value2_1)
}

func TestParallelSet(t *testing.T) {
	rand.Seed(time.Now().Unix())
	var b uint8 = 21
	m := New(b)
	wg := &sync.WaitGroup{}
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			key := rand.Uint32()
			m.Set(key, key)
			if key%2 > 0 {
				time.Sleep(time.Millisecond * 10)
			}

			value, exists := m.Get(key)
			if exists {
				assert.Equal(t, key, value)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log("ParallelSet 100000g done")
	m = New(b)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				key := rand.Uint32()
				m.Set(key, key)
				if key%2 > 0 {
					time.Sleep(time.Millisecond * 10)
				}

				value, exists := m.Get(key)
				if exists {
					assert.Equal(t, key, value)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkIntcache(b *testing.B) {
	rand.Seed(1)
	var B uint8 = 21
	m := New(B)
	for i := 0; i < b.N; i++ {
		intcacheBenchmarkFn(m)
	}
}

func BenchmarkFastcache(b *testing.B) {
	rand.Seed(1)
	m := fastcache.New(1 << 21 * (64 + 8))
	for i := 0; i < b.N; i++ {
		fastcacheBenchmarkFn(m)
	}
}

func intcacheBenchmarkFn(m *IntCache) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 300; j++ {
				key := rand.Uint32()
				m.Set(key, key)
				m.Get(key)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func fastcacheBenchmarkFn(m *fastcache.Cache) {
	wg := &sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 300; j++ {
				key := rand.Uint32()
				b := make([]byte, 4) // uint64的大小为8字节
				binary.LittleEndian.PutUint32(b, key)
				m.Set(b, b)
				r := make([]byte, 4)
				m.Get(r, b)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestUpdLru(t *testing.T) {
	var b uint8 = 1
	m := New(b)
	m.lruBuckets[0] = 0x765
	m.updLru(0, 1)
	assert.Equal(t, uint32(0x675), m.lruBuckets[0])
	m.lruBuckets[0] = 0x36217504
	m.updLru(0, 1)
	assert.Equal(t, uint32(0x25106473), m.lruBuckets[0])
}
