package common_test

import (
	"flag"
	"testing"

	"github.com/peczenyj/structalign/pkg/common"
)

func TestColorizeString(t *testing.T) {
	// -trimprefix=Colorize -transform=lower => "auto" / "always" / "never".
	cases := map[common.Colorize]string{
		common.ColorizeAuto:   "auto",
		common.ColorizeAlways: "always",
		common.ColorizeNever:  "never",
	}
	for mode, want := range cases {
		if got := mode.String(); got != want {
			t.Errorf("Colorize(%d).String() = %q, want %q", mode, got, want)
		}
	}
}

func TestColorizeParse(t *testing.T) {
	got, err := common.ColorizeString("always")
	if err != nil {
		t.Fatalf("ColorizeString(always): %v", err)
	}
	if got != common.ColorizeAlways {
		t.Errorf("ColorizeString(always) = %v, want ColorizeAlways", got)
	}
	if _, err := common.ColorizeString("bogus"); err == nil {
		t.Error("ColorizeString(bogus): want error, got nil")
	}
}

func TestColorizeText(t *testing.T) {
	b, err := common.ColorizeAuto.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(b) != "auto" {
		t.Errorf("MarshalText = %q, want auto", b)
	}

	var c common.Colorize
	if err := c.UnmarshalText([]byte("never")); err != nil {
		t.Fatalf("UnmarshalText: %v", err)
	}
	if c != common.ColorizeNever {
		t.Errorf("UnmarshalText(never) = %v, want ColorizeNever", c)
	}
}

func TestColorizeImplementsFlagValue(t *testing.T) {
	var c common.Colorize
	var _ flag.Value = &c // compile-time: *Colorize satisfies flag.Value

	if err := c.Set("always"); err != nil {
		t.Fatalf("Set(always): %v", err)
	}
	if c != common.ColorizeAlways {
		t.Errorf("after Set(always) = %v, want ColorizeAlways", c)
	}
	if err := c.Set("bogus"); err == nil {
		t.Error("Set(bogus): want error, got nil")
	}
}
