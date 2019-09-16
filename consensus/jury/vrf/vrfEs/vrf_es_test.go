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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
package vrfEs

import (
    "testing"
    "crypto/ecdsa"
    "github.com/btcsuite/btcd/btcec"
    "crypto/rand"
)

/*
func TestH1(t *testing.T) {
    for i := 0; i < 10; i++ {
        m := make([]byte, 100)
        if _, err := rand.Read(m); err != nil {
            t.Fatalf("Failed generating random message: %v", err)
        }
        x, y := H1(m)
        if x == nil {
            t.Errorf("H1(%v)=%v, want curve point", m, x)
        }
        if got := curve.IsOnCurve(x, y); !got {
            t.Errorf("H1(%v)=[%v, %v], is not on curve", m, x, y)
        }
    }
}

func TestH2(t *testing.T) {
    l := 32
    for i := 0; i < 10000; i++ {
        m := make([]byte, 100)
        if _, err := rand.Read(m); err != nil {
            t.Fatalf("Failed generating random message: %v", err)
        }
        x := H2(m)
        if got := len(x.Bytes()); got < 1 || got > l {
            t.Errorf("len(h2(%v)) = %v, want: 1 <= %v <= %v", m, got, got, l)
        }
    }
}

func TestVRF(t *testing.T) {
    k, pk := GenerateKey()

    m1 := []byte("data1")
    m2 := []byte("data2")
    m3 := []byte("data2")
    index1, proof1 := k.Evaluate(m1)
    index2, proof2 := k.Evaluate(m2)
    index3, proof3 := k.Evaluate(m3)

    t.Logf("index1[%v]  index2[%v], index3[%v]", index1[0:4], index2[0:4], index3[0:4])
    t.Logf("proof1[%v]  proof2[%v], proof3[%v]", proof1[0:4], proof2[0:4], proof3[0:4])

    for i, tc := range []struct {
        m     []byte
        index [32]byte
        proof []byte
        err   error
    }{
        {m1, index1, proof1, nil},
        {m2, index2, proof2, nil},
        {m3, index3, proof3, nil},
        {m3, index3, proof2, nil},
        {m3, index3, proof1, ErrInvalidVRF},
    } {
        index, err := pk.ProofToHash(tc.m, tc.proof)
        if got, want := err, tc.err; got != want {
            t.Errorf("ProofToHash(%s, %x): %v, want %v", tc.m, tc.proof, got, want)
        }
        if err != nil {
            t.Logf("index[%d]", i)
            continue
        }
        if got, want := index, tc.index; got != want {
            t.Errorf("ProofToInex(%s, %x): %x, want %x", tc.m, tc.proof, got, want)
        }
    }
}

func TestRightTruncateProof(t *testing.T) {
    k, pk := GenerateKey()

    data := []byte("data")
    _, proof := k.Evaluate(data)
    proofLen := len(proof)
    for i := 0; i < proofLen; i++ {
        proof = proof[:len(proof)-1]
        if _, err := pk.ProofToHash(data, proof); err == nil {
            t.Errorf("Verify unexpectedly succeeded after truncating %v bytes from the end of proof", i)
        }
    }
}

func TestLeftTruncateProof(t *testing.T) {
    k, pk := GenerateKey()

    data := []byte("data")
    _, proof := k.Evaluate(data)
    proofLen := len(proof)
    for i := 0; i < proofLen; i++ {
        proof = proof[1:]
        if _, err := pk.ProofToHash(data, proof); err == nil {
            t.Errorf("Verify unexpectedly succeeded after truncating %v bytes from the beginning of proof", i)
        }
    }
}

func TestBitFlip(t *testing.T) {
    k, pk := GenerateKey()

    data := []byte("data")
    _, proof := k.Evaluate(data)
    for i := 0; i < len(proof)*8; i++ {
        // Flip bit in position i.
        if _, err := pk.ProofToHash(data, flipBit(proof, i)); err == nil {
            t.Errorf("Verify unexpectedly succeeded after flipping bit %v of vrf", i)
        }
    }
}

func flipBit(a []byte, pos int) []byte {
    index := int(math.Floor(float64(pos) / 8))
    b := a[index]
    b ^= (1 << uint(math.Mod(float64(pos), 8.0)))

    var buf bytes.Buffer
    buf.Write(a[:index])
    buf.Write([]byte{b})
    buf.Write(a[index+1:])
    return buf.Bytes()
}
*/
func TestVRFForS256(t *testing.T) {
    key, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
    if err != nil {
        t.Error(err)
        return
    }
    k, err := NewVRFSigner(key)
    if err != nil {
        t.Error(err)
        return
    }
    pk, err := NewVRFVerifier(&key.PublicKey)
    if err != nil {
        t.Error(err)
        return
    }
    msg := []byte("data1")
    index, proof := k.Evaluate(msg)
    _index, _proof := k.Evaluate(msg)
    index1, err1 := pk.ProofToHash(msg, proof)
    index2, err2 := pk.ProofToHash(msg, _proof)

    if err1 != nil {
        t.Error(err1)
    }
    if err2 != nil {
        t.Error(err2)
    }

    if index1 != index {
        t.Error("index not equal")
    }
    t.Log(index, _index)
    t.Log(index1, index2)
    t.Log(proof, _proof)
}