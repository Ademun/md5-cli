package utils

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

var tickFrequency = 1 * time.Microsecond

type Profiler struct {
	ctx             context.Context
	cancel          context.CancelFunc
	ticker          *time.Ticker
	totalGoroutines int
	ticks           int
}

func NewProfiler() *Profiler {
	ticker := time.NewTicker(tickFrequency)
	return &Profiler{ticker: ticker}
}

func (p *Profiler) Start() {
	if p.ctx != nil {
		p.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.ctx, p.cancel = ctx, cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.ticker.C:
				p.totalGoroutines += runtime.NumGoroutine()
				p.ticks++
			}
		}
	}()
}

func (p *Profiler) Stop() {
	if p.ctx == nil {
		return
	}
	p.cancel()
}

func (p *Profiler) avgGoroutines() int {
	if p.totalGoroutines == 0 {
		return 0
	}
	return p.totalGoroutines / p.ticks
}

func (p *Profiler) totalTime() time.Duration {
	return time.Duration(p.ticks)
}

func (p *Profiler) Stat(timeUnit time.Duration) {
	fmt.Printf("total time: %s, average goroutine number: %d\n", p.totalTime(), p.avgGoroutines())
}
