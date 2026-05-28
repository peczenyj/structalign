package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/peczenyj/structalign/pkg/common"
)

type jsonDocument struct {
	Version  string        `json:"version"`
	Mode     string        `json:"mode"`
	Findings []jsonFinding `json:"findings,omitempty"`
	Layouts  []jsonLayout  `json:"layouts,omitempty"`
	Summary  *jsonSummary  `json:"summary,omitempty"`
}

type jsonFinding struct {
	Package    string `json:"package"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Name       string `json:"name"`
	TypeParams string `json:"typeParams,omitempty"`
	Message    string `json:"message"`
	OldSize    int64  `json:"oldSize"`
	NewSize    int64  `json:"newSize"`
	BytesSaved int64  `json:"bytesSaved"`
	Original   string `json:"original"`
	Proposed   string `json:"proposed"`
}

type jsonSummary struct {
	StructsAffected int   `json:"structsAffected"`
	BytesSaved      int64 `json:"bytesSaved"`
}

type jsonLayout struct {
	Package    string            `json:"package"`
	Name       string            `json:"name"`
	TypeParams string            `json:"typeParams,omitempty"`
	Note       string            `json:"note,omitempty"`
	Size       int64             `json:"size"`
	Align      int64             `json:"align"`
	Padding    int64             `json:"padding"`
	Fields     []jsonLayoutField `json:"fields"`
}

type jsonLayoutField struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Tag     string `json:"tag,omitempty"`
	Assume  string `json:"assume,omitempty"`
	Offset  int64  `json:"offset"`
	Size    int64  `json:"size"`
	Align   int64  `json:"align"`
	Padding int64  `json:"padding"`
}

// RenderJSON emitted a structured JSON document for findings or layouts.
func (p *Printer) RenderJSON(version string, findings []common.Finding, layouts []common.Layout) {
	doc := jsonDocument{
		Version: version,
	}

	if layouts != nil {
		doc.Mode = "inspect"
		doc.Layouts = make([]jsonLayout, len(layouts))
		for i, l := range layouts {
			doc.Layouts[i] = jsonLayout{
				Package:    l.Package,
				Name:       l.Name,
				TypeParams: l.TypeParams,
				Note:       l.Note,
				Size:       l.Total,
				Align:      l.Align,
				Padding:    l.Padding,
				Fields:     make([]jsonLayoutField, len(l.Fields)),
			}
			for j, f := range l.Fields {
				doc.Layouts[i].Fields[j] = jsonLayoutField{
					Name:    f.Name,
					Type:    f.Type,
					Tag:     f.Tag,
					Assume:  f.Assume,
					Offset:  f.Offset,
					Size:    f.Size,
					Align:   f.Align,
					Padding: f.Padding,
				}
			}
		}
	} else {
		doc.Mode = "diff"
		doc.Findings = make([]jsonFinding, len(findings))
		var totalSaved int64
		for i, f := range findings {
			pos := f.Fset.Position(f.Pos)
			saved := f.OldSize - f.NewSize
			if saved < 0 {
				saved = 0
			}
			totalSaved += saved
			doc.Findings[i] = jsonFinding{
				Package:    f.Package,
				File:       relPath(pos.Filename),
				Line:       pos.Line,
				Column:     pos.Column,
				Name:       f.Name,
				TypeParams: f.TypeParams,
				Message:    f.Message,
				OldSize:    f.OldSize,
				NewSize:    f.NewSize,
				BytesSaved: saved,
				Original:   f.Original,
				Proposed:   f.Proposed,
			}
		}
		doc.Summary = &jsonSummary{
			StructsAffected: len(findings),
			BytesSaved:      totalSaved,
		}
	}

	enc := json.NewEncoder(p.Out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		fmt.Fprintf(os.Stderr, "structalign: json encode: %v\n", err)
	}
}
