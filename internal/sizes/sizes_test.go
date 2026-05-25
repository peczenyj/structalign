package sizes_test

import (
	"go/types"
	"testing"

	"github.com/peczenyj/structalign/internal/sizes"
)

func TestForArchAmd64(t *testing.T) {
	s := sizes.ForArch("amd64")
	if got := s.Sizeof(types.Typ[types.Int64]); got != 8 {
		t.Errorf("Sizeof(int64) = %d, want 8", got)
	}
	if got := s.Alignof(types.Typ[types.Bool]); got != 1 {
		t.Errorf("Alignof(bool) = %d, want 1", got)
	}
	if got := s.Sizeof(types.Typ[types.String]); got != 16 {
		t.Errorf("Sizeof(string) = %d on amd64, want 16", got)
	}
}

func TestNewWrapsGivenSizes(t *testing.T) {
	s := sizes.New(types.SizesFor("gc", "amd64"))
	if got := s.Sizeof(types.Typ[types.Int32]); got != 4 {
		t.Errorf("Sizeof(int32) = %d, want 4", got)
	}
}
