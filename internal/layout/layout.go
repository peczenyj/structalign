// Package layout computes the memory layout of named structs in a Target,
// returning the per-field offset/size/align/padding as data (no rendering).
package layout

import (
	"go/types"
	"sort"

	"github.com/peczenyj/structalign/internal/match"
	"github.com/peczenyj/structalign/pkg/common"
)

// Inspector implements common.Inspector.
type Inspector struct{}

// New returns an Inspector.
func New() *Inspector { return &Inspector{} }

// Layouts returns the layout of each named struct in t matching patterns
// (nil = all). Generic types (no concrete layout until instantiated) are
// skipped.
func (i *Inspector) Layouts(t common.Target, patterns []string) []common.Layout {
	if t.Types == nil || t.Sizes == nil {
		return nil
	}
	scope := t.Types.Scope()
	names := scope.Names()
	sort.Strings(names)

	var out []common.Layout
	for _, n := range names {
		if len(patterns) > 0 && !match.MatchAny(patterns, n) {
			continue
		}
		tn, ok := scope.Lookup(n).(*types.TypeName)
		if !ok {
			continue
		}
		if named, ok := tn.Type().(*types.Named); ok && named.TypeParams().Len() > 0 {
			continue // generic: no concrete layout until instantiated
		}
		st, ok := tn.Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}
		out = append(out, computeLayout(n, st, t.Sizes))
	}
	return out
}

func computeLayout(name string, st *types.Struct, s common.Sizes) common.Layout {
	nf := st.NumFields()
	vars := make([]*types.Var, nf)
	for i := range nf {
		vars[i] = st.Field(i)
	}
	offsets := s.Offsetsof(vars)
	total := s.Sizeof(st)
	align := s.Alignof(st)

	fields := make([]common.LayoutField, nf)
	var totalPad int64
	for i := range nf {
		fsize := s.Sizeof(vars[i].Type())
		lf := common.LayoutField{
			Name:   vars[i].Name(),
			Type:   types.TypeString(vars[i].Type(), qualifierForLayout(vars[i])),
			Tag:    st.Tag(i),
			Offset: offsets[i],
			Size:   fsize,
			Align:  s.Alignof(vars[i].Type()),
		}
		end := offsets[i] + fsize
		next := total
		if i+1 < nf {
			next = offsets[i+1]
		}
		if next > end {
			lf.Padding = next - end
		}
		totalPad += lf.Padding
		fields[i] = lf
	}
	return common.Layout{Name: name, Total: total, Align: align, Padding: totalPad, Fields: fields}
}

func qualifierForLayout(v *types.Var) types.Qualifier {
	home := v.Pkg()
	return func(p *types.Package) string {
		if p == home {
			return ""
		}
		return p.Name()
	}
}
