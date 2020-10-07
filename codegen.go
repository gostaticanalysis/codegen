package codegen

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"golang.org/x/tools/go/analysis"
)

// A Generator describes a code generator function and its options.
type Generator struct {
	// The Name of the generator must be a valid Go identifier
	// as it may appear in command-line flags, URLs, and so on.
	Name string

	// Doc is the documentation for the generator.
	// The part before the first "\n\n" is the title
	// (no capital or period, max ~60 letters).
	Doc string

	// Flags defines any flags accepted by the generator.
	// The manner in which these flags are exposed to the user
	// depends on the driver which runs the generator.
	Flags flag.FlagSet

	// Run applies the generator to a package.
	// It returns an error if the generator failed.
	//
	// To pass analysis results of depended analyzers between packages (and thus
	// potentially between address spaces), use Facts, which are
	// serializable.
	Run func(*Pass) error

	// RunDespiteErrors allows the driver to invoke
	// the Run method of this generator even on a
	// package that contains parse or type errors.
	RunDespiteErrors bool

	// Requires is a set of analyzers that must run successfully
	// before this one on a given package. This analyzer may inspect
	// the outputs produced by each analyzer in Requires.
	// The graph over analyzers implied by Requires edges must be acyclic.
	//
	// Requires establishes a "horizontal" dependency between
	// analysis passes (different analyzers, same package).
	Requires []*analysis.Analyzer
}

// ToAnalyzer converts the generator to an analyzer.
func (g *Generator) ToAnalyzer(output io.Writer) *analysis.Analyzer {
	requires := make([]*analysis.Analyzer, len(g.Requires))
	for i := range requires {
		a := *g.Requires[i] // copy
		a.Run = func(pass *analysis.Pass) (interface{}, error) {
			pass.Report = func(analysis.Diagnostic) {}
			return g.Requires[i].Run(pass)
		}
		requires[i] = &a
	}

	return &analysis.Analyzer{
		Name: g.Name,
		Doc:  g.Doc,
		Run: func(pass *analysis.Pass) (interface{}, error) {
			gpass := &Pass{
				Generator:         g,
				Fset:              pass.Fset,
				Files:             pass.Files,
				OtherFiles:        pass.OtherFiles,
				Pkg:               pass.Pkg,
				TypesInfo:         pass.TypesInfo,
				TypesSizes:        pass.TypesSizes,
				ResultOf:          pass.ResultOf,
				Output:            output,
				ImportObjectFact:  pass.ImportObjectFact,
				ImportPackageFact: pass.ImportPackageFact,
			}
			return nil, g.Run(gpass)
		},
		RunDespiteErrors: g.RunDespiteErrors,
		Requires:         requires,
	}
}

// A Pass provides information to the Run function that applies a specific
// generator to a single Go package.
// The Run function should not call any of the Pass functions concurrently.
type Pass struct {
	Generator *Generator // the identity of the current generator

	// syntax and type information
	Fset       *token.FileSet // file position information
	Files      []*ast.File    // the abstract syntax tree of each file
	OtherFiles []string       // names of non-Go files of this package
	Pkg        *types.Package // type information about the package
	TypesInfo  *types.Info    // type information about the syntax trees
	TypesSizes types.Sizes    // function for computing sizes of types

	// ResultOf provides the inputs to this analysis pass, which are
	// the corresponding results of its prerequisite analyzers.
	// The map keys are the elements of Analysis.Required,
	// and the type of each corresponding value is the required
	// analysis's ResultType.
	ResultOf map[*analysis.Analyzer]interface{}

	// Output is the destination of the generator.
	// Pass's Print, Println, Printf outputs to this writer.
	Output io.Writer

	// ImportObjectFact retrieves a fact associated with obj.
	// Given a value ptr of type *T, where *T satisfies Fact,
	// ImportObjectFact copies the value to *ptr.
	//
	// ImportObjectFact panics if called after the pass is complete.
	// ImportObjectFact is not concurrency-safe.
	ImportObjectFact func(obj types.Object, fact analysis.Fact) bool

	// ImportPackageFact retrieves a fact associated with package pkg,
	// which must be this package or one of its dependencies.
	// See comments for ImportObjectFact.
	ImportPackageFact func(pkg *types.Package, fact analysis.Fact) bool
}

// Print is a wrapper of fmt.Fprint with pass.Output.
func (pass *Pass) Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(pass.Output, a...)
}

// Printf is a wrapper of fmt.Fprintf with pass.Output.
func (pass *Pass) Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(pass.Output, format, a...)
}

// Println is a wrapper of fmt.Fprintln with pass.Output.
func (pass *Pass) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(pass.Output, a...)
}
