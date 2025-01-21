package testutils

import (
	"os"
	"path/filepath"
	"runtime"
)

var RecognitionTestFile []byte

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current file path")
	}
	
	testFilePath := filepath.Join(filepath.Dir(filename), "nodes.json")
	f, err := os.ReadFile(testFilePath)
	if err != nil {
		panic(err)
	}

	RecognitionTestFile = f
}
