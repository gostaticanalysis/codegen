package main

import (
	"mockgen"

	"github.com/gostaticanalysis/codegen/singlegenerator"
)

func main() {
	singlegenerator.Main(mockgen.Generator)
}
