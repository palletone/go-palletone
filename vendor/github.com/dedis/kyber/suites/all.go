package suites

import (
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

func init() {
	register(edwards25519.NewBlakeSHA256Ed25519())
}
