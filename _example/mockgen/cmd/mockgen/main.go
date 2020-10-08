package main

import (
	"github.com/gostaticanalysis/codegen/_example/mockgen"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	singlegenerator.Main(mockgen.Generator)
}
