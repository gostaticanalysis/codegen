package a

//go:generate go run ../../cmd/mockgen/main.go -o mockgen.golden -type DB a.go

type DB interface {
	Get(id string) int
	Set(id string, v int)
}

type db struct {}

func (db) Get(id string) int {
	return 0
}

func (db) Set(id string, v int) {}
