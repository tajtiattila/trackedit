package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// modulePath() returns the path to the source when main was built
func modulePath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't get module path")
	}
	return filepath.Dir(file)
}

func stripGoSrcPath(path string) string {
	return filepath.ToSlash(stripGoPathSub(path, "src"))
}

func stripGoPathSub(path, sub string) string {
	gp := os.Getenv("GOPATH")
	if gp == "" {
		return path
	}

	gp = filepath.Join(gp, sub)
	fp := filepath.Clean(path)
	rel, err := filepath.Rel(gp, fp)
	logErr(err)
	if err == nil {
		return rel
	}
	return path
}

func verify(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
