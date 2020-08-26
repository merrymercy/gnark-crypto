// Copyright 2020 ConsenSys AG
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

// Code generated by gurvy DO NOT EDIT

package bw761

import (
	"math"
	"runtime"

	"github.com/consensys/gurvy/bw761/fr"
)

// MultiExp implements section 4 of https://eprint.iacr.org/2012/549.pdf
func (p *G1Jac) MultiExp(points []G1Affine, scalars []fr.Element) *G1Jac {
	// note:
	// each of the msmCX method is the same, except for the c constant it declares
	// duplicating (through template generation) these methods allows to declare the buckets on the stack
	// the choice of c needs to be improved:
	// there is a theoritical value that gives optimal asymptotics
	// but in practice, other factors come into play, including:
	// * if c doesn't divide 64, the word size, then we're bound to select bits over 2 words of our scalars, instead of 1
	// * number of CPUs
	// * cache friendliness (which depends on the host, G1 or G2... )
	//	--> for example, on BN256, a G1 point fits into one cache line of 64bytes, but a G2 point don't.

	// for each msmCX
	// step 1
	// we compute, for each scalars over c-bit wide windows, nbChunk digits
	// if the digit is larger than 2^{c-1}, then, we borrow 2^c from the next window and substract
	// 2^{c} to the current digit, making it negative.
	// negative digits will be processed in the next step as adding -G into the bucket instead of G
	// (computing -G is cheap, and this saves us half of the buckets)
	// step 2
	// buckets are declared on the stack
	// notice that we have 2^{c-1} buckets instead of 2^{c} (see step1)
	// we use jacobian extended formulas here as they are faster than mixed addition
	// msmProcessChunk places points into buckets base on their selector and return the weighted bucket sum in given channel
	// step 3
	// reduce the buckets weigthed sums into our result (msmReduceChunk)

	// approximate cost (in group operations)
	// cost = bits/c * (nbPoints + 2^{c-1})
	// this needs to be verified empirically.
	// for example, on a MBP 2016, for G2 MultiExp > 8M points, hand picking c gives better results
	implementedCs := []int{4, 8, 16}

	nbPoints := len(points)
	min := math.MaxFloat64
	bestC := 0
	for _, c := range implementedCs {
		cc := fr.Limbs * 64 * (nbPoints + (1 << (c - 1)))
		cost := float64(cc) / float64(c)
		if cost < min {
			min = cost
			bestC = c
		}
	}

	// semaphore to limit number of cpus iterating through points and scalrs at the same time
	numCpus := runtime.NumCPU()
	chCpus := make(chan struct{}, numCpus)
	for i := 0; i < numCpus; i++ {
		chCpus <- struct{}{}
	}

	switch bestC {

	case 4:
		return p.msmC4(points, scalars, chCpus)

	case 8:
		return p.msmC8(points, scalars, chCpus)

	case 16:
		return p.msmC16(points, scalars, chCpus)

	default:
		panic("unimplemented")
	}
}

// msmReduceChunkG1 reduces the weighted sum of the buckets into the result of the multiExp
func msmReduceChunkG1(p *G1Jac, c int, chChunks []chan G1Jac) *G1Jac {
	totalj := <-chChunks[len(chChunks)-1]
	p.Set(&totalj)
	for j := len(chChunks) - 2; j >= 0; j-- {
		for l := 0; l < c; l++ {
			p.DoubleAssign()
		}
		totalj := <-chChunks[j]
		p.AddAssign(&totalj)
	}
	return p
}

