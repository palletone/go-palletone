package asset

import (
	"testing"
	"fmt"
)

func TestNewAsset(t *testing.T) {
	uuid := NewAsset()
	fmt.Println("new uuid =", uuid.String())
}
