package ggpool

import (
	"reflect"
	"sync"
)

type collection struct {
	sync.RWMutex
	allItems  map[uintptr]*item
	idleItems map[uintptr]*item
}

func newCollection() *collection {
	return &collection{
		allItems:  make(map[uintptr]*item),
		idleItems: make(map[uintptr]*item),
	}
}

func getObjectKey(object *interface{}) uintptr {
	return reflect.ValueOf(object).Pointer()
}

func (c *collection) len() int {
	c.RLock()
	defer c.RUnlock()

	return len(c.allItems)
}

func (c *collection) lenIdle() int {
	c.RLock()
	defer c.RUnlock()

	return len(c.idleItems)
}

func (c *collection) acquire() *item {
	c.Lock()
	defer c.Unlock()

	for key, item := range c.idleItems {
		delete(c.idleItems, key)
		return item
	}
	return nil
}

func (c *collection) acquireAll() []*item {
	c.Lock()
	defer c.Unlock()

	var res []*item

	for key, item := range c.idleItems {
		delete(c.idleItems, key)
		res = append(res, item)
	}
	return res
}

func (c *collection) get(key uintptr) *item {
	c.RLock()
	defer c.RUnlock()

	return c.allItems[key]
}

func (c *collection) getAll() []*item {
	c.Lock()
	defer c.Unlock()

	var res []*item

	for _, item := range c.allItems {
		res = append(res, item)
	}
	return res
}

func (c *collection) put(key uintptr, value *item) {
	c.Lock()
	defer c.Unlock()

	c.allItems[key] = value
}

func (c *collection) release(key uintptr) {
	c.Lock()
	defer c.Unlock()

	item := c.allItems[key]

	c.idleItems[key] = item
}

func (c *collection) remove(key uintptr) {
	c.Lock()
	defer c.Unlock()

	delete(c.allItems, key)
	delete(c.idleItems, key)
}
