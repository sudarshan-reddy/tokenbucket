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
func (b *Bucket) Wait(count int64) {
	if d := b.Take(count); d > 0 {
		time.Sleep(d)
	}
}

//ChangeInterval is used to dynamically alter fill interval midway
func (b *Bucket) ChangeInterval(fillInterval time.Duration) {
	b.fillInterval = fillInterval
}

// Take takes count tokens from the bucket without blocking. It returns
// the time that the caller should wait until the tokens are actually
// available.
func (b *Bucket) Take(count int64) time.Duration {
	b.Lock()
	defer b.Unlock()
	d, _ := b.take(time.Now(), count, infinityDuration)
	return d
}

func (b *Bucket) take(now time.Time, count int64, maxWait time.Duration) (time.Duration, bool) {
	if count <= 0 {
		return 0, true
	}

	tick := b.currentTick(now)
	b.adjustAvailableTokens(tick)
	availableTokens := b.availableTokens - count
	if availableTokens >= 0 {
		b.availableTokens = availableTokens
		return 0, true
	}

	waitTime := b.computeWaitTime(now, tick, availableTokens)
	if waitTime > maxWait {
		return 0, false
	}
	b.availableTokens = availableTokens
	return waitTime, true
}

func (b *Bucket) currentTick(now time.Time) int64 {
	return int64(now.Sub(b.startTime) / b.fillInterval)
}

func (b *Bucket) computeWaitTime(now time.Time, tick int64,
	availableTokens int64) time.Duration {
	tickWhenTokensWillBeAvail := tick +
		(-availableTokens+b.tokensToBeAdded-1)/b.tokensToBeAdded
	timeWhenTokensWillBeAvail := b.startTime.
		Add(time.Duration(tickWhenTokensWillBeAvail) * b.fillInterval)
	return timeWhenTokensWillBeAvail.Sub(now)
}

func (b *Bucket) adjustAvailableTokens(tick int64) {
	if b.availableTokens >= b.capacity {
		return
	}
	b.availableTokens += (tick - b.latestTick) * b.tokensToBeAdded
	if b.availableTokens > b.capacity {
		b.availableTokens = b.capacity
	}
	b.latestTick = tick
	return
}
