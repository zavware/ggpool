//Package ggpool provides functionality of simple generic objects pool
package ggpool

import (
	"context"
	"errors"
	"reflect"
	"time"
)

//Pool is a pool of generic objects
type Pool struct {
	config              Config
	items               chan *Item
	poolLengthCounter   chan int
	hasPendingNewItem   chan bool
	createItemLastError chan error
	ctx                 context.Context
	cancel              context.CancelFunc
}

//NewPool returns a new Pool instanse
func NewPool(ctx context.Context, config Config) (*Pool, error) {
	var p *Pool

	if ok, err := p.validateConfig(config); ok != true {
		return p, err
	}

	ctx, cancel := context.WithCancel(ctx)

	p = &Pool{
		config:            config,
		items:             make(chan *Item, config.Capacity),
		poolLengthCounter: make(chan int, 1),
		hasPendingNewItem: make(chan bool),
		ctx:               ctx,
		cancel:            cancel,
	}

	p.poolLengthCounter <- 0

	go func() { p.putPending() }()
	go func() { p.checkMinCapacity() }()
	go func() { p.cleanUp() }()

	return p, nil
}

//Get returns Item or error of Item getting/creation
func (p *Pool) Get() (*Item, error) {
	var item *Item
	var err error

	if p.ctx.Err() == context.Canceled {
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

//Len returns pool current length
func (p *Pool) Len() int {
	c := <-p.poolLengthCounter
	p.poolLengthCounter <- c
	return c
}

//Close clears and closes pool
func (p *Pool) Close() {
	p.cancel()
	p.destroyItems(true)
}

func (p *Pool) putPending() {
	for {
		select {
		case <-p.hasPendingNewItem:
			c := <-p.poolLengthCounter
			if len(p.items) == 0 && c < p.config.Capacity {
				if item, err := p.createItem(); err == nil {
					p.createItemLastError = nil
					p.items <- item
					c++
				} else {
					p.createItemLastError = nil
					p.createItemLastError = make(chan error, 1)
					p.createItemLastError <- err
				}
			}

			p.poolLengthCounter <- c

		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) checkMinCapacity() {
	ticker := time.NewTicker(time.Duration(10 * time.Second))
	defer ticker.Stop()

	p.keepMinCapacity()

	for {
		select {
		case <-ticker.C:
			p.keepMinCapacity()
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) cleanUp() {
	ticker := time.NewTicker(p.config.ItemLifetimeCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.destroyItems(false)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) createItem() (*Item, error) {
	var item *Item
	var err error

	object, err := p.config.Factory.Create(p.ctx)

	if err != nil {
		return item, err
	}

	if reflect.ValueOf(object).Kind() != reflect.Ptr {
		return item, errors.New("ggpool.Config.Factory must return object pointer")
	}

	if _, ok := object.(Object); !ok {
		return item, errors.New("ggpool.Config.Factory must create object which implement ggpool.Object interface")
	}

	item = &Item{
		object:       &object,
		pool:         p,
		releasedTime: time.Now().UTC(),
	}

	return item, err
}

func (p *Pool) destroyItems(force bool) {
	var itemsBuffer []*Item

	for len(p.items) > 0 {
		item := <-p.items
		if force || !item.isActive() {
			p.poolLengthCounter <- <-p.poolLengthCounter - 1
			item.destroy()
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
