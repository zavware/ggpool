//Package ggpool provides functionality of simple generic objects pool
package ggpool

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
)

type timeoutError string

func (e timeoutError) Error() string {
	return string(e)
}

//TimeoutError is type of temporary error
const TimeoutError = timeoutError("timeout exceeded - cannot get pool item")

//Pool is a pool of generic objects
type Pool struct {
	config                Config
	itemReleasedCh        chan bool
	itemDestroyedCh       chan bool
	createItemLastErrorCh chan error
	ctx                   context.Context
	cancel                context.CancelFunc

	sync.RWMutex
	itemCollection *collection
	isInitialized  bool
}

//NewPool returns a new Pool instanse
func NewPool(ctx context.Context, config Config) (*Pool, error) {
	var p *Pool

	ctx, cancel := context.WithCancel(ctx)

	p = &Pool{
		config:                config,
		itemReleasedCh:        make(chan bool),
		itemDestroyedCh:       make(chan bool),
		createItemLastErrorCh: make(chan error),
		ctx:                   ctx,
		cancel:                cancel,
		itemCollection:        newCollection(),
		isInitialized:         false,
	}

	if err := config.validate(); err != nil {
		p.Close()
		return p, err
	}

	go p.keepMinCapacity()
	go p.cleanUp()

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
	p.release(object, true)
}

//Destroy removes and destroys Pool Object
func (p *Pool) Destroy(object *interface{}) {
	objectList := []*interface{}{object}

	p.destroy(objectList)
}

//Len returns pool current length
func (p *Pool) Len() int {
	return p.itemCollection.len()
}

//Close clears and closes pool
func (p *Pool) Close() error {
	if p.itemCollection.len() > p.itemCollection.lenIdle() {
		return errors.New("pool cannot be closed - there are unreleased items")
	}

	p.cancel()

	p.itemCollection.close()
	items := p.itemCollection.getAll()
	for _, item := range items {
		item.destroy()
	}

	return nil
}

func (p *Pool) release(object *interface{}, updateReleaseTime bool) {
	key := getObjectKey(object)
	item := p.itemCollection.get(key)

	if item != nil {
		if updateReleaseTime {
			item.release()
		}

		p.itemCollection.release(key)

		select {
		case p.itemReleasedCh <- true:
			break
		default:
			break
		}
	}
}

func (p *Pool) destroy(objectList []*interface{}) {
	isItemDestroyed := false

	for _, object := range objectList {
		key := getObjectKey(object)
		item := p.itemCollection.get(key)

		if item != nil {
			item.destroy()
			p.itemCollection.remove(key)

			isItemDestroyed = true
		}
	}

	if isItemDestroyed {
		select {
		case p.itemDestroyedCh <- true:
			break
		default:
			break
		}
	}
}

func (p *Pool) getIdleItemWithTimeout(timeout time.Duration) (*item, error) {
	timeoutCtx, cancel := context.WithTimeout(p.ctx, timeout)
	defer cancel()

	itemCh := make(chan *item)
	errCh := make(chan error)

	//waiting for idle item
	go func() {
		p.RLock()
		isPollInitialized := p.isInitialized
		p.RUnlock()

		if isPollInitialized {
			//try to acquire item immediately
			if item := p.itemCollection.acquire(); item != nil {
				itemCh <- item
				return
			}
			go p.putItem()
		}

		//waiting for idle item or timeout
		for {
			select {
			case <-timeoutCtx.Done():
				errCh <- TimeoutError
				return
			case err := <-p.createItemLastErrorCh:
				errCh <- err
				return
			case <-p.itemReleasedCh:
				if item := p.itemCollection.acquire(); item != nil {
					itemCh <- item
					return
				}
				go p.putItem()
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

func (p *Pool) putItem() {
	p.Lock()
	defer p.Unlock()

	if p.itemCollection.len() < p.config.Capacity {
		if item, err := p.createItem(); err == nil {
			for len(p.createItemLastErrorCh) > 0 {
				<-p.createItemLastErrorCh
			}

			if p.itemCollection.put(getObjectKey(item.object), item) {
				p.release(item.object, true)
				//we assume that pool is initialized when a first object has been added to pool collection
				p.isInitialized = true
			} else {
				item.destroy()
			}
		} else {
			select {
			case p.createItemLastErrorCh <- err:
				break
			default:
				break
			}
		}
	}
}

func (p *Pool) keepMinCapacity() {
	keepMinCapacity := func() {
		delta := p.config.MinCapacity - p.itemCollection.len()

		for i := 0; i < delta; i++ {
			go p.putItem()
		}
	}

	keepMinCapacity()

	for {
		select {
		case <-p.itemDestroyedCh:
			keepMinCapacity()
		case <-p.ctx.Done():
			return
		}
	}
}

//cleanUp clears inactive pool elements
func (p *Pool) cleanUp() {
	if p.config.ItemLifetime == 0 {
		return
	}

	ticker := time.NewTicker(p.config.ItemLifetimeCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			var itemsToDestroy []*interface{}

			for _, item := range p.itemCollection.acquireAll() {
				if item.isActive() {
					p.release(item.object, false)
				} else {
					itemsToDestroy = append(itemsToDestroy, item.object)
				}
			}

			p.destroy(itemsToDestroy)

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
