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

package iop

import (
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/fft"
)

func printVector(v []fr.Element) {
	fmt.Printf("[")
	for i := 0; i < len(v); i++ {
		fmt.Printf("Fr(%s), ", v[i].String())
	}
	fmt.Printf("]\n")
}

func printPolynomials(p []*Polynomial) {
	fmt.Printf("[\n")
	for i := 0; i < len(p); i++ {
		printVector(p[i].Coefficients)
		fmt.Printf(",\n")
	}
	fmt.Printf("]\n")
}

func printLayout(f Form) {

	if f.Basis == Canonical {
		fmt.Printf("CANONICAL")
	} else if f.Basis == LagrangeCoset {
		fmt.Printf("LAGRANGE_COSET")
	} else {
		fmt.Printf("LAGRANGE")
	}
	fmt.Println("")

	if f.Layout == Regular {
		fmt.Printf("REGULAR")
	} else {
		fmt.Printf("BIT REVERSED")
	}
	fmt.Println("")

	if f.Status == Locked {
		fmt.Printf("LOCKED")
	} else {
		fmt.Printf("UNLOCKED")
	}
	fmt.Println("")
}

// return a copy of p
// return a copy of p
func copyPoly(p Polynomial) Polynomial {
	size := len(p.Coefficients)
	var r Polynomial
	r.Coefficients = make([]fr.Element, size)
	copy(r.Coefficients, p.Coefficients)
	r.Form = p.Form
	return r
}

// return an ID corresponding to the polynomial extra data
func getShapeID(p Polynomial) int {
	return int(p.Basis)*4 + int(p.Layout)*2 + int(p.Status)
}

//----------------------------------------------------
// toLagrange

// the numeration corresponds to the following formatting:
// num = int(p.Basis)*4 + int(p.Layout)*2 + int(p.Status)

