package ggpool

import "time"

//Item is wrapper of pool Object
type Item struct {
	object       *interface{}
	pool         *Pool
	releasedTime time.Time
}

//GetObject returns Object
func (i *Item) GetObject() *interface{} {
	return i.object
}

//Release puts Item back to Pool
func (i *Item) Release() {
	i.releasedTime = time.Now().UTC()
	i.pool.items <- i
}

func (i *Item) destroy() {
	(*i.object).(Object).Destroy()
}

func (i *Item) isActive() bool {
	expireTime := time.Now().UTC().Add(i.pool.config.ItemLifetime)
	return i.releasedTime.Before(expireTime) && (*i.object).(Object).IsActive()
}