func msmProcessChunkG1(chunk uint64,
	chRes chan<- G1Jac,
	chCpus chan struct{},
	buckets []g1JacExtended,
	c uint64,
	points []G1Affine,
	scalars []fr.Element) {

	<-chCpus // wait and decrement avaiable CPUs on the semaphore

	mask := uint64((1 << c) - 1) // low c bits are 1
	msbWindow := uint64(1 << (c - 1))

	for i := 0; i < len(buckets); i++ {
		buckets[i].SetInfinity()
	}

	jc := uint64(chunk * c)
	s := selector{}
	s.index = jc / 64
	s.shift = jc - (s.index * 64)
	s.mask = mask << s.shift
	s.multiWordSelect = (64%c) != 0 && s.shift > (64-c) && s.index < (fr.Limbs-1)
	if s.multiWordSelect {
		nbBitsHigh := s.shift - uint64(64-c)
		s.maskHigh = (1 << nbBitsHigh) - 1
		s.shiftHigh = (c - nbBitsHigh)
	}

	// for each scalars, get the digit corresponding to the chunk we're processing.
	for i := 0; i < len(scalars); i++ {
		bits := (scalars[i][s.index] & s.mask) >> s.shift
		if s.multiWordSelect {
			bits += (scalars[i][s.index+1] & s.maskHigh) << s.shiftHigh
		}

		if bits == 0 {
			continue
		}

		// if msbWindow bit is set, we need to substract
		if bits&msbWindow == 0 {
			// add
			buckets[bits-1].mAdd(&points[i])
		} else {
			// sub
			buckets[bits & ^msbWindow].mSub(&points[i])
		}
	}

	// reduce buckets into total
	// total =  bucket[0] + 2*bucket[1] + 3*bucket[2] ... + n*bucket[n-1]

	var runningSum, tj, total G1Jac
	runningSum.Set(&g1Infinity)
	total.Set(&g1Infinity)
	for k := len(buckets) - 1; k >= 0; k-- {
		if !buckets[k].ZZ.IsZero() {
			runningSum.AddAssign(tj.unsafeFromJacExtended(&buckets[k]))
		}
		total.AddAssign(&runningSum)
	}

	chRes <- total
	close(chRes)
	chCpus <- struct{}{} // increment avaiable CPUs into the semaphore
}

func (p *G1Jac) msmC4(points []G1Affine, scalars []fr.Element, chCpus chan struct{}) *G1Jac {
	const c = 4                          // scalars partitioned into c-bit radixes
	const nbChunks = (fr.Limbs * 64 / c) // number of c-bit radixes in a scalar

	// partition the scalars
	// note: we do that before the actual chunk processing, as for each c-bit window (starting from LSW)
	// if it's larger than 2^{c-1}, we have a carry we need to propagate up to the higher window
	pScalars := PartitionScalars(scalars, c)

	// for each chunk, spawn a go routine that'll loop through all the scalars
	var chChunks [nbChunks]chan G1Jac
	for chunk := nbChunks - 1; chunk >= 0; chunk-- {
		chChunks[chunk] = make(chan G1Jac, 1)
		go func(j uint64) {
			var buckets [1 << (c - 1)]g1JacExtended
			msmProcessChunkG1(j, chChunks[j], chCpus, buckets[:], c, points, pScalars)
		}(uint64(chunk))
	}

	return msmReduceChunkG1(p, c, chChunks[:])
}

func (p *G1Jac) msmC8(points []G1Affine, scalars []fr.Element, chCpus chan struct{}) *G1Jac {
	const c = 8                          // scalars partitioned into c-bit radixes
	const nbChunks = (fr.Limbs * 64 / c) // number of c-bit radixes in a scalar

	// partition the scalars
	// note: we do that before the actual chunk processing, as for each c-bit window (starting from LSW)
	// if it's larger than 2^{c-1}, we have a carry we need to propagate up to the higher window
	pScalars := PartitionScalars(scalars, c)

	// for each chunk, spawn a go routine that'll loop through all the scalars
	var chChunks [nbChunks]chan G1Jac
	for chunk := nbChunks - 1; chunk >= 0; chunk-- {
		chChunks[chunk] = make(chan G1Jac, 1)
		go func(j uint64) {
			var buckets [1 << (c - 1)]g1JacExtended
			msmProcessChunkG1(j, chChunks[j], chCpus, buckets[:], c, points, pScalars)
		}(uint64(chunk))
	}

	return msmReduceChunkG1(p, c, chChunks[:])
}

func (p *G1Jac) msmC16(points []G1Affine, scalars []fr.Element, chCpus chan struct{}) *G1Jac {
	const c = 16                         // scalars partitioned into c-bit radixes
	const nbChunks = (fr.Limbs * 64 / c) // number of c-bit radixes in a scalar

	// partition the scalars
	// note: we do that before the actual chunk processing, as for each c-bit window (starting from LSW)
	// if it's larger than 2^{c-1}, we have a carry we need to propagate up to the higher window
	pScalars := PartitionScalars(scalars, c)

	// for each chunk, spawn a go routine that'll loop through all the scalars
	var chChunks [nbChunks]chan G1Jac
	for chunk := nbChunks - 1; chunk >= 0; chunk-- {
		chChunks[chunk] = make(chan G1Jac, 1)
		go func(j uint64) {
			var buckets [1 << (c - 1)]g1JacExtended
			msmProcessChunkG1(j, chChunks[j], chCpus, buckets[:], c, points, pScalars)
		}(uint64(chunk))
	}

	return msmReduceChunkG1(p, c, chChunks[:])
}