// CANONICAL REGULAR LOCKED
func (p *Polynomial) toLagrange0(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Lagrange
	_p.Layout = BitReverse
	_p.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

// CANONICAL REGULAR UNLOCKED
func (p *Polynomial) toLagrange1(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = BitReverse
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

// CANONICAL BITREVERSE LOCKED
func (p *Polynomial) toLagrange2(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Lagrange
	_p.Layout = Regular
	_p.Status = Unlocked
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

// CANONICAL BITREVERSE UNLOCKED
func (p *Polynomial) toLagrange3(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = Regular
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE REGULAR LOCKED
func (p *Polynomial) toLagrange4(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE REGULAR UNLOCKED
func (p *Polynomial) toLagrange5(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE BITREVERSE LOCKED
func (p *Polynomial) toLagrange6(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE BITREVERSE UNLOCKED
func (p *Polynomial) toLagrange7(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE_COSET REGULAR LOCKED
func (p *Polynomial) toLagrange8(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Lagrange
	_p.Layout = Regular
	_p.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	d.FFT(_p.Coefficients, fft.DIT)
	return &_p
}

// LAGRANGE_COSET REGULAR UNLOCKED
func (p *Polynomial) toLagrange9(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	d.FFTInverse(p.Coefficients, fft.DIF, true)
	d.FFT(p.Coefficients, fft.DIT)
	return p
}

// LAGRANGE_COSET BITREVERSE LOCKED
func (p *Polynomial) toLagrange10(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Lagrange
	_p.Layout = BitReverse
	_p.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIT, true)
	d.FFT(_p.Coefficients, fft.DIF)
	return &_p
}

// LAGRANGE_COSET BITREVERSE UNLOCKED
func (p *Polynomial) toLagrange11(d *fft.Domain) *Polynomial {
	p.Basis = Lagrange
	p.Layout = BitReverse
	p.Status = Unlocked
	d.FFTInverse(p.Coefficients, fft.DIT, true)
	d.FFT(p.Coefficients, fft.DIF)
	return p
}

// toLagrange changes or returns a copy of p (according to its
// status, Locked or Unlocked), or modifies p to put it in Lagrange
// basis. The result is not bit reversed.
func (p *Polynomial) ToLagrange(d *fft.Domain) *Polynomial {
	id := getShapeID(*p)
	switch id {
	case 0:
		return p.toLagrange0(d)
	case 1:
		return p.toLagrange1(d)
	case 2:
		return p.toLagrange2(d)
	case 3:
		return p.toLagrange3(d)
	case 4:
		return p.toLagrange4(d)
	case 5:
		return p.toLagrange5(d)
	case 6:
		return p.toLagrange6(d)
	case 7:
		return p.toLagrange7(d)
	case 8:
		return p.toLagrange8(d)
	case 9:
		return p.toLagrange9(d)
	case 10:
		return p.toLagrange10(d)
	case 11:
		return p.toLagrange11(d)
	default:
		panic("unknown ID")
	}
}

//----------------------------------------------------
// toCanonical

// CANONICAL REGULAR LOCKED
func (p *Polynomial) toCanonical0(d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL REGULAR UNLOCKED
func (p *Polynomial) toCanonical1(d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL BITREVERSE LOCKED
func (p *Polynomial) toCanonical2(d *fft.Domain) *Polynomial {
	return p
}

// CANONICAL BITREVERSE UNLOCKED
func (p *Polynomial) toCanonical3(d *fft.Domain) *Polynomial {
	return p
}

// LAGRANGE REGULAR LOCKED
func (p *Polynomial) toCanonical4(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Canonical
	_p.Layout = BitReverse
	_p.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF)
	return &_p
}

// LAGRANGE REGULAR UNLOCKED
func (p *Polynomial) toCanonical5(d *fft.Domain) *Polynomial {
	d.FFTInverse(p.Coefficients, fft.DIF)
	p.Basis = Canonical
	p.Layout = BitReverse
	return p
}

// LAGRANGE BITREVERSE LOCKED
func (p *Polynomial) toCanonical6(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Canonical
	_p.Layout = Regular
	_p.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIT)
	return &_p
}

// LAGRANGE BITREVERSE UNLOCKED
func (p *Polynomial) toCanonical7(d *fft.Domain) *Polynomial {
	d.FFTInverse(p.Coefficients, fft.DIT)
	p.Basis = Canonical
	p.Layout = Regular
	return p
}

// LAGRANGE_COSET REGULAR LOCKED
func (p *Polynomial) toCanonical8(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Canonical
	_p.Layout = BitReverse
	_p.Status = Unlocked
	d.FFTInverse(_p.Coefficients, fft.DIF, true)
	return &_p
}

// LAGRANGE_COSET REGULAR UNLOCKED
func (p *Polynomial) toCanonical9(d *fft.Domain) *Polynomial {
	p.Basis = Canonical
	p.Layout = BitReverse
	d.FFTInverse(p.Coefficients, fft.DIF, true)
	return p
}

// LAGRANGE_COSET BITREVERSE LOCKED
func (p *Polynomial) toCanonical10(d *fft.Domain) *Polynomial {
	_p := copyPoly(*p)
	_p.Basis = Canonical
	_p.Layout = Regular
	d.FFTInverse(_p.Coefficients, fft.DIT, true)
	return &_p
}

// LAGRANGE_COSET BITREVERSE UNLOCKED
func (p *Polynomial) toCanonical11(d *fft.Domain) *Polynomial {
	p.Basis = Canonical
	p.Layout = Regular
	d.FFTInverse(p.Coefficients, fft.DIT, true)
	return p
}

// toCanonical changes or returns a copy of p (according to its
// status, Locked or Unlocked), or modifies p to put it in Lagrange
// basis. The result is not bit reversed.
func (p *Polynomial) ToCanonical(d *fft.Domain) *Polynomial {
	id := getShapeID(*p)
	switch id {
	case 0:
		return p.toCanonical0(d)
	case 1:
		return p.toCanonical1(d)
	case 2:
		return p.toCanonical2(d)
	case 3:
		return p.toCanonical3(d)
	case 4:
		return p.toCanonical4(d)
	case 5:
		return p.toCanonical5(d)
	case 6:
		return p.toCanonical6(d)
	case 7:
		return p.toCanonical7(d)
	case 8:
		return p.toCanonical8(d)
	case 9:
		return p.toCanonical9(d)
	case 10:
		return p.toCanonical10(d)
	case 11:
		return p.toCanonical11(d)
	default:
		panic("unknown ID")
	}
}

//----------------------------------------------------
// exp functions until 5

func exp0(x fr.Element) fr.Element {
	var res fr.Element
	res.SetOne()
	return res
}

func exp1(x fr.Element) fr.Element {
	return x
}

func exp2(x fr.Element) fr.Element {
	return *x.Square(&x)
}

func exp3(x fr.Element) fr.Element {
	var res fr.Element
	res.Square(&x).Mul(&res, &x)
	return res
}

func exp4(x fr.Element) fr.Element {
	x.Square(&x).Square(&x)
	return x
}

func exp5(x fr.Element) fr.Element {
	var res fr.Element
	res.Square(&x).Square(&res).Mul(&res, &x)
	return res
}

// doesn't return any errors, it is a private method, that
// is assumed to be called with correct arguments.
func smallExp(x fr.Element, n int) fr.Element {
	if n == 0 {
		return exp0(x)
	}
	if n == 1 {
		return exp1(x)
	}
	if n == 2 {
		return exp2(x)
	}
	if n == 3 {
		return exp3(x)
	}
	if n == 4 {
		return exp4(x)
	}
	if n == 5 {
		return exp5(x)
	}
	return fr.Element{}
}
