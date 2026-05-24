// Package sizes adapts go/types sizing to the common.Sizes interface, so the
// target architecture can be injected (real sizes in production, a fixed
// amd64 sizer in tests).
package sizes

import (
	"go/types"

	"github.com/peczenyj/structalign/pkg/common"
)

type adapter struct{ s types.Sizes }

// New wraps a go/types.Sizes as a common.Sizes.
func New(s types.Sizes) common.Sizes { return adapter{s: s} }

// ForArch returns deterministic gc sizes for the given GOARCH (e.g. "amd64").
func ForArch(goarch string) common.Sizes { return New(types.SizesFor("gc", goarch)) }

func (a adapter) Sizeof(t types.Type) int64        { return a.s.Sizeof(t) }
func (a adapter) Alignof(t types.Type) int64       { return a.s.Alignof(t) }
func (a adapter) Offsetsof(f []*types.Var) []int64 { return a.s.Offsetsof(f) }
