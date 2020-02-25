package ptnapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt_UnmarshalJSON(t *testing.T) {
	str := "123"
	i := &Int{}
	i.UnmarshalJSON([]byte(str))
	assert.EqualValues(t, 123, i.Uint32())
	str2 := "\"123\""
	i.UnmarshalJSON([]byte(str2))
	assert.EqualValues(t, 123, i.Uint32())
	str3 := ""
	i.UnmarshalJSON([]byte(str3))
	assert.EqualValues(t, 0, i.Uint32())
	i.UnmarshalJSON(nil)
	assert.EqualValues(t, 0, i.Uint32())
}
