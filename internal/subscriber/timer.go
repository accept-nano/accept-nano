package subscriber

import "time"

type Timer struct {
	*time.Timer
	d       time.Duration
	running bool
}

func NewTimer(d time.Duration) *Timer {
	t := &Timer{
		Timer:   time.NewTimer(d),
		d:       d,
		running: true,
	}
	t.Stop()
	return t
}

func (t *Timer) Stop() {
	if t.running && !t.Timer.Stop() {
		<-t.C
	}
	t.running = false
}

func (t *Timer) Delay() {
	if !t.running {
		t.running = true
		t.Timer.Reset(t.d)
	}
}

func (t *Timer) SetNotRunning() {
	t.running = false
}
