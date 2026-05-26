// Package layout computes the memory layout of named structs in a Target,
// returning the per-field offset/size/align/padding as data (no rendering).
package layout

import (
	"go/token"
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

	var out []common.Layout
	// discovery via types.Info: finds all named structs, including local ones
	// inside function bodies, which t.Types.Scope() would miss.
	names := i.discoverStructNames(t)

	// To avoid duplicates (multiple Defs for the same type name in different
	// scopes), we track what we've seen.
	seen := make(map[token.Pos]bool)

	for _, n := range names {
		if len(opts.Patterns) > 0 && !match.MatchAny(opts.Patterns, n) {
			continue
		}
		// Search for the TypeName in the Info.Defs map
		for _, obj := range t.TypesInfo.Defs {
			tn, ok := obj.(*types.TypeName)
			if !ok || tn.Name() != n || seen[tn.Pos()] {
				continue
			}
			if l, ok := i.buildLayout(t, n, tn, opts); ok {
				out = append(out, l)
				seen[tn.Pos()] = true
			}
		}
	}
	return out
}

func (i *Inspector) discoverStructNames(t common.Target) []string {
	var names []string
	for id, obj := range t.TypesInfo.Defs {
		if tn, ok := obj.(*types.TypeName); ok {
			if _, ok := tn.Type().Underlying().(*types.Struct); ok {
				names = append(names, id.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}

func (i *Inspector) buildLayout(t common.Target, n string, tn *types.TypeName, opts common.Options) (common.Layout, bool) {
	named, _ := tn.Type().(*types.Named)
	st, display, typeParams, note, assumed := resolveStruct(named, tn.Type())
	if st == nil {
		return common.Layout{}, false
	}
	if !opts.IncludeGenerated && structfilter.InGeneratedFile(t, tn.Pos()) {
		return common.Layout{}, false
	}
	if opts.SkipCachePadded && structfilter.HasCacheLinePad(st) {
		return common.Layout{}, false
	}
	l := computeLayout(n, st, display, assumed, t.Sizes)
	l.TypeParams = typeParams
	l.Note = note
	return l, true
}

// resolveStruct returns the struct to measure (st) and the struct whose field
// types drive the rendered declarations (display), plus — for a generic type —
// its type-parameter names, a disclaimer note, and the assumed substitution per
// parameter (indexed by TypeParam.Index, e.g. ["T=any"]). A generic type is
// instantiated with a representative type per parameter so st has a concrete
// layout, while display stays the origin (un-substituted) struct so fields read
// as written (`Value T`, not `Value any`). For a non-generic type, display == st
// and assumed is nil. Returns a nil st when typ is not a struct or cannot be
// instantiated.
func resolveStruct(named *types.Named, typ types.Type) (st, display *types.Struct, typeParams, note string, assumed []string) {
	if named == nil || named.TypeParams().Len() == 0 {
		st, _ = typ.Underlying().(*types.Struct)
		return st, st, "", "", nil
	}
	tps := named.TypeParams()
	args := make([]types.Type, tps.Len())
	pnames := make([]string, tps.Len())
	assumed = make([]string, tps.Len())
	for i := range tps.Len() {
		rep := representativeType(tps.At(i))
		args[i] = rep
		pnames[i] = tps.At(i).Obj().Name()
		assumed[i] = tps.At(i).Obj().Name() + "=" + types.TypeString(rep, nil)
	}
	inst, err := types.Instantiate(nil, named.Origin(), args, false)
	if err != nil {
		return nil, nil, "", "", nil
	}
	st, ok := inst.Underlying().(*types.Struct)
	if !ok {
		return nil, nil, "", "", nil
	}
	display, _ = named.Origin().Underlying().(*types.Struct)
	typeParams = "[" + strings.Join(pnames, ", ") + "]"
	note = "generic type — layout assumes " + strings.Join(assumed, ", ") +
		"; the real layout depends on the type argument(s)"
	return st, display, typeParams, note, assumed
}

// representativeType picks a concrete stand-in for a type parameter: its
// constraint's single core type (e.g. int for `~int`), or interface{} when the
// constraint has no single core type (e.g. `any`, `comparable`, unions).
func representativeType(tp *types.TypeParam) types.Type {
	anyType := types.Universe.Lookup("any").Type()
	c, ok := tp.Constraint().Underlying().(*types.Interface)
	if !ok || c.Empty() {
		return anyType
	}
	// Best-effort core type: if it embeds exactly one type, use its underlying.
	// This covers `T ~int`, `T int`, `T MyInt`.
	if c.NumEmbeddeds() == 1 {
		emb := c.EmbeddedType(0)
		if u, ok := emb.Underlying().(*types.Union); ok {
			if u.Len() == 1 {
				return u.Term(0).Type().Underlying()
			}
		} else {
			return emb.Underlying()
		}
	}
	return anyType
}

// computeLayout sizes st (the instantiated struct) but renders field types and
// assumption markers from display (the origin struct, equal to st when not
// generic). assumed holds the per-parameter substitution text used by the
// markers; it is nil for a non-generic struct.
func computeLayout(name string, st, display *types.Struct, assumed []string, s common.Sizes) common.Layout {
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
		dv := display.Field(i) // origin field: drives the rendered type + marker
		lf := common.LayoutField{
			Name:   vars[i].Name(),
			Type:   types.TypeString(dv.Type(), qualifierForLayout(dv)),
			Tag:    st.Tag(i),
			Assume: fieldAssume(dv.Type(), assumed),
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

// fieldAssume reports the assumed substitutions for the type parameters a field
// type references, in declaration order (e.g. "T=any" for `Value T`, or
// "K=any, V=any" for `map[K]V`). Returns "" when typ references no type
// parameter, or for a non-generic struct (assumed is nil).
func fieldAssume(typ types.Type, assumed []string) string {
	if len(assumed) == 0 {
		return ""
	}
	used := make([]bool, len(assumed))
	collectTypeParams(typ, used, make(map[types.Type]bool))
	parts := make([]string, 0, len(assumed))
	for i, on := range used {
		if on {
			parts = append(parts, assumed[i])
		}
	}
	return strings.Join(parts, ", ")
}

// collectTypeParams marks used[tp.Index()] for every type parameter referenced
// (directly or nested) by typ. seen guards against cycles in recursive types.
func collectTypeParams(typ types.Type, used []bool, seen map[types.Type]bool) {
	if typ == nil || seen[typ] {
		return
	}
	seen[typ] = true
	switch t := typ.(type) {
	case *types.TypeParam:
		if idx := t.Index(); idx >= 0 && idx < len(used) {
			used[idx] = true
		}
	case *types.Map:
		collectTypeParams(t.Key(), used, seen)
		collectTypeParams(t.Elem(), used, seen)
	case interface{ Elem() types.Type }: // *Pointer, *Slice, *Array, *Chan
		collectTypeParams(t.Elem(), used, seen)
	case *types.Named:
		targs := t.TypeArgs()
		for i := range targs.Len() {
			collectTypeParams(targs.At(i), used, seen)
		}
	case *types.Struct:
		for i := range t.NumFields() {
			collectTypeParams(t.Field(i).Type(), used, seen)
		}
	case *types.Signature:
		collectTupleTypeParams(t.Params(), used, seen)
		collectTupleTypeParams(t.Results(), used, seen)
	}
}

// collectTupleTypeParams marks type parameters referenced by any element of a
// signature tuple (params or results); a nil tuple is a no-op.
func collectTupleTypeParams(tup *types.Tuple, used []bool, seen map[types.Type]bool) {
	if tup == nil {
		return
	}
	for i := range tup.Len() {
		collectTypeParams(tup.At(i).Type(), used, seen)
	}
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
