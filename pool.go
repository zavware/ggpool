package ggpool

import (
	"context"
	"errors"
	"reflect"
	"time"
)

//TimeoutError is temporary error
//Can be catched by Temporary interface
type TimeoutError struct {
	err string
}

func (e *TimeoutError) Error() string {
	return e.err
}

//IsTemporary says that TimeoutError is temporary error
func (e *TimeoutError) IsTemporary() bool {
	return true
}

//Temporary is interface for temporary pool errors
type Temporary interface {
	IsTemporary() bool
}

type pool struct {
	config              Config
	items               chan *item
	poolLengthCounter   chan int
	hasPendingNewItem   chan bool
	createItemLastError chan error
	ctx                 context.Context
	cancel              context.CancelFunc
}

//Pool is interface of pool available methods
type Pool interface {
	Get() (*item, error)
	Len() int
	Close()
}

//NewPool returns new pool instanse
func NewPool(ctx context.Context, config Config) (Pool, error) {
	var p *pool

	if ok, err := p.validateConfig(config); ok != true {
		return p, err
	}

	ctx, cancel := context.WithCancel(ctx)

	p = &pool{
		config:            config,
		items:             make(chan *item, config.Capacity),
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

//Get returns pool item or error of item getting/creation
func (p *pool) Get() (*item, error) {
	var item *item
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

//Len returns pool length
func (p *pool) Len() int {
	c := <-p.poolLengthCounter
	p.poolLengthCounter <- c
	return c
}

//Close closes and clears pool
func (p *pool) Close() {
	p.cancel()
	p.destroyItems(true)
}

func (p *pool) putPending() {
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

func (p *pool) checkMinCapacity() {
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

func (p *pool) cleanUp() {
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

func (p *pool) createItem() (*item, error) {
	var res *item
	var err error

	object, err := p.config.Factory.Create(p.ctx)

	if err != nil {
		return res, err
	}

	if reflect.ValueOf(object).Kind() != reflect.Ptr {
		return res, errors.New("ggpool.Config.Factory must return object pointer")
	}

	if _, ok := object.(Object); !ok {
		return res, errors.New("ggpool.Config.Factory must create object which implement ggpool.Object interface")
	}

	res = &item{
		object:       &object,
		pool:         p,
		releasedTime: time.Now().UTC(),
	}

	return res, err
}

func (p *pool) destroyItems(force bool) {
	var itemsBuffer []*item

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

func (p *pool) keepMinCapacity() {
	c := <-p.poolLengthCounter
	p.poolLengthCounter <- c

	delta := p.config.MinCapacity - c

	for i := 0; i < delta; i++ {
		p.hasPendingNewItem <- true
	}
}

func (p *pool) validateConfig(config Config) (bool, error) {
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
