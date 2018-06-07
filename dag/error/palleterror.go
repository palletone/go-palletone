package error

import (
	"errors"
)

var (
	ErrGraphHasCycle = errors.New("dag: Graph has a cycle")
	ErrSetEmpty      = errors.New("dag: Set is empty")
	ErrNotFound      = errors.New("dag: Not found")
)
