package sample

import "sync"

// Bad: bool, int64, bool -> padding waste. Optimal: int64, bool, bool.
type Mixed struct {
	A bool  `json:"a"`
	B int64 `json:"b"`
	C bool  `json:"c"`
}

// A tracking-ID-ish record with pointer-bytes implications.
type Record struct {
	Flag    bool
	ID      string
	Count   uint32
	Ptr     *uint64
	Enabled bool
}

// Already optimal -- should NOT be reported.
type Good struct {
	B int64
	A bool
	C bool
}

// A tagged struct, to demonstrate -tags output and tag stripping.
type Tagged struct {
	Flag    bool   `json:"flag"`
	ID      string `json:"id" db:"id"`
	Count   uint32 `json:"count"`
	Ptr     *uint64
	Enabled bool `json:"enabled"`
}

// Generic: a container parameterized by T, to see how layout analysis behaves
// when a field's size depends on the type argument.
type Generic[T any] struct {
	Flag  bool
	Value T
	Count uint32
}

// FuncField: a struct with a function-typed field (a func value is
// pointer-sized) wedged between smaller fields.
type FuncField struct {
	Retries uint8
	OnError func(error)
	Verbose bool
}

// MutexLast: a struct ending in a blank sync.Mutex -- a common "guard" idiom.
// Worth seeing whether the analyzer wants to move the mutex around.
type MutexLast struct {
	Count int32
	Ready bool
	_     sync.Mutex
}
