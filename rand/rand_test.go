package pkgrand

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIntRange(t *testing.T) {
	r := NewMathRand(time.Now().UnixNano())
	min, max := int64(0), int64(10)

	for i := 0; i < 1000; i++ {
		n := IntRange(r, min, max)
		assert.GreaterOrEqual(t, n, min)
		assert.LessOrEqual(t, n, max)
	}
}
