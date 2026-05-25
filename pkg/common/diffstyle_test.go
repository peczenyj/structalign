package common_test

import (
	"flag"
	"testing"

	"github.com/peczenyj/structalign/pkg/common"
)

func TestDiffStyleString(t *testing.T) {
	// -trimprefix=Diff -transform=lower => "unified" / "side" / "none".
	cases := map[common.DiffStyle]string{
		common.DiffUnified: "unified",
		common.DiffSide:    "side",
		common.DiffNone:    "none",
	}
	for style, want := range cases {
		if got := style.String(); got != want {
			t.Errorf("DiffStyle(%d).String() = %q, want %q", style, got, want)
		}
	}
}

func TestDiffStyleParse(t *testing.T) {
	got, err := common.DiffStyleString("side")
	if err != nil {
		t.Fatalf("DiffStyleString(side): %v", err)
	}
	if got != common.DiffSide {
		t.Errorf("DiffStyleString(side) = %v, want DiffSide", got)
	}
	if _, err := common.DiffStyleString("bogus"); err == nil {
		t.Error("DiffStyleString(bogus): want error, got nil")
	}
}

func TestDiffStyleText(t *testing.T) {
	b, err := common.DiffUnified.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(b) != "unified" {
		t.Errorf("MarshalText = %q, want unified", b)
	}

	var d common.DiffStyle
	if err := d.UnmarshalText([]byte("none")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}
	if d != common.DiffNone {
		t.Errorf("UnmarshalText(none) = %v, want DiffNone", d)
	}
}

func TestDiffStyleImplementsFlagValue(t *testing.T) {
	var d common.DiffStyle
	var _ flag.Value = &d // compile-time: *DiffStyle satisfies flag.Value

	if err := d.Set("side"); err != nil {
		t.Fatalf("Set(side): %v", err)
	}
	if d != common.DiffSide {
		t.Errorf("after Set(side) = %v, want DiffSide", d)
	}
	if err := d.Set("bogus"); err == nil {
		t.Error("Set(bogus): want error, got nil")
	}
}
