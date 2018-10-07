package ggpool

import "time"

type item struct {
	object       *interface{}
	pool         *pool
	releasedTime time.Time
	clock        clock
}

func (i *item) GetObject() *interface{} {
	return i.object
}

func (i *item) Release() {
	i.releasedTime = i.clock.Now().UTC()
	i.pool.items <- i
}

func (i *item) Destroy() {
	(*i.GetObject()).(Object).Destroy()
}

func (i *item) isActive() bool {
	expireTime := i.clock.Now().UTC().Add(i.pool.config.ItemLifetime)
	return i.releasedTime.Before(expireTime) && (*i.GetObject()).(Object).IsActive()
}
