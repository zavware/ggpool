package ggpool

import (
	"errors"
	"testing"
	"time"

	"github.com/alex-zz/ggpool/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetObject(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockObject := mocks.NewObject(mockCtrl)

	item := &item{
		object: mockObject,
	}

	//ON
	object := item.GetObject()

	//SHOULD
	assert.Equal(t, mockObject, object)
}

func TestRelease(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//GIVEN
	layout := "2006-01-02T15:04:05.000Z"
	releaseTimeStr := "2018-09-30T14:42:42.421Z"
	releaseTime, _ := time.Parse(layout, releaseTimeStr)

	mockClock := mocks.NewClock(mockCtrl)
	mockObject := mocks.NewObject(mockCtrl)
	pool := &pool{
		items: make(chan *item, 1),
	}

	mockClock.EXPECT().Now().Return(releaseTime)

	item := &item{
		object: mockObject,
		clock:  mockClock,
		pool:   pool,
	}

	//ON
	item.Release()

	//SHOULD
	assert.Equal(t, releaseTime, item.releasedTime)

	timeout := time.After(time.Duration(3 * time.Second))

	select {
	case <-timeout:
		t.Error("Test channel is empty. Timeout exceeded")
	case channelItem := <-pool.items:
		assert.Equal(t, item, channelItem)
	}
}

func TestDestroy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//GIVEN
	mockObject := mocks.NewObject(mockCtrl)

	gomock.InOrder(
		mockObject.EXPECT().Destroy().Return(true, nil),
		mockObject.EXPECT().Destroy().Return(false, errors.New("test destroy error")),
	)

	item := &item{
		object: mockObject,
	}

	tests := map[string]struct {
		res bool
		err error
	}{
		"successful destroy": {
			res: true,
			err: nil,
		},
		"unsuccessfull destroy": {
			res: false,
			err: errors.New("test destroy error"),
		},
	}

	for testName, test := range tests {
		t.Logf("Running test case %s", testName)

		//ON
		res, err := item.Destroy()

		//SHOULD
		assert.Equal(t, test.res, res)
		assert.Equal(t, test.err, err)
	}
}
