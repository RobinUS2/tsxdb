package tools

import (
	"github.com/OneOfOne/xxhash"
	"github.com/satori/go.uuid"
)

func RandomInsecureIdentifier() uint64 {
	u := uuid.Must(uuid.NewV4()).Bytes()
	h := xxhash.New64()
	_, err := h.Write(u)
	if err != nil {
		panic(err)
	}
	return h.Sum64()
}
