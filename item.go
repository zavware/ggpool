package ggpool

import "time"

type item struct {
	object       Object
	pool         *pool
	releasedTime time.Time
	clock        clock
}

func (i *item) GetObject() Object {
	return i.object
}

func (i *item) Release() {
	i.releasedTime = i.clock.Now().UTC()
	i.pool.items <- i
}

func (i *item) Destroy() (bool, error) {
	return (i.GetObject()).Destroy()
}

func (i *item) isActive() bool {
	expireTime := i.clock.Now().Local().Add(i.pool.config.ItemLifetime)
	return i.releasedTime.Before(expireTime) && (i.GetObject()).IsActive()
}
