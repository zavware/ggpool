package ggpool_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

type MockConnection struct {
	factory *MockFactory
}

func (c *MockConnection) Destroy() {
	c.factory.Lock()
	defer c.factory.Unlock()

	c.factory.destroyedCount++

	return
}

func (c *MockConnection) RunCommand() {
	fmt.Println("run command")
}

type MockFactory struct {
	sync.RWMutex
	createdCount   int
	destroyedCount int
}

func (f *MockFactory) Create(ctx context.Context) (interface{}, error) {
	c, err := CreateMockConnection(f)
	return c, err
}

func (f *MockFactory) GetCreatedCount() int {
	f.RLock()
	defer f.RUnlock()

	return f.createdCount
}

func (f *MockFactory) GetDestroyedCount() int {
	f.RLock()
	defer f.RUnlock()

	return f.destroyedCount
}

func CreateMockConnection(f *MockFactory) (*MockConnection, error) {
	f.Lock()
	defer f.Unlock()

	f.createdCount++
	c := &MockConnection{
		factory: f,
	}

	return c, nil
}

func assertEqual(t *testing.T, expected interface{}, actual interface{}, message string) {
	if expected != actual {
		t.Fatalf("%s. Expected: %v, Actual: %v", message, expected, actual)
	}
}
