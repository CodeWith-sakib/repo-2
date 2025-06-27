package worker

import (
	"sync/atomic"
)

type FencedLeaseToken struct {
	token int64
}

func NewFencedLeaseToken(initial int64) *FencedLeaseToken {
	return &FencedLeaseToken{token: initial}
}

func (flt *FencedLeaseToken) NextToken() int64 {
	return atomic.AddInt64(&flt.token, 1)
}

func (flt *FencedLeaseToken) ValidateToken(expected int64) bool {
	return atomic.LoadInt64(&flt.token) == expected
}
