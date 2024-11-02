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

type Cache struct {
	Size  int
	queue Queue[string]
	names map[string]*RequestAnswer
}

func NewCache(size int) *Cache{
	return &Cache{
		Size: size,
		queue: make(Queue[string],0),
		names: make(map[string]*RequestAnswer, 0),
	}
}

func (c *Cache) Get(name string, rType uint16) ([]byte, bool) {
	key := createIdentityKey(name, rType)
	var rw sync.RWMutex
	rw.Lock()
	defer rw.Unlock()
	if answer, ok := c.names[key]; ok {
		answer.R = true
		return answer.Answer, true
	}
	return nil, false
}

func createIdentityKey(name string, rType uint16) string {
	return fmt.Sprintf("%s:%d", name, rType)
}

func (c *Cache) Put(name string, rType uint16, answer []byte) {
	var rw sync.RWMutex
	rw.Lock()
	defer rw.Unlock()
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
	return fmt.Sprintf("%v", len(h.queue))
}