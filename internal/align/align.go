// Package align runs the upstream fieldalignment analyzer over a Target and
// returns the suggested struct reorderings as data (no rendering).
package align

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/peczenyj/structalign/internal/match"
	"github.com/peczenyj/structalign/internal/structfilter"
	"github.com/peczenyj/structalign/pkg/common"
)

// sizeNumRE matches integer tokens in the fieldalignment diagnostic message,
// e.g. "struct of size 24 could be 16" → ["24", "16"].
var sizeNumRE = regexp.MustCompile(`\d+`)

// Aligner implements common.Aligner.
type Aligner struct{}

// New returns an Aligner.
func New() *Aligner { return &Aligner{} }

// Findings runs the analyzer over t and returns one Finding per suggested
// reordering, controlled by opts.
func (a *Aligner) Findings(t common.Target, opts common.Options) ([]common.Finding, error) {
	if len(t.Syntax) == 0 {
		return nil, nil
	}
	names := structNameIndex(t.Syntax)
	nolints := nolintIndex(t.Syntax, t.Fset)
	insp := inspector.New(t.Syntax)

	var findings []common.Finding
	pass := &analysis.Pass{
		Analyzer:   fieldalignment.Analyzer,
		Fset:       t.Fset,
		Files:      t.Syntax,
		Pkg:        t.Types,
		TypesInfo:  t.TypesInfo,
		TypesSizes: t.Sizes, // common.Sizes satisfies types.Sizes
		ResultOf:   map[*analysis.Analyzer]any{inspect.Analyzer: insp},
		Report: func(d analysis.Diagnostic) {
			f := buildFinding(t, d, names, nolints, opts)
			if f == nil {
				return
			}
			findings = append(findings, *f)
		},
	}
	if _, err := fieldalignment.Analyzer.Run(pass); err != nil {
		return nil, fmt.Errorf("%s: analyzer: %w", t.PkgPath, err)
	}
	sort.Slice(findings, func(i, j int) bool {
		pi, pj := t.Fset.Position(findings[i].Pos), t.Fset.Position(findings[j].Pos)
		if pi.Filename != pj.Filename {
			return pi.Filename < pj.Filename
		}
		return pi.Offset < pj.Offset
	})
	return findings, nil
}

// buildFinding converts one analyzer diagnostic into a Finding, applying tag
// stripping and all active filters. Returns nil when the finding should be
// suppressed.
func buildFinding(t common.Target, d analysis.Diagnostic, names map[token.Pos]string, nolints map[token.Pos]nolintInfo, opts common.Options) *common.Finding {
	f := common.Finding{Fset: t.Fset, Pos: d.Pos, Message: d.Message}
	if len(d.SuggestedFixes) > 0 && len(d.SuggestedFixes[0].TextEdits) > 0 {
		e := d.SuggestedFixes[0].TextEdits[0]
		f.Pos = e.Pos
		f.Proposed = string(e.NewText)
		f.Original = readSource(t.Fset, e.Pos, e.End)
		if !opts.KeepTags {
			if s, err := stripStructTags(f.Original); err == nil {
				f.Original = s
			}
			if s, err := stripStructTags(f.Proposed); err == nil {
				f.Proposed = s
			}
		}
	}
	f.Name = names[f.Pos]
	f.TypeParams = typeParamNames(t, f.Name)
	f.OldSize, f.NewSize = parseSizes(f.Message)

	if len(opts.Patterns) > 0 && !match.MatchAny(opts.Patterns, f.Name) {
		return nil
	}
	if !opts.IncludeGenerated && structfilter.InGeneratedFile(t, f.Pos) {
		return nil
	}
	if opts.SkipCachePadded && isCachePadded(t, f.Name) {
		return nil
	}
	if opts.RespectNolint {
		if info, ok := nolints[f.Pos]; ok && info.suppressed(opts.NolintLinters) {
			return nil
		}
	}
	return &f
}

// isCachePadded reports whether the named type in t's scope contains a
// cpu.CacheLinePad field. Returns false for anonymous structs or missing names.
func isCachePadded(t common.Target, name string) bool {
	if name == "" {
		return false
	}
	obj := t.Types.Scope().Lookup(name)
	if obj == nil {
		return false
	}
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return false
	}
	st, ok := tn.Type().Underlying().(*types.Struct)
	return ok && structfilter.HasCacheLinePad(st)
}

// typeParamNames returns the bracketed type-parameter names of the named type,
// e.g. "[T]" or "[K, V]"; empty for a non-generic or anonymous type.
func typeParamNames(t common.Target, name string) string {
	if name == "" {
		return ""
	}
	tn, ok := t.Types.Scope().Lookup(name).(*types.TypeName)
	if !ok {
		return ""
	}
	named, ok := tn.Type().(*types.Named)
	if !ok {
		return ""
	}
	tps := named.TypeParams()
	if tps == nil || tps.Len() == 0 {
		return ""
	}
	parts := make([]string, tps.Len())
	for i := range tps.Len() {
		parts[i] = tps.At(i).Obj().Name()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// parseSizes extracts the two integers the fieldalignment message reports,
// e.g. "struct of size 24 could be 16" -> (24, 16). Returns (0,0) if absent.
func parseSizes(msg string) (old, neu int64) {
	nums := sizeNumRE.FindAllString(msg, 2)
	if len(nums) == 2 {
		old, _ = strconv.ParseInt(nums[0], 10, 64)
		neu, _ = strconv.ParseInt(nums[1], 10, 64)
	}
	return old, neu
}

// structNameIndex maps the position of each named struct type's StructType node
// to its declared name, for both `type Foo struct{...}` and grouped
// `type ( Foo struct{...}; Bar struct{...} )` declarations. The fieldalignment
// analyzer reports diagnostics at StructType.Pos(), so this lets us label and
// filter findings by type name. Anonymous structs and struct literals are not
// indexed (they have no name).
func structNameIndex(files []*ast.File) map[token.Pos]string {
	index := make(map[token.Pos]string)
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			if st, ok := ts.Type.(*ast.StructType); ok {
				index[st.Pos()] = ts.Name.Name
			}
			return true
		})
	}
	return index
}

// readSource returns the raw source text between two positions.
func readSource(fset *token.FileSet, pos, end token.Pos) string {
	pf := fset.File(pos)
	if pf == nil {
		return ""
	}
	data, err := os.ReadFile(pf.Name())
	if err != nil {
		return ""
	}
	start := pf.Offset(pos)
	stop := pf.Offset(end)
	if start < 0 || stop > len(data) || start > stop {
		return ""
	}
	return string(data[start:stop])
}

// stripStructTags removes field tags from a struct type's source text. It
// parses the text (wrapped as a type declaration), clears each field's Tag, and
// reprints it with go/format, which also re-aligns the now-tagless fields. On
// any parse/format error it returns the error so the caller can fall back to
// the original text.
func stripStructTags(src string) (string, error) {
	wrapped := "package p\ntype _ " + src
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wrapped, parser.ParseComments)
	if err != nil {
		return "", err
	}
	var st *ast.StructType
	ast.Inspect(file, func(n ast.Node) bool {
		if s, ok := n.(*ast.StructType); ok && st == nil {
			st = s
			return false
		}
		return true
	})
	if st == nil {
		return "", fmt.Errorf("no struct type found")
	}
	if st.Fields != nil {
		for _, fld := range st.Fields.List {
			fld.Tag = nil
		}
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, st); err != nil {
		return "", err
	}
	return buf.String(), nil
}
