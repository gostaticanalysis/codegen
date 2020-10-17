// Package codegentest provides utilities for testing generators.
package codegentest

import (
	"bytes"
	"fmt"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gostaticanalysis/codegen"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

var TestData = analysistest.TestData

type Testing = analysistest.Testing

// A Result holds the result of applying a generator to a package.
type Result struct {
	Dir    string
	Pass   *codegen.Pass
	Facts  map[types.Object][]analysis.Fact
	Err    error
	Output *bytes.Buffer
}

// Run applies a generator to the packages denoted by the "go list" patterns.
//
// It loads the packages from the specified GOPATH-style project
// directory using golang.org/x/tools/go/packages, runs the generator on
// them.
func Run(t Testing, dir string, g *codegen.Generator, patterns ...string) []*Result {
	outputs := map[*types.Package]*bytes.Buffer{}
	outputFunc := g.Output
	_g := *g
	g = &_g
	g.Output = func(pkg *types.Package) io.Writer {
		var buf bytes.Buffer
		if outputFunc == nil {
			outputs[pkg] = &buf
			return &buf
		}
		w := outputFunc(pkg)
		return io.MultiWriter(w, &buf)
	}

	a := g.ToAnalyzer()
	rs := analysistest.Run(t, dir, a, patterns...)
	results := make([]*Result, len(rs))
	for i := range rs {
		gpass := &codegen.Pass{
			Generator:         g,
			Fset:              rs[i].Pass.Fset,
			Files:             rs[i].Pass.Files,
			OtherFiles:        rs[i].Pass.OtherFiles,
			Pkg:               rs[i].Pass.Pkg,
			TypesInfo:         rs[i].Pass.TypesInfo,
			TypesSizes:        rs[i].Pass.TypesSizes,
			ResultOf:          rs[i].Pass.ResultOf,
			Output:            outputs[rs[i].Pass.Pkg],
			ImportObjectFact:  rs[i].Pass.ImportObjectFact,
			ImportPackageFact: rs[i].Pass.ImportPackageFact,
		}
		results[i] = &Result{
			Dir:    filepath.Join(dir, "src", filepath.ToSlash(rs[i].Pass.Pkg.Path())),
			Pass:   gpass,
			Facts:  rs[i].Facts,
			Err:    rs[i].Err,
			Output: outputs[rs[i].Pass.Pkg],
		}
	}

	return results
}

// Golden compares the results with golden files.
// Golden creates read a golden file which name is codegen.Generator.Name + ".golden".
// The golden file is stored in same directory of the package.
// If Golden cannot find a golden file or the result of Generator test is not same with the golden,
// Golden reports error via *testing.T.
// If update is true, golden files would be updated.
//
// 	var flagUpdate bool
//
// 	func TestMain(m *testing.M) {
// 		flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
// 		flag.Parse()
// 		os.Exit(m.Run())
// 	}
//
// 	func TestGenerator(t *testing.T) {
// 		rs := codegentest.Run(t, codegentest.TestData(), example.Generator, "example")
// 		codegentest.Golden(t, rs, flagUpdate)
// 	}
func Golden(t *testing.T, results []*Result, update bool) {
	t.Helper()
	for _, r := range results {
		golden(t, r, update)
	}
}

func golden(t *testing.T, r *Result, update bool) {
	t.Helper()

	fname := fmt.Sprintf("%s.golden", r.Pass.Generator.Name)
	fpath := filepath.Join(r.Dir, fname)
	gf, err := ioutil.ReadFile(fpath)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	got := r.Output.String()
	r.Output = bytes.NewBufferString(got)

	if !update {
		if diff := cmp.Diff(string(gf), got); diff != "" {
			gname := r.Pass.Generator.Name
			t.Errorf("%s's output is different from the golden file(%s):\n%s", gname, fpath, diff)
		}
		return
	}

	newGolden, err := os.Create(fpath)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if _, err := io.Copy(newGolden, strings.NewReader(got)); err != nil {
		t.Fatal("unexpected error:", err)
	}
	if err := newGolden.Close(); err != nil {
		t.Fatal("unexpected error:", err)
	}
}
