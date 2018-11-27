package tokenbucket

import (
	"io"
	"sync"
	"time"
)

const infinityDuration time.Duration = 0x7fffffffffffffff

//Bucket ...
type Bucket struct {
	sync.Mutex

	startTime       time.Time
	capacity        int64
	tokensToBeAdded int64
	fillInterval    time.Duration

	availableTokens int64
	latestTick      int64
}

//NewBucket gives a new instance of bucket
func NewBucket(fillInterval time.Duration, capacity int64) *Bucket {
	return &Bucket{
		startTime:       time.Now(),
		latestTick:      0,
		fillInterval:    fillInterval,
		capacity:        capacity,
		tokensToBeAdded: 1,
		availableTokens: capacity,
	}
}

type writer struct {
	w      io.Writer
	bucket *Bucket
}

// NewWriter returns a reader that is rate limited by
// the given token bucket. Each token in the bucket
// represents one byte.
func NewWriter(w io.Writer, bucket *Bucket) io.Writer {
	return &writer{
		w:      w,
		bucket: bucket,
	}
}

func (w *writer) Write(buf []byte) (int, error) {
	w.bucket.Wait(int64(len(buf)))
	return w.w.Write(buf)
}

// Wait takes count tokens from the bucket, waiting until they are
// available.
func (tb *Bucket) Wait(count int64) {
	if d := tb.Take(count); d > 0 {
		time.Sleep(d)
	}
}

//ChangeInterval is used to dynamically alter fill interval midway
func (tb *Bucket) ChangeInterval(fillInterval time.Duration) {
	tb.fillInterval = fillInterval
}

// Take takes count tokens from the bucket without blocking. It returns
// the time that the caller should wait until the tokens are actually
// available.
func (tb *Bucket) Take(count int64) time.Duration {
	tb.Lock()
	defer tb.Unlock()
	d, _ := tb.take(time.Now(), count, infinityDuration)
	return d
}

// Capacity returns the capacity that the bucket was created with.
func (tb *Bucket) Capacity() int64 {
	return tb.capacity
}

// Rate returns the fill rate of the bucket, in tokens per second.
func (tb *Bucket) Rate() float64 {
	return 1e9 * float64(tb.tokensToBeAdded) / float64(tb.fillInterval)
}

func (tb *Bucket) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}

	tick := tb.currentTick(now)
	tb.adjustAvailableTokens(tick)
	availableTokens := tb.availableTokens - count
	if availableTokens >= 0 {
		tb.availableTokens = availableTokens
		return 0, true
	}

	//calculate how long to wait for the tokens to be available.
	endTick := tick + (-availableTokens+tb.tokensToBeAdded-1)/tb.tokensToBeAdded
	endTime := tb.startTime.Add(time.Duration(endTick) * tb.fillInterval)
	waitTime := endTime.Sub(now)
	if waitTime > maxWait {
		return 0, false
	}
	tb.availableTokens = availableTokens
	return waitTime, true
}

func (tb *Bucket) currentTick(now time.Time) int64 {
	return int64(now.Sub(tb.startTime) / tb.fillInterval)
}

func (tb *Bucket) adjustAvailableTokens(tick int64) {
	if tb.availableTokens >= tb.capacity {
		return
	}
	tb.availableTokens += (tick - tb.latestTick) * tb.tokensToBeAdded
	if tb.availableTokens > tb.capacity {
		tb.availableTokens = tb.capacity
	}
	tb.latestTick = tick
	return
}
