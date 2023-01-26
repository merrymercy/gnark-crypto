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

package schnorr

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

func TestSchnorr(t *testing.T) {

	t.Parallel()
	parameters := gopter.DefaultTestParameters()
	properties := gopter.NewProperties(parameters)

	properties.Property("[SECP256K1] test the signing and verification", prop.ForAll(
		func() bool {

			privKey, _ := GenerateKey(rand.Reader)
			publicKey := privKey.PublicKey

			msg := []byte("testing Schnorr")
			hFunc := sha256.New()
			sig, _ := privKey.Sign(msg, hFunc)
			flag, _ := publicKey.Verify(sig, msg, hFunc)

			return flag
		},
	))

	properties.Property("[SECP256K1] test the signing and verification (pre-hashed)", prop.ForAll(
		func() bool {

			privKey, _ := GenerateKey(rand.Reader)
			publicKey := privKey.PublicKey

			msg := []byte("testing Schnorr")
			sig, _ := privKey.Sign(msg, nil)
			flag, _ := publicKey.Verify(sig, msg, nil)

			return flag
		},
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkSignSchnorr(b *testing.B) {

	privKey, _ := GenerateKey(rand.Reader)

	msg := []byte("benchmarking Schnorr sign()")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		privKey.Sign(msg, nil)
	}
}

func BenchmarkVerifySchnorr(b *testing.B) {

	privKey, _ := GenerateKey(rand.Reader)
	msg := []byte("benchmarking Schnorr sign()")
	sig, _ := privKey.Sign(msg, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		privKey.PublicKey.Verify(sig, msg, nil)
	}
}
