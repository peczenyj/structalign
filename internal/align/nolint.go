package align

import (
	"go/ast"
	"go/token"
	"strings"
)

// nolintInfo records the //nolint directive found on a struct type declaration.
type nolintInfo struct {
	bare   bool                // a bare //nolint (suppresses all linters)
	tokens map[string]struct{} // named tokens, e.g. {"fieldalignment"}
}

// suppressed reports whether this directive suppresses a finding given the set
// of honored named linters. A bare //nolint always suppresses.
func (info nolintInfo) suppressed(linters []string) bool {
	if info.bare {
		return true
	}
	for _, l := range linters {
		if _, ok := info.tokens[l]; ok {
			return true
		}
	}
	return false
}

// nolintIndex maps each named struct type's StructType.Pos() to the //nolint
// directive on its declaration. A directive is recognized from the type's doc
// comment (TypeSpec.Doc, or the enclosing GenDecl.Doc for grouped `type ( ... )`
// blocks) and from any comment on the type's opening line (e.g. a trailing
// `type T struct { //nolint`, which the AST does not attach to TypeSpec.Comment).
// The analyzer reports at StructType.Pos(), so this key matches a Finding's Pos.
func nolintIndex(files []*ast.File, fset *token.FileSet) map[token.Pos]nolintInfo {
	index := make(map[token.Pos]nolintInfo)
	for _, f := range files {
		// Directives keyed by source line, from every comment in the file —
		// used to catch a trailing directive on a struct's opening line.
		byLine := make(map[int]nolintInfo)
		for _, cg := range f.Comments {
			for _, c := range cg.List {
				line := fset.Position(c.Pos()).Line
				info := byLine[line]
				parseNolint(c.Text, &info)
				byLine[line] = info
			}
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				info := collectNolint(ts.Doc, gd.Doc, ts.Comment)
				mergeNolint(&info, byLine[fset.Position(st.Pos()).Line])
				if info.bare || len(info.tokens) > 0 {
					index[st.Pos()] = info
				}
			}
		}
	}
	return index
}

// collectNolint scans the given comment groups for //nolint directives.
func collectNolint(groups ...*ast.CommentGroup) nolintInfo {
	var info nolintInfo
	for _, g := range groups {
		if g == nil {
			continue
		}
		for _, c := range g.List {
			parseNolint(c.Text, &info)
		}
	}
	return info
}

// mergeNolint folds src into dst (bare wins, tokens unioned).
func mergeNolint(dst *nolintInfo, src nolintInfo) {
	if src.bare {
		dst.bare = true
	}
	for tok := range src.tokens {
		if dst.tokens == nil {
			dst.tokens = make(map[string]struct{})
		}
		dst.tokens[tok] = struct{}{}
	}
}

// parseNolint reads one "//nolint" or "//nolint:a,b,c" comment into info.
// "//nolint" (alone or followed by space) is bare; "//nolint:list" adds named
// tokens. A comment like "//nolintfoo" is not a directive.
func parseNolint(text string, info *nolintInfo) {
	text = strings.TrimSpace(text)
	switch {
	case strings.HasPrefix(text, "//"):
		text = strings.TrimPrefix(text, "//")
	case strings.HasPrefix(text, "/*") && strings.HasSuffix(text, "*/"):
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
	default:
		return
	}
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "nolint") {
		return
	}
	rest := strings.TrimPrefix(text, "nolint")
	switch {
	case rest == "" || strings.HasPrefix(rest, " "):
		info.bare = true
	case strings.HasPrefix(rest, ":"):
		for tok := range strings.SplitSeq(rest[1:], ",") {
			if tok = strings.TrimSpace(tok); tok != "" {
				if info.tokens == nil {
					info.tokens = make(map[string]struct{})
				}
				info.tokens[tok] = struct{}{}
			}
		}
	}
}
