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
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

//-----------------------------------------------------
// univariate polynomials

// Enum to tell in which basis a polynomial is represented.
type Basis int64

const (
	Canonical Basis = iota
	Lagrange
	LagrangeCoset
)

// Enum to tell if a polynomial is in bit reverse form or
// in the regular form.
type Layout int64

const (
	Regular Layout = iota
	BitReverse
)

// Enum to tell if the polynomial can be modified.
// If the polynomial can not be modified, then whenever
// a function has to do a transformation on it (FFT, bitReverse, etc)
// then a new vector is allocated.
type Status int64

const (
	Locked Status = iota
	Unlocked
)

// Form describes the form of a polynomial.
type Form struct {
	Basis  Basis
	Layout Layout
	Status Status
}

// Polynomial represents a polynomial, the vector of coefficients
// along with the basis and the layout.
type Polynomial struct {
	Coefficients []fr.Element
	Info         Form
}

//-----------------------------------------------------
// multivariate polynomials

// errors related to the polynomials.
var ErrInconsistantNumberOfVariable = errors.New("the number of variables is not consistant")

// monomial represents a monomial encoded as
// coeff*X₁^{i₁}*..*X_n^{i_n} if exponents = [i₁,..iₙ]
type monomial struct {
	coeff     fr.Element
	exponents []int
}

// it is supposed that the number of variables matches
func (m monomial) evaluate(x []fr.Element) fr.Element {

	var res, tmp fr.Element

	nbVars := len(x)
	res.SetOne()
	for i := 0; i < nbVars; i++ {
		if m.exponents[i] <= 5 {
			tmp = smallExp(x[i], m.exponents[i])
			res.Mul(&res, &tmp)
			continue
		}
		bi := big.NewInt(int64(i))
		tmp.Exp(x[i], bi)
		res.Mul(&res, &tmp)
	}
	res.Mul(&res, &m.coeff)

	return res

}

// reprensents a multivariate polynomial as a list of monomial,
// the multivariate polynomial being the sum of the monomials.
type MultivariatePolynomial []monomial

// degree returns the total degree
func (m MultivariatePolynomial) Degree() uint64 {
	r := 0
	for i := 0; i < len(m); i++ {
		t := 0
		for j := 0; j < len(m[i].exponents); j++ {
			t += m[i].exponents[j]
		}
		if t > r {
			r = t
		}
	}
	return uint64(r)
}

// AddMonomial adds a monomial to m. If m is empty, the monomial is
// added no matter what. But if m is already populated, an error is
// returned if len(e)\neq size of the previous list of exponents. This
// ensure that the number of variables is given by the size of any of
// the slices of exponent in any monomial.
func (m *MultivariatePolynomial) AddMonomial(c fr.Element, e []int) error {

	// if m is empty, we add the first monomial.
	if len(*m) == 0 {
		r := monomial{c, e}
		*m = append(*m, r)
		return nil
	}

	// at this stage all of exponennt in m are supposed to be of
	// the same size.
	if len((*m)[0].exponents) != len(e) {
		return ErrInconsistantNumberOfVariable
	}
	r := monomial{c, e}
	*m = append(*m, r)
	return nil

}

// evaluate a multivariate polynomial in x
// /!\ It is assumed that the multivariate polynomial has been
// built correctly, that is the sizes of the slices in exponents
// are the same /!\
func (m MultivariatePolynomial) Evaluate(x []fr.Element) fr.Element {

	var res fr.Element

	for i := 0; i < len(m); i++ {
		tmp := m[i].evaluate(x)
		res.Add(&res, &tmp)
	}
	return res
}
