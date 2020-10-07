// Package singlegenerator defines the main function for only a single code generator.
//
// For example, if example.org/mockgen is a generator package,
// all that is needed to define a standalone tool is a file,
// example.org/findbadness/cmd/mockgen/main.go, containing:
//
//      // The mockgen command runs a code generator.
// 	package main
//
// 	import (
// 		"example.org/mockgen"
// 		"github.com/gostaticanalysis/singlegenerator"
// 	)
//
// 	func main() { singlegenerator.Main(mockgen.Generator) }
//
package singlegenerator

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gostaticanalysis/codegen"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

// Main is the main function for a code generation command for a single generator.
// It is a wrapper of singlechecker.Main.
// See golang.org/x/tools/go/analysis/singlechecker.
func Main(g *codegen.Generator) {
	g.Flags.Parse(os.Args[1:])
	os.Args = make([]string, g.Flags.NArg()+1)
	os.Args[0] = os.Args[0]
	copy(os.Args[1:], g.Flags.Args())
	flag.CommandLine.SetOutput(ioutil.Discard)

	a := g.ToAnalyzer(os.Stdout)
	var requires []*analysis.Analyzer

	requires, a.Requires = a.Requires, nil // Requires will be set after validation
	if err := analysis.Validate([]*analysis.Analyzer{a}); err != nil {
		errMsg := strings.ReplaceAll(err.Error(), "analyzer", "generator")
		if g.Name != "" {
			fmt.Fprintln(os.Stderr, g.Name+":", errMsg)
		} else {
			fmt.Fprintln(os.Stderr, errMsg)
		}
		os.Exit(1)
	}
	a.Requires = requires

	g.Flags.Usage = func() {
		paras := strings.Split(g.Doc, "\n\n")
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", g.Name, paras[0])
		fmt.Fprintf(os.Stderr, "Usage: %s [-flag] [package]\n\n", g.Name)
		if len(paras) > 1 {
			fmt.Fprintln(os.Stderr, strings.Join(paras[1:], "\n\n"))
		}
		fmt.Fprintln(os.Stderr, "\nFlags:")
		g.Flags.PrintDefaults()
	}

	if g.Flags.NArg() == 0 {
		g.Flags.Usage()
		os.Exit(1)
	}

	singlechecker.Main(a)
}
