// Package layout computes the memory layout of named structs in a Target,
// returning the per-field offset/size/align/padding as data (no rendering).
package layout

import (
	"go/types"
	"sort"
	"strings"

	"github.com/peczenyj/structalign/internal/match"
	"github.com/peczenyj/structalign/internal/structfilter"
	"github.com/peczenyj/structalign/pkg/common"
)

// Inspector implements common.Inspector.
type Inspector struct{}

// New returns an Inspector.
func New() *Inspector { return &Inspector{} }

// Layouts returns the layout of each named struct in t controlled by opts
// (nil patterns = all). A generic type has no concrete layout until it is
// instantiated, so it is measured with each type parameter replaced by a
// representative type (its constraint's core type, else interface{}) and the
// returned Layout carries a Note disclaiming the approximation.
func (i *Inspector) Layouts(t common.Target, opts common.Options) []common.Layout {
	if t.Types == nil || t.Sizes == nil {
		return nil
	}
	scope := t.Types.Scope()
	names := scope.Names()
	sort.Strings(names)

	var out []common.Layout
	for _, n := range names {
		if len(opts.Patterns) > 0 && !match.MatchAny(opts.Patterns, n) {
			continue
		}
		tn, ok := scope.Lookup(n).(*types.TypeName)
		if !ok {
			continue
		}
		named, _ := tn.Type().(*types.Named)
		st, typeParams, note := resolveStruct(named, tn.Type())
		if st == nil {
			continue // not a struct, or a generic we couldn't instantiate
		}
		if !opts.IncludeGenerated && structfilter.InGeneratedFile(t, tn.Pos()) {
			continue
		}
		if opts.SkipCachePadded && structfilter.HasCacheLinePad(st) {
			continue
		}
		l := computeLayout(n, st, t.Sizes)
		l.TypeParams = typeParams
		l.Note = note
		out = append(out, l)
	}
	return out
}

// resolveStruct returns the struct to measure for a named type, plus (for a
// generic type) its type-parameter names and a disclaimer note. A generic type
// is instantiated with a representative type per parameter so it has a concrete
// layout. Returns a nil struct when typ is not a struct or cannot be instantiated.
func resolveStruct(named *types.Named, typ types.Type) (st *types.Struct, typeParams, note string) {
	if named == nil || named.TypeParams().Len() == 0 {
		st, _ = typ.Underlying().(*types.Struct)
		return st, "", ""
	}
	tps := named.TypeParams()
	args := make([]types.Type, tps.Len())
	pnames := make([]string, tps.Len())
	assumed := make([]string, tps.Len())
	for i := range tps.Len() {
		rep := representativeType(tps.At(i))
		args[i] = rep
		pnames[i] = tps.At(i).Obj().Name()
		assumed[i] = tps.At(i).Obj().Name() + "=" + types.TypeString(rep, nil)
	}
	inst, err := types.Instantiate(nil, named.Origin(), args, false)
	if err != nil {
		return nil, "", ""
	}
	st, ok := inst.Underlying().(*types.Struct)
	if !ok {
		return nil, "", ""
	}
	typeParams = "[" + strings.Join(pnames, ", ") + "]"
	note = "generic type — layout assumes " + strings.Join(assumed, ", ") +
		"; the real layout depends on the type argument(s)"
	return st, typeParams, note
}

// representativeType picks a concrete stand-in for a type parameter: its
// constraint's single core type (e.g. int for `~int`), or interface{} when the
// constraint has no single core type (e.g. `any`, `comparable`, unions).
func representativeType(tp *types.TypeParam) types.Type {
	anyType := types.Universe.Lookup("any").Type()
	c, ok := tp.Constraint().Underlying().(*types.Interface)
	if !ok {
		return anyType
	}
	if c.NumEmbeddeds() == 1 {
		if u, ok := c.EmbeddedType(0).(*types.Union); ok && u.Len() == 1 {
			return u.Term(0).Type()
		}
	}
	return anyType
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
