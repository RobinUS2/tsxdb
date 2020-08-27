package performance_test

import (
	"fmt"
	hashicorp "github.com/hashicorp/golang-lru"
	"github.com/karlseguin/ccache/v2"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

//go version go1.14.2 darwin/amd64
//goos: darwin
//goarch: amd64
//pkg: github.com/RobinUS2/tsxdb/server/backend/performance
//BenchmarkHashicorp
//BenchmarkHashicorp-8   	  944810	      1181 ns/op
//BenchmarkCcache
//BenchmarkCcache-8      	  582338	      1923 ns/op
//PASS
// Conclusion: ccache 60% slower yet supports advanced options like TTL which we need, worth the trade off

func BenchmarkHashicorp(b *testing.B) {
	c, err := hashicorp.New(1000)
	if err != nil {
		b.Error(err)
	}
	for n := 0; n < b.N; n++ {
		key := strconv.FormatInt(rand.Int63(), 10)
		value := rand.Float32()
		c.Add(key, value)
		res, found := c.Get(key)
		if !found {
			b.Error()
		}
		if fmt.Sprintf("%f", res) != fmt.Sprintf("%f", value) {
			b.Error("wrong")
		}
	}
}

func BenchmarkCcache(b *testing.B) {
	c := ccache.New(ccache.Configure().MaxSize(1000))
	if c == nil {
		b.Error()
	}
	for n := 0; n < b.N; n++ {
		key := strconv.FormatInt(rand.Int63(), 10)
		value := rand.Float32()
		c.Set(key, value, 1*time.Hour)
		res := c.Get(key)
		if res == nil {
			b.Error()
		}
		if fmt.Sprintf("%f", res.Value()) != fmt.Sprintf("%f", value) {
			b.Error("wrong")
		}
	}
}
