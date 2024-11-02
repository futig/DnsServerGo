package utils

import (
	"fmt"
	"sync"
	"time"
)

type RequestAnswer struct {
	Expire    time.Time
	R      bool
	Answer []byte
}

type Cache struct {
	Size  int
	queue Queue[string]
	names map[string]*RequestAnswer
}

func NewCache(size int) *Cache {
	return &Cache{
		Size:  size,
		queue: make(Queue[string], 0),
		names: make(map[string]*RequestAnswer, 0),
	}
}

func (c *Cache) Get(name string, rType uint16) ([]byte, bool) {
	key := createIdentityKey(name, rType)
	var m sync.Mutex
	m.Lock()
	defer m.Unlock()
	if answer, ok := c.names[key]; ok {
		if time.Since(answer.Expire).Seconds() < 0 {
			answer.R = true
			return answer.Answer, true
		}
	}
	return nil, false
}

func createIdentityKey(name string, rType uint16) string {
	return fmt.Sprintf("%s:%d", name, rType)
}

func (c *Cache) Put(name string, rType uint16, expire time.Time, answer []byte) {
	var m sync.Mutex
	m.Lock()
	defer m.Unlock()
	for len(c.queue) == c.Size {
		queue, identityKey, err := c.queue.Dequeue()
		if err != nil {
			return
		}
		c.queue = queue
		requestAnswer := c.names[identityKey]
		elapsed := time.Since(requestAnswer.Expire).Seconds()
		if requestAnswer.R && elapsed < 0{
			requestAnswer.R = false
			c.queue = c.queue.Enqueue(identityKey)
		} else {
			delete(c.names, identityKey)
		}
	}
	identityKey := createIdentityKey(name, rType)
	
	requestAnswer := RequestAnswer{
		Expire: expire,
		Answer: answer,
		R:      true,
	}
	c.names[identityKey] = &requestAnswer
	c.queue = c.queue.Enqueue(identityKey)
}