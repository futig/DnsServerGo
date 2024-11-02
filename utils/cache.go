package utils

import (
	"fmt"
	"sync"
)

type RequestAnswer struct {
	Name   string
	RType  uint16
	R      bool
	Answer []byte
}

type identityKey struct {
	Name  string
	RType uint16
}

type Cache struct {
	Size  int
	queue Queue[identityKey]
	names map[identityKey]*RequestAnswer
	rw *sync.RWMutex
}

func NewCache(size int) *Cache{
	return &Cache{
		Size: size,
		queue: make(Queue[identityKey],0),
		names: make(map[identityKey]*RequestAnswer, 0),
	}
}

func (c *Cache) Get(name string, rType uint16) ([]byte, bool) {
	key := createIdentityKey(name, rType)
	c.rw.Lock()
	defer c.rw.Unlock()
	if answer, ok := c.names[key]; ok {
		answer.R = true
		return answer.Answer, true
	}
	return nil, false
}

func createIdentityKey(name string, rType uint16) identityKey {
	return identityKey{
		Name:  name,
		RType: rType,
	}
}

func (c *Cache) Put(name string, rType uint16, answer []byte) {
	c.rw.Lock()
	defer c.rw.Unlock()
	for len(c.queue) == c.Size {
		queue, identityKey, err := c.queue.Dequeue()
		if err != nil {
			return
		}
		c.queue = queue
		requestAnswer := c.names[identityKey]
		if requestAnswer.R {
			requestAnswer.R = false
			c.queue = c.queue.Enqueue(identityKey)
		} else {
			delete(c.names, identityKey)
		}
	}
	identityKey := createIdentityKey(name, rType)
	requestAnswer := RequestAnswer{
		Name:   name,
		RType:  rType,
		Answer: answer,
		R:      true,
	}
	c.names[identityKey] = &requestAnswer
	c.queue = c.queue.Enqueue(identityKey)
}

func (h Cache) String() string {
	return fmt.Sprint("%v", len(h.queue))
}