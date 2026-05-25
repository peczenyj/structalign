// Command structalign shows how a struct's fields could be reordered to use
// less memory -- as a diff, not a rewrite -- and can inspect a struct's memory
// layout. It reuses the upstream golang.org/x/tools fieldalignment analyzer; see
// the internal packages for the implementation.
package main

import (
	"os"

	"github.com/peczenyj/structalign/internal/app"
)

func main() {
	os.Exit(app.New(os.Stdout, os.Stderr).Run(os.Args[1:]))
}
