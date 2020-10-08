package mockgen_test

import (
	"flag"
	"os"
	"testing"

	"github.com/gostaticanalysis/codegen/_example/mockgen"
	"github.com/gostaticanalysis/codegen/codegentest"
)

var flagUpdate bool

func TestMain(m *testing.M) {
	flag.BoolVar(&flagUpdate, "update", false, "update the golden files")
	flag.Parse()
	os.Exit(m.Run())
}

func TestGenerator(t *testing.T) {
	mockgen.Generator.Flags.Set("type", "DB")
	rs := codegentest.Run(t, codegentest.TestData(), mockgen.Generator, "a")
	codegentest.Golden(t, rs, flagUpdate)
}
