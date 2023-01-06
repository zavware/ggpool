package ggpool_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/zav0x/ggpool"
)

type NewPoolTestCase struct {
	description            string
	poolConfig             ggpool.Config
	delay                  time.Duration
	expectedPoolLen        int
	expectedCreatedCount   int
	expectedDestroyedCount int
	expectedError          error
}

var newPoolTests = []NewPoolTestCase{

	NewPoolTestCase{
		description: "TestNewPool, case 1: Check keep min capacity",
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            20 * time.Second,
			ItemLifetimeCheckPeriod: 3 * time.Second,
			Timeout:                 0,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        3,
		expectedCreatedCount:   3,
		expectedDestroyedCount: 0,
		expectedError:          nil,
	},

	NewPoolTestCase{
		description: "TestNewPool, case 2: Check keep min capacity + clean-up with expired lifetime",
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            5 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 0,
		},
		delay:                  14 * time.Millisecond,
		expectedPoolLen:        3,
		expectedCreatedCount:   9,
		expectedDestroyedCount: 6,
		expectedError:          nil,
	},

	NewPoolTestCase{
		description: "TestNewPool, case 3: Check keep min capacity + infinity item lifetime",
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            0,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 0,
		},
		delay:                  4 * time.Millisecond,
		expectedPoolLen:        3,
		expectedCreatedCount:   3,
		expectedDestroyedCount: 0,
		expectedError:          nil,
	},

	NewPoolTestCase{
		description: "TestNewPool, case 4: Check keep min capacity for zero min capacity",
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             0,
			ItemLifetime:            20 * time.Second,
			ItemLifetimeCheckPeriod: 3 * time.Second,
			Timeout:                 0,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          nil,
	},

	NewPoolTestCase{
		description: "TestNewPool, case 5: Check capacity validation",
		poolConfig: ggpool.Config{
			Capacity:                0,
			MinCapacity:             6,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 0,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("pool capacity value must be more than 0"),
	},

	NewPoolTestCase{
		description: "TestNewPool, case 6: Check min capacity validation",
		poolConfig: ggpool.Config{
			Capacity:                1,
			MinCapacity:             -1,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 0,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("min pool capacity value must not be negative"),
	},

	NewPoolTestCase{
		description: "TestNewPool, case 7: Check that capacity cannot be less than min capacty",
		poolConfig: ggpool.Config{
			Capacity:                3,
			MinCapacity:             5,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 1 * time.Millisecond,
			Timeout:                 0,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("pool capacity value cannot be less than init capacity value"),
	},

	NewPoolTestCase{
		description: "TestNewPool, case 8: Check ItemLifetimeCheckPeriod validation",
		poolConfig: ggpool.Config{
			Capacity:                5,
			MinCapacity:             3,
			ItemLifetime:            4 * time.Millisecond,
			ItemLifetimeCheckPeriod: 0,
			Timeout:                 0,
		},
		delay:                  2 * time.Millisecond,
		expectedPoolLen:        0,
		expectedCreatedCount:   0,
		expectedDestroyedCount: 0,
		expectedError:          errors.New("please specify ItemLifetimeCheckPeriod"),
	},
}

func TestNewPool(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	for _, testCase := range newPoolTests {
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
				fmt.Sprintf("%s: Unexpected NewPool method error", testCase.description),
			)
		} else if err != nil {
			assertEqual(
				t,
				nil,
				err.Error(),
				fmt.Sprintf("%s: Unexpected NewPool method error", testCase.description),
			)
		}

		//we need to wait for pool items initialization
		time.Sleep(testCase.delay)

		assertEqual(
			t,
			testCase.expectedPoolLen,
			pool.Len(),
			fmt.Sprintf("%s: Unexpected pool length", testCase.description),
		)

		assertEqual(
			t,
			testCase.expectedCreatedCount,
			factory.GetCreatedCount(),
			fmt.Sprintf("%s: Unexpected created items count", testCase.description),
		)

		assertEqual(
			t,
			testCase.expectedDestroyedCount,
			factory.GetDestroyedCount(),
			fmt.Sprintf("%s: Unexpected destroyed items count", testCase.description),
		)

		pool.Close()
	}
}
