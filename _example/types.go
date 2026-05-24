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

// A tagged struct, to demonstrate -tags output and tag stripping.
type Tagged struct {
	Flag    bool   `json:"flag"`
	ID      string `json:"id" db:"id"`
	Count   uint32 `json:"count"`
	Ptr     *uint64
	Enabled bool `json:"enabled"`
}
