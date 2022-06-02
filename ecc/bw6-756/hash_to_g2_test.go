// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package bw6756

import (
	"github.com/consensys/gnark-crypto/ecc/bw6-756/fp"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"math/rand"
	"testing"
)

func TestG2SqrtRatio(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	gen := GenFp()

	properties.Property("G2SqrtRatio must square back to the right value", prop.ForAll(
		func(u fp.Element, v fp.Element) bool {

			var seen fp.Element
			qr := g2SqrtRatio(&seen, &u, &v) == 0

			seen.
				Square(&seen).
				Mul(&seen, &v)

			var ref fp.Element
			if qr {
				ref = u
			} else {
				g2MulByZ(&ref, &u)
			}

			return seen.Equal(&ref)
		}, gen, gen))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

//TODO: Crude. Do something clever in Jacobian
func isOnEPrimeG2(p G2Affine) bool {

	var A, B fp.Element

	// TODO: Value already in Mont form, set string without mont conversion

	A.SetString(
		"366214298489676225608748285593816060502004793143294716400170581597694044098131723083800748871288574463946039491660810740595875521255828936150127543994076275496865575201302523356132387940696554359017695480324003519841793632648747",
	)

	B.SetString(
		"123700894234907950132354945578170275435433180409034092978696279417897820098231450798376842588105055260469163151822541495836570304917528780887970917933871659898856211667296627282615124577313799611865528162645151540457563642437575",
	)

	A.FromMont()
	B.FromMont()

	var LHS fp.Element
	LHS.
		Square(&p.Y).
		Sub(&LHS, &B)

	var RHS fp.Element
	RHS.
		Square(&p.X).
		Add(&RHS, &A).
		Mul(&RHS, &p.X)

	return LHS.Equal(&RHS)
}

func TestG2SSWU(t *testing.T) {
	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	if testing.Short() {
		parameters.MinSuccessfulTests = nbFuzzShort
	} else {
		parameters.MinSuccessfulTests = nbFuzz
	}

	properties := gopter.NewProperties(parameters)

	properties.Property("[G2] hash outputs must be in appropriate groups", prop.ForAll(
		func(a fp.Element) bool {

			g := mapToCurve2(&a)

			if !isOnEPrimeG2(g) {
				t.Log("SSWU output not on E' curve")
				return false
			}

			g2Isogeny(&g)

			if !g.IsOnCurve() {
				t.Log("Isogeny/SSWU output not on curve")
				return false
			}

			return true
		},
		GenFp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func g2TestMatchCoord(t *testing.T, coordName string, msg string, expectedStr string, seen *fp.Element) {
	var expected fp.Element

	expected.SetString(expectedStr)

	if !expected.Equal(seen) {
		t.Errorf("mismatch on \"%s\", %s:\n\texpected %s\n\tsaw      %s", msg, coordName, expected.String(), seen)
	}
}

func g2TestMatch(t *testing.T, c hashTestCase, seen *G2Affine) {
	g2TestMatchCoord(t, "x", c.msg, c.x, &seen.X)
	g2TestMatchCoord(t, "y", c.msg, c.y, &seen.Y)
}

func TestEncodeToG2(t *testing.T) {
	t.Parallel()
	for _, c := range g2EncodeToCurveSSWUVector.cases {
		seen, err := EncodeToG2([]byte(c.msg), g2EncodeToCurveSSWUVector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g2TestMatch(t, c, &seen)
	}
}

func TestHashToG2(t *testing.T) {
	t.Parallel()
	for _, c := range g2HashToCurveSSWUVector.cases {
		seen, err := HashToG2([]byte(c.msg), g2HashToCurveSSWUVector.dst)
		if err != nil {
			t.Fatal(err)
		}
		g2TestMatch(t, c, &seen)
	}
	t.Log(len(g2HashToCurveSSWUVector.cases), "cases verified")
}

func BenchmarkEncodeToG2(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := g2EncodeToCurveSSWUVector.dst
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		bytes[rand.Int()%size] = byte(rand.Int())

		if _, err := EncodeToG2(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

func BenchmarkHashToG2(b *testing.B) {
	const size = 54
	bytes := make([]byte, size)
	dst := g2HashToCurveSSWUVector.dst
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		bytes[rand.Int()%size] = byte(rand.Int())

		if _, err := HashToG2(bytes, dst); err != nil {
			b.Fail()
		}
	}
}

var g2HashToCurveSSWUVector hashTestVector
var g2EncodeToCurveSSWUVector hashTestVector
