package backend

import (
	"fmt"
	"testing"
)

func TestGetKeyScoreMember(t *testing.T) {
	// validate how we create buckets, keys and set members
	var b *RedisBackend
	var ctx = ContextWrite{
		Context{
			Series:    123,
			Namespace: 2,
		},
	}
	const ts = 1598261325123
	const value = 1.234
	key, score, member := b.getKeyScoreAndMember(ctx, ts, value)
	if key != "d_2-123-1598227200000" {
		t.Error()
	}
	if fmt.Sprintf("%f", score) != "1598261325123.060547" {
		t.Error()
	}
	if member != "1.234:61325123.060547" {
		t.Error()
	}
}
