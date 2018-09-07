package ggpool

import (
	"errors"
	"time"
)

type TimeoutError struct {
	err string
}

func (e *TimeoutError) Error() string {
	return e.err
}

func (e *TimeoutError) IsTemporary() bool {
	return true
}

type Temporary interface {
	IsTemporary() bool
}

type Pool struct {
	config                 Config
	items                  chan *Item
	poolLengthCounter      chan int
	hasPendingNewItem      chan bool
	createItemLastError    chan error
	isClosed               bool
	cleanUpTicker          *time.Ticker
	checkMinCapacityTicker *time.Ticker
}

func New(config Config) (*Pool, error) {
	var p *Pool

	if ok, err := p.validateConfig(config); ok != true {
		return p, err
	}

	p = &Pool{
		config:                 config,
		items:                  make(chan *Item, config.Capacity),
		poolLengthCounter:      make(chan int, 1),
		hasPendingNewItem:      make(chan bool),
		createItemLastError:    make(chan error, 1),
		isClosed:               false,
		cleanUpTicker:          time.NewTicker(config.ItemLifetimeCheckPeriod),
		checkMinCapacityTicker: time.NewTicker(time.Duration(10 * time.Second)),
	}

	p.poolLengthCounter <- 0

	go func() { p.putPending() }()
	go func() { p.checkMinCapacity() }()
	go func() { p.cleanUp() }()

	return p, nil
}

func (p *Pool) Get() (*Item, error) {
	var item *Item
	var err error

	if p.isClosed {
		err = errors.New("pool is closed")
		return nil, err
	}

	timeout := time.After(p.config.Timeout)
	p.hasPendingNewItem <- true

	select {
	case <-timeout:
		err = &TimeoutError{"timeout exceeded - cannot get pool item"}
	case err = <-p.createItemLastError:
		break
	case item = <-p.items:
		break
	}

	return item, err
}

func (p *Pool) Len() int {
	c := <-p.poolLengthCounter
	p.poolLengthCounter <- c
	return c
}

func (p *Pool) Close() {
	p.isClosed = true
	p.cleanUpTicker.Stop()
	p.checkMinCapacityTicker.Stop()
	close(p.hasPendingNewItem)

	p.destroyItems(true)
}

func (p *Pool) putPending() {
	for <-p.hasPendingNewItem {
		c := <-p.poolLengthCounter

		if len(p.items) == 0 && c < p.config.Capacity {
			if item, err := p.createItem(); err == nil {
				<-p.createItemLastError
				p.items <- item
				c++
			} else {
				<-p.createItemLastError
				p.createItemLastError <- err
			}
		}

		p.poolLengthCounter <- c
	}
}

func (p *Pool) checkMinCapacity() {
	for ; true; <-p.checkMinCapacityTicker.C {
		p.keepMinCapacity()
	}
}

func (p *Pool) cleanUp() {
	for range p.cleanUpTicker.C {
		p.destroyItems(false)
	}
}

func (p *Pool) createItem() (*Item, error) {
	if object, err := p.config.Factory.Create(); err == nil {
		item := &Item{
			object:       &object,
			pool:         p,
			releasedTime: time.Now().UTC(),
		}
		return item, nil
	} else {
		return nil, err
	}
}

func (p *Pool) destroyItems(force bool) {
	var itemsBuffer []*Item

	for len(p.items) > 0 {
		item := <-p.items
		if force || !item.isActive() {
			p.poolLengthCounter <- <-p.poolLengthCounter - 1
			item.Destroy()
		} else {
			itemsBuffer = append(itemsBuffer, item)
		}
	}

	for _, item := range itemsBuffer {
		p.items <- item
	}
}

func (p *Pool) keepMinCapacity() {
	c := <-p.poolLengthCounter
	p.poolLengthCounter <- c

	delta := p.config.MinCapacity - c

	for i := 0; i < delta; i++ {
		p.hasPendingNewItem <- true
	}
}

func (p *Pool) validateConfig(config Config) (bool, error) {
	res := true
	var err error

	if config.Capacity < 1 {
		res = false
		err = errors.New("pool capacity value must be more then 0")
	}

	if config.MinCapacity < 0 {
		res = false
		err = errors.New("min pool capacity value must not be negative")
	}

	if config.Capacity < config.MinCapacity {
		res = false
		err = errors.New("pool capacity value cannot be less then init capacity value")
	}

	return res, err
}
