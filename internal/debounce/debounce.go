package debounce

import (
	"sync"
	"time"
)

type Debounce struct {
	timer    *time.Timer
	duration time.Duration
	mu       sync.Mutex
}

func New(duration time.Duration) *Debounce {
	return &Debounce{
		duration: duration,
	}
}

func (d *Debounce) Debounce(callback func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.duration, func() {
		d.mu.Lock()

		d.timer = nil
		d.mu.Unlock()
		callback()
	})
}
