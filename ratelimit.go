//go:build go1.23

package flareio

import (
	"time"
)

type limiter struct {
	tickInterval time.Duration
	sleeper      func(time.Duration)
	sleptFor     time.Duration
	nextTick     time.Time
}

func newLimiter(
	tickInterval time.Duration,
) *limiter {
	return &limiter{
		tickInterval: tickInterval,
	}
}

func (l *limiter) sleep(duration time.Duration) {
	if l.sleeper != nil {
		l.sleeper(duration)
	} else {
		time.Sleep(duration)
	}
	l.sleptFor += duration
}

func (l *limiter) tick() {
	untilNextTick := time.Until(l.nextTick)
	l.sleep(untilNextTick)
	l.sleptFor += max(untilNextTick, 0)
}
