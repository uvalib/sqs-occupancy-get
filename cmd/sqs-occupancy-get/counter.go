package main

import "sync/atomic"

// our counter object
type Counter struct {
	SuccessCount uint64
	ErrorCount   uint64
}

func (c *Counter) AddSuccess(count uint64) {
	atomic.AddUint64(&c.SuccessCount, count)
}

func (c *Counter) AddError(count uint64) {
	atomic.AddUint64(&c.ErrorCount, count)
}

func (c *Counter) Get() (uint64, uint64) {
	s := atomic.LoadUint64(&c.SuccessCount)
	e := atomic.LoadUint64(&c.ErrorCount)
	return s, e
}

//
// end of file
//
