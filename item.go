package ggpool

import "time"

type item struct {
	object       *interface{}
	pool         *pool
	releasedTime time.Time
}

func (i *item) GetObject() *interface{} {
	return i.object
}

func (i *item) Release() {
	i.releasedTime = time.Now().UTC()
	i.pool.items <- i
}

func (i *item) Destroy() {
	(*i.object).(Object).Destroy()
}

func (i *item) isActive() bool {
	expireTime := time.Now().UTC().Add(i.pool.config.ItemLifetime)
	return i.releasedTime.Before(expireTime) && (*i.object).(Object).IsActive()
}
