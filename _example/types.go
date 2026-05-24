package sample

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
