/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package mediatorplugin

import (
	"fmt"
	"testing"

	"github.com/palletone/go-palletone/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/share/dkg/pedersen"
	"go.dedis.ch/kyber/v3/share/vss/pedersen"
	"go.dedis.ch/kyber/v3/sign/bls"
	"go.dedis.ch/kyber/v3/sign/tbls"
)

var (
	suite = bn256.NewSuiteG2()

	nbParticipants = 21
	ntThreshold    = nbParticipants*2/3 + 1

	partSec  []kyber.Scalar
	partPubs []kyber.Point

	dkgs []*dkg.DistKeyGenerator
)

func init() {
	partPubs = make([]kyber.Point, nbParticipants)
	partSec = make([]kyber.Scalar, nbParticipants)

	for i := 0; i < nbParticipants; i++ {
		sec, pub := genPair()
		partSec[i] = sec
		partPubs[i] = pub

		fmt.Printf("Key[%v] priv: %v, pub %v\n", i, sec.String(), pub.String())
	}
}

func genPair() (kyber.Scalar, kyber.Point) {
	sc := suite.Scalar().Pick(suite.RandomStream())
	return sc, suite.Point().Mul(sc, nil)
}

func dkgGen(t *testing.T) []*dkg.DistKeyGenerator {
	dkgs := make([]*dkg.DistKeyGenerator, nbParticipants)
	for i := 0; i < nbParticipants; i++ {
		//dkg, err := dkg.NewDistKeyGeneratorWithoutSecret(suite, partSec[i], partPubs, ntThreshold)
		dkg, err := dkg.NewDistKeyGenerator(suite, partSec[i], partPubs, ntThreshold)
		require.Nil(t, err)
		dkgs[i] = dkg
	}
	return dkgs
}

func fullExchange(t *testing.T) {
	dkgs = dkgGen(t)

	// full secret sharing exchange
	// 1. broadcast deals
	resps := make([]*dkg.Response, 0, nbParticipants*nbParticipants)
	for _, dkg := range dkgs {
		deals, err := dkg.Deals()
		//fmt.Printf("the number of deals is: %v\n", len(deals))
		require.Nil(t, err)
		for i, d := range deals {
			resp, err := dkgs[i].ProcessDeal(d)
			require.Nil(t, err)
			require.Equal(t, vss.StatusApproval, resp.Response.Status)
			resps = append(resps, resp)
		}
	}

	// 2. Broadcast responses
	for _, resp := range resps {
		for i, dkg := range dkgs {
			// ignore all messages from ourselves
			if resp.Response.Index == uint32(i) {
				continue
			}

			j, err := dkg.ProcessResponse(resp)
			require.Nil(t, err)
			require.Nil(t, j)
		}
	}
}

func TestTBLS(t *testing.T) {
	fullExchange(t)

	msg := []byte("Hello DKG, VSS, TBLS and BLS!")

	sigShares := make([][]byte, 0)
	for i, d := range dkgs {
		if i == ntThreshold {
			break
		}

		dks, err := d.DistKeyShare()
		assert.Nil(t, err)
		sig, err := tbls.Sign(suite, dks.PriShare(), msg)
		require.Nil(t, err)
		sigShares = append(sigShares, sig)
	}

	dkg := dkgs[0]
	dks, err := dkg.DistKeyShare()
	require.Nil(t, err)

	pubPoly := share.NewPubPoly(suite, suite.Point().Base(), dks.Commitments())
	sig, err := tbls.Recover(suite, pubPoly, msg, sigShares, ntThreshold, nbParticipants)
	require.Nil(t, err)

	fmt.Printf("threshold sign: %v", hexutil.Encode(sig))

	err = bls.Verify(suite, pubPoly.Commit(), msg, sig)
	require.Nil(t, err)

	err = bls.Verify(suite, dks.Public(), msg, sig)
	assert.Nil(t, err)

	require.Equal(t, pubPoly.Commit(), dks.Public())

	dks2, err := dkgs[1].DistKeyShare()
	assert.Nil(t, err)

	err = bls.Verify(suite, dks2.Public(), msg, sig)
	assert.Nil(t, err)

	//require.Equal(t, dks.Public(), dks2.Public())
	require.NotEqual(t, dks.Public(), dks2.Public())

	maybepub2 := dks.Commitments()[1]
	err = bls.Verify(suite, maybepub2, msg, sig)
	assert.NotNil(t, err)

	require.NotEqual(t, maybepub2, dks2.Public())
}
