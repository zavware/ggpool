package ggpool_test

import (
	"context"
	"testing"
	"time"

	"github.com/alex-zz/ggpool"
)

func TestGet(t *testing.T) {

	factory := &MockFactory{
		destroyedCount: 0,
		createdCount:   0,
	}

	pool, err := ggpool.NewPool(context.Background(), ggpool.Config{
		Capacity:                5,
		MinCapacity:             3,
		ItemLifetime:            20 * time.Second,
		ItemLifetimeCheckPeriod: 3 * time.Second,
		Timeout:                 3 * time.Second,
		Factory:                 factory,
	})

	if err != nil {
		t.Fatalf("TestGet: Unexpected NewPool() method error: %s", err)
	}

	object, err := pool.Get()

	if err != nil {
		t.Fatalf("TestGet: Unexpected Get() method error: %s", err)
	}

	if _, ok := (*object).(*MockConnection); !ok {
		t.Fatal("TestGet: Incorrect object type")
	}

	//we need to wait for pool items initialization
	time.Sleep(2 * time.Millisecond)

	assertEqual(
		t,
		3,
		factory.GetCreatedCount(),
		"TestGet: Unexpected created items count",
	)

	assertEqual(
		t,
		0,
		factory.GetDestroyedCount(),
		"TestGet: Unexpected destroyed items count",
	)

	pool.Close()
}

func TestGetDelay(t *testing.T) {

	factory := &MockFactory{
		destroyedCount: 0,
		createdCount:   0,
	}

	pool, err := ggpool.NewPool(context.Background(), ggpool.Config{
		Capacity:                1,
		MinCapacity:             1,
		ItemLifetime:            20 * time.Second,
		ItemLifetimeCheckPeriod: 3 * time.Second,
		Timeout:                 5 * time.Millisecond,
		Factory:                 factory,
	})

	if err != nil {
		t.Fatalf("TestGetDelay: Unexpected NewPool() method error: %s", err)
	}

	object, err := pool.Get()

	if err != nil {
		t.Fatalf("TestGetDelay: Unexpected Get() method error: %s", err)
	}

	if _, ok := (*object).(*MockConnection); !ok {
		t.Fatal("TestGetDelay: Incorrect object type")
	}

	objectCh := make(chan *interface{})
	errorCh := make(chan error)

	go func() {
		object, err := pool.Get()
		if err != nil {
			errorCh <- err
		} else {
			objectCh <- object
		}
	}()

	time.Sleep(3 * time.Millisecond)

	pool.Release(object)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		t.Fatal("TestGetDelay: test timeout")
	case err := <-errorCh:
		t.Fatalf("TestGetDelay: Unexpected Get() method error: %s", err)
	case object := <-objectCh:
		if _, ok := (*object).(*MockConnection); !ok {
			t.Fatal("TestGetDelay: Incorrect object type")
		}
	}

	pool.Close()
}

func TestGetDelayTimeoutError(t *testing.T) {

	factory := &MockFactory{
		destroyedCount: 0,
		createdCount:   0,
	}

	pool, err := ggpool.NewPool(context.Background(), ggpool.Config{
		Capacity:                1,
		MinCapacity:             1,
		ItemLifetime:            20 * time.Second,
		ItemLifetimeCheckPeriod: 3 * time.Second,
		Timeout:                 5 * time.Millisecond,
		Factory:                 factory,
	})

	if err != nil {
		t.Fatalf("TestGetDelayTimeoutError: Unexpected NewPool() method error: %s", err)
	}

	object, err := pool.Get()

	if err != nil {
		t.Fatalf("TestGetDelayTimeoutError: Unexpected Get() method error: %s", err)
	}

	if _, ok := (*object).(*MockConnection); !ok {
		t.Fatal("TestGetDelayTimeoutError: Incorrect object type")
	}

	objectCh := make(chan *interface{})
	errorCh := make(chan error)

	go func() {
		object, err := pool.Get()
		if err != nil {
			errorCh <- err
		} else {
			objectCh <- object
		}
	}()

	time.Sleep(6 * time.Millisecond)

	pool.Release(object)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		t.Fatal("TestGetDelayTimeoutError: test timeout")
	case err := <-errorCh:
		if err != ggpool.TimeoutError {
			t.Fatalf("TestGetDelayTimeoutError: Unexpected Get() method error: %s", err)
		}
	case <-objectCh:
		t.Fatal("TestGetDelayTimeoutError: Unexpected object")
	}

	pool.Close()
}
