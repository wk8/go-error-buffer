package go_error_buffer

import (
	"time"

	"github.com/friendsofgo/errors"
	"github.com/hashicorp/go-multierror"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// An ErrorBuffer is meant to make sure that something doesn't error out
// too much, too fast.
type ErrorBuffer struct {
	maxCount uint
	window   time.Duration

	errors            *orderedmap.OrderedMap[int64, error]
	previousTimestamp int64
}

// NewErrorBuffer creates a new error buffer, that will accept up to maxCount errors in
// the given time window.
func NewErrorBuffer(maxCount uint, window time.Duration) *ErrorBuffer {
	return &ErrorBuffer{
		maxCount: maxCount,
		window:   window,
		errors:   orderedmap.New[int64, error](orderedmap.WithCapacity[int64, error](int(maxCount) + 1)),
	}
}

// to be able to mock in tests
var now = time.Now

// Add adds an error to the buffer. If there have been too many errors (i.e. more than the buffer's maxCount)
// in too short a time (i.e. in the buffer's current sliding window), then it returns an error aggregating those;
// otherwise returns nil.
// This is not thread-safe.
func (b *ErrorBuffer) Add(err error) error {
	if err == nil {
		return nil
	}

	now := now()
	timestamp := now.UnixNano()
	if b.previousTimestamp == timestamp {
		// ensure no duplicates
		timestamp++
	}
	b.previousTimestamp = timestamp

	_, _ = b.errors.Set(timestamp, err)
	b.prune(now)

	newLen := uint(b.errors.Len())
	if newLen <= b.maxCount {
		return nil
	}
	if newLen == 1 {
		return err
	}

	// too many errors
	errs := make([]error, 0, newLen)
	for pair := b.errors.Oldest(); pair != nil; pair = pair.Next() {
		errs = append(
			errs,
			errors.Wrapf(pair.Value, "at %v", time.Unix(0, pair.Key)),
		)
	}

	span := b.errors.Newest().Key - b.errors.Oldest().Key
	return errors.Wrapf(
		&multierror.Error{Errors: errs},
		"too many errors! %d errors in %v", newLen, time.Duration(span),
	)
}

func (b *ErrorBuffer) prune(now time.Time) {
	cutoff := now.Add(-b.window).UnixNano()
	for pair := b.errors.Oldest(); pair != nil && pair.Key < cutoff; pair = pair.Next() {
		_, _ = b.errors.Delete(pair.Key)
	}
}
