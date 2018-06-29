package files

import (
	"testing"
)

func TestMakeDirAndFileThenRemove(t *testing.T) {
	paths := []string{
		"log2/log2_1/log2_2/log.log",
		"log1/log1_1/log.log",
		"test.log",
	}
	for _, p := range paths {
		MakeDirAndFile(p)
		if !IsExist(p) {
			t.Error("MakeDirAndFile error")
		}
	}
	for _, p := range paths {
		RemoveFileAndEmptyFolder(p)
		if IsExist(p) {
			t.Error("RemoveFileAndEmptyFolder error")
		}
	}

}
