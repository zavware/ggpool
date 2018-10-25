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
	config                Config
	itemCollection        *collection
	itemReleasedCh        chan bool
	hasPendingNewItemCh   chan bool
	createItemLastErrorCh chan error
	ctx                   context.Context
	cancel                context.CancelFunc
}

//NewPool returns a new Pool instanse
func NewPool(ctx context.Context, config Config) (*Pool, error) {
	var p *Pool

	if err := config.validate(); err != nil {
		return p, err
	}

	ctx, cancel := context.WithCancel(ctx)

	p = &Pool{
		config:                config,
		itemCollection:        newCollection(),
		itemReleasedCh:        make(chan bool),
		hasPendingNewItemCh:   make(chan bool),
		createItemLastErrorCh: make(chan error),
		ctx:    ctx,
		cancel: cancel,
	}

	go func() { p.putPending() }()
	go func() { p.keepMinCapacity() }()
	go func() { p.cleanUp() }()

	return p, nil
}

//Get returns Object or error of Object getting/creation
func (p *Pool) Get() (*interface{}, error) {
	if p.ctx.Err() == context.Canceled {
		return nil, errors.New("pool is closed")
	}

	item, err := p.getIdleItemWithTimeout(p.config.Timeout)

	if err != nil {
		return nil, err
	}
	return item.object, err
}

//Release puts Object back to Pool
func (p *Pool) Release(object *interface{}) {
	key := getObjectKey(object)
	item := p.itemCollection.get(key)

	if item != nil {
		item.release()
		p.itemCollection.release(key)

		select {
		case p.itemReleasedCh <- true:
			break
		default:
			break
		}
	}
}

//Destroy removes and destroys Pool Object
func (p *Pool) Destroy(object *interface{}) {
	key := getObjectKey(object)
	item := p.itemCollection.get(key)

	if item != nil {
		item.destroy()
		p.itemCollection.remove(key)
	}
}

//Len returns pool current length
func (p *Pool) Len() int {
	return p.itemCollection.len()
}

//Close clears and closes pool
func (p *Pool) Close() {
	p.cancel()

	items := p.itemCollection.getAll()
	for _, item := range items {
		go item.destroy()
	}
}

func (p *Pool) getIdleItemWithTimeout(timeout time.Duration) (*item, error) {
	timeoutCtx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	itemCh := make(chan *item)
	errCh := make(chan error)

	//waiting for idle item
	go func() {
		//try to acquire item immediately
		if item := p.itemCollection.acquire(); item != nil {
			itemCh <- item
			return
		}
		p.hasPendingNewItemCh <- true

		//waiting for idle item or timeout
		for {
			select {
			case <-timeoutCtx.Done():
				errCh <- &TimeoutError{"timeout exceeded - cannot get pool item"}
				return
			case err := <-p.createItemLastErrorCh:
				errCh <- err
				return
			case <-p.itemReleasedCh:
				if item := p.itemCollection.acquire(); item != nil {
					itemCh <- item
					return
				}
				p.hasPendingNewItemCh <- true
			}
		}
	}()

	var item *item
	var err error

	select {
	case item = <-itemCh:
		break
	case err = <-errCh:
		break
	}

	return item, err
}

func (p *Pool) putPending() {
	for {
		select {
		case <-p.hasPendingNewItemCh:
			if p.itemCollection.len() < p.config.Capacity {
				if item, err := p.createItem(); err == nil {
					for len(p.createItemLastErrorCh) > 0 {
						<-p.createItemLastErrorCh
					}

					p.itemCollection.put(getObjectKey(item.object), item)
					p.Release(item.object)
				} else {
					p.createItemLastErrorCh <- err
				}
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) keepMinCapacity() {
	ticker := time.NewTicker(time.Duration(10 * time.Second))
	defer ticker.Stop()

	keepMinCapacity := func() {
		delta := p.config.MinCapacity - p.itemCollection.len()

		for i := 0; i < delta; i++ {
			p.hasPendingNewItemCh <- true
		}
	}

	keepMinCapacity()

	for {
		select {
		case <-ticker.C:
			keepMinCapacity()
		case <-p.ctx.Done():
			return
		}
	}
}

//cleanUp clears inactive pool elements
func (p *Pool) cleanUp() {
	ticker := time.NewTicker(p.config.ItemLifetimeCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, item := range p.itemCollection.acquireAll() {
				if item.isActive() {
					p.Release(item.object)
				} else {
					p.itemCollection.remove(getObjectKey(item.object))
					go item.destroy()
				}
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) createItem() (*item, error) {
	var err error

	object, err := p.config.Factory.Create(p.ctx)

	if err != nil {
		return nil, err
	}

	if reflect.ValueOf(object).Kind() != reflect.Ptr {
		return nil, errors.New("ggpool.Config.Factory must return object pointer")
	}

	if _, ok := object.(Object); !ok {
		return nil, errors.New("ggpool.Config.Factory must create object which implement ggpool.Object interface")
	}

	return newItem(&object, p.config.ItemLifetime), err
}
