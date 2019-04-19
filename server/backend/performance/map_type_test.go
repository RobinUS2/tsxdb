package performance_test

import (
	"math/rand"
	"testing"
)

//go version go1.12.1 darwin/amd64
//goos: darwin
//goarch: amd64
//BenchmarkMapNativeKey-8       	 5000000	       352 ns/op
//BenchmarkMapTypedKey-8        	 5000000	       274 ns/op
//BenchmarkMapTypedValue-8      	 5000000	       232 ns/op
//BenchmarkMapTypedKeyValue-8   	 5000000	       248 ns/op
//PASS
// Conclusion: doesn't really matter, use types wherever possible for type safety

type WrappedKey int
type WrappedValue float64

func BenchmarkMapNativeKey(b *testing.B) {
	var m = make(map[int]float64)
	for n := 0; n < b.N; n++ {
		m[n] = rand.Float64()
	}
}

func BenchmarkMapTypedKey(b *testing.B) {
	var m = make(map[WrappedKey]float64)
	for n := 0; n < b.N; n++ {
		m[WrappedKey(n)] = rand.Float64()
	}
}

func BenchmarkMapTypedValue(b *testing.B) {
	var m = make(map[int]WrappedValue)
	for n := 0; n < b.N; n++ {
		m[n] = WrappedValue(rand.Float64())
	}
}
func BenchmarkMapTypedKeyValue(b *testing.B) {
	var m = make(map[WrappedKey]WrappedValue)
	for n := 0; n < b.N; n++ {
		m[WrappedKey(n)] = WrappedValue(rand.Float64())
	}
}
