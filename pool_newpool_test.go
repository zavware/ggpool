package ggpool_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alex-zz/ggpool"
)

type NewPoolTestCase struct {
	poolConfig             ggpool.Config
	delay                  time.Duration
	expectedPoolLen        int
	expectedCreatedCount   int
	expectedDestroyedCount int
	expectedError          error
}

var newPoolTests = []NewPoolTestCase{
	//Test case 1: Check keep min capacity
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            20 * time.Second,
			ItemLifetimeCheckPeriod: 3 * time.Second,
			Timeout:                 3 * time.Second,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        3,
		expectedCreatedCount:   3,
		expectedDestroyedCount: 0,
		expectedError:          nil,
	},

	//Test case 2: Check keep min capacity + clean-up with expired lifetime
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 3 * time.Second,
		},
		delay:                  11 * time.Millisecond,
		expectedPoolLen:        3,
		expectedCreatedCount:   9,
		expectedDestroyedCount: 6,
		expectedError:          nil,
	},

	//Test case 3: Check keep min capacity + infinity item lifetime
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            0,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 3 * time.Second,
		},
		delay:                  4 * time.Millisecond,
		expectedPoolLen:        3,
		expectedCreatedCount:   3,
		expectedDestroyedCount: 0,
		expectedError:          nil,
	},

	//Test case 4: Check keep min capacity for zero min capacity
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             0,
			ItemLifetime:            20 * time.Second,
			ItemLifetimeCheckPeriod: 3 * time.Second,
			Timeout:                 3 * time.Second,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          nil,
	},

	//Test case 5: Check capacity validation
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                0,
			MinCapacity:             6,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 3 * time.Second,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("pool capacity value must be more than 0"),
	},

	//Test case 6: Check min capacity validation
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                1,
			MinCapacity:             -1,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 3 * time.Second,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("min pool capacity value must not be negative"),
	},

	//Test case 7: Check that capacity cannot be less than min capacty
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                3,
			MinCapacity:             5,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 3 * time.Second,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("pool capacity value cannot be less than init capacity value"),
	},

	//Test case 8: Check ItemLifetimeCheckPeriod validation
	NewPoolTestCase{
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 0,
			Timeout:                 3 * time.Second,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("please specify ItemLifetimeCheckPeriod"),
	},
}

func TestNewPool(t *testing.T) {
	for testIndex, testCase := range newPoolTests {
		factory := &MockFactory{
			destroyedCount: 0,
			createdCount:   0,
		}
		testCase.poolConfig.Factory = factory

		pool, err := ggpool.NewPool(context.Background(), testCase.poolConfig)

		if err != nil && testCase.expectedError != nil {
			assertEqual(
				t,
				testCase.expectedError.Error(),
				err.Error(),
				fmt.Sprintf("Test case %d: - Unexpected NewPool method error", testIndex+1),
			)
		} else if err != nil {
			assertEqual(
				t,
				nil,
				err.Error(),
				fmt.Sprintf("Test case %d: - Unexpected NewPool method error", testIndex+1),
			)
		}

		time.Sleep(testCase.delay)

		assertEqual(
			t,
			testCase.expectedPoolLen,
			pool.Len(),
			fmt.Sprintf("Test case %d: - Unxpected pool length", testIndex+1),
		)

		assertEqual(
			t,
			testCase.expectedCreatedCount,
			factory.createdCount,
			fmt.Sprintf("Test case %d: - Unxpected created items count", testIndex+1),
		)

		assertEqual(
			t,
			testCase.expectedDestroyedCount,
			factory.destroyedCount,
			fmt.Sprintf("Test case %d: - Unxpected destroyed items count", testIndex+1),
		)
	}
}
