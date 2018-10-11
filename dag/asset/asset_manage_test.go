package asset

import (
	"testing"
)

func TestNewAsset(t *testing.T) {
	uuid := NewAsset()
	t.Log("new uuid =", uuid.String())
}
