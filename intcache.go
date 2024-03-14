package intcache

import (
	"sync/atomic"
)

const soltNum = 8

type IntCache struct {
	b          uint8
	buckets    [][soltNum]uint64
	lruBuckets []uint32
}

func New(b uint8) *IntCache {
	cap := 1 << b
	return &IntCache{
		b:          b,
		buckets:    make([][soltNum]uint64, cap),
		lruBuckets: make([]uint32, cap),
	}
}

func (c *IntCache) Get(key uint32) (value uint32, exists bool) {
	bucketi := key & (1<<c.b - 1)
	for i := 0; i < soltNum; i++ {
		e := atomic.LoadUint64(&c.buckets[bucketi][i])
		if e == 0 {
			break
		}

		if uint32(e>>32) == key {
			value = uint32(e)
			exists = true
			c.updLru(bucketi, i)
			break
		}
	}

	return
}

func (c *IntCache) Set(key uint32, value uint32) {
	if key == 0 && value == 0 {
		panic("key and value can't be 0")
	}
	bucketi := key & (1<<c.b - 1)
	for i := 0; i < soltNum; i++ {
		e := atomic.LoadUint64(&c.buckets[bucketi][i])
		if e == 0 {
			atomic.StoreUint64(&c.buckets[bucketi][i], uint64(key)<<32|uint64(value))
			c.updLru(bucketi, i)
			return
		}

		if uint32(e>>32) == key {
			e = uint64(key)<<32 | uint64(value)
			atomic.StoreUint64(&c.buckets[bucketi][i], e)
			c.updLru(bucketi, i)
			return
		}
	}

	// find the min lru
	lrus := atomic.LoadUint32(&c.lruBuckets[bucketi])
	var (
		mini int
	)
	for i := 0; i < soltNum; i++ {
		lru := lrus | 0b1111<<uint32(i)
		if lru == 0 {
			mini = i
			break
		}
	}

	atomic.StoreUint64(&c.buckets[bucketi][mini], uint64(key)<<32|uint64(value))
	c.updLru(bucketi, mini)
}

func (c *IntCache) updLru(bucketi uint32, ei int) {
	lrus := atomic.LoadUint32(&c.lruBuckets[bucketi])
	ebiti := ei << 2
	oldLru := lrus & (0b1111 << ebiti) >> ebiti
	for i := 0; i < soltNum; i++ {
		biti := i << 2
		if ei == i {
			lrus = lrus&^(0b1111<<biti) | 7<<biti
		} else {
			lru := lrus >> uint32(biti) & 0xf
			if lru > oldLru {
				lru--
				lrus = lrus&^(0b1111<<biti) | lru<<biti
			}
		}
	}

	atomic.StoreUint32(&c.lruBuckets[bucketi], lrus)
}
