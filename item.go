package ggpool

import "time"

type item struct {
	object       *interface{}
	lifetime     time.Duration
	releasedTime time.Time
}

func newItem(object *interface{}, lifetime time.Duration) *item {
	return &item{
		object:       object,
		lifetime:     lifetime,
		releasedTime: time.Now().UTC(),
	}
}

func (i *item) release() {
	i.releasedTime = time.Now().UTC()
}

func (i *item) destroy() {
	(*i.object).(Object).Destroy()
}

func (i *item) isActive() bool {
	expireTime := time.Now().UTC().Add(i.lifetime)
	return i.releasedTime.Before(expireTime)
}
