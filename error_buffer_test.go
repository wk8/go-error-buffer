package go_error_buffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorBuffer(t *testing.T) {
	t.Run("basic usage", func(t *testing.T) {
		maxCount := uint(5)
		buffer := NewErrorBuffer(maxCount, time.Duration(maxCount-1)*time.Minute)

		clock := &testClock{}

		// if we have just an error each minute, the buffer should never reach capacity
		for i := uint(0); i < 2*maxCount; i++ {
			clock.ticks(t, time.Minute)
			require.NoError(t, buffer.Add(errors.Errorf("error at minute %d", i)))
		}

		// now let's add another couple of errors, should overflow the buffer
		clock.ticks(t, 10*time.Second, 20*time.Second)
		require.NoError(t, buffer.Add(errors.New("first extra error")))
		err := buffer.Add(errors.New("second extra error"))
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), fmt.Sprintf("%d errors in 3m30s", maxCount+1))

			var errs *multierror.Error
			if assert.True(t, errors.As(err, &errs)) && assert.Equal(t, int(maxCount+1), len(errs.Errors)) {
				for i, j := maxCount+1, 0; i < 2*maxCount; i, j = i+1, j+1 {
					assert.Contains(t, errs.Errors[j].Error(), fmt.Sprintf("error at minute %d", i))
				}

				assert.Contains(t, errs.Errors[len(errs.Errors)-2].Error(), "first extra error")
				assert.Contains(t, errs.Errors[len(errs.Errors)-1].Error(), "second extra error")
			}
		}
	})

	t.Run("properly handles errors added at exactly the same time", func(t *testing.T) {
		buffer := NewErrorBuffer(1, time.Minute)

		clock := &testClock{}
		clock.ticks(t, 30*time.Second, 0)

		require.NoError(t, buffer.Add(errors.New("first error")))
		err := buffer.Add(errors.New("second error"))
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "2 errors in 1ns")
		}
	})

	t.Run("with 0 maxCount", func(t *testing.T) {
		buffer := NewErrorBuffer(0, 0)

		err := errors.New("test error")
		assert.Equal(t, err, buffer.Add(err))
	})

	t.Run("it ignores nil errors", func(t *testing.T) {
		buffer := NewErrorBuffer(0, 0)
		assert.NoError(t, buffer.Add(nil))
	})
}

// Helpers below

var t0 = time.Now()

// allows overriding the now function to return the given "times" in order.
// The times are in fact durations, relative to an arbitrary t0 time - makes it easier
// to read in tests.
// Also checks at the end of the test that all times have been exhausted - fails the test
// if not.
func withTimes(t *testing.T, times ...time.Duration) {
	previous := now

	nextIndex := 0
	n := len(times)
	now = func() time.Time {
		require.Less(t, nextIndex, n, "no more times!")
		timestamp := t0.Add(times[nextIndex])
		nextIndex++
		return timestamp
	}

	t.Cleanup(func() {
		now = previous

		assert.Equal(t, n, nextIndex, "%d unused times", n-nextIndex)
	})
}

type testClock struct {
	elapsed time.Duration
}

func (c *testClock) ticks(t *testing.T, deltas ...time.Duration) {
	times := make([]time.Duration, len(deltas))
	for i, delta := range deltas {
		c.elapsed += delta
		times[i] = c.elapsed
	}

	withTimes(t, times...)
}
