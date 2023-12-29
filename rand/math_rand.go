package pkgrand

import (
	"math/rand"
	"sync"
)

type mathRand struct {
	mRand *rand.Rand
	mu    sync.Mutex
}

// NewMathRand returns a new Rand that uses math/rand.
func NewMathRand(seed int64) Rand {
	src := rand.NewSource(seed)
	r := rand.New(src)

	rand.Float64()

	return &mathRand{mRand: r}
}

func (r *mathRand) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	n, err = r.mRand.Read(p)
	r.mu.Unlock()
	return
}

func (r *mathRand) Float64() (n float64) {
	r.mu.Lock()
	n = r.mRand.Float64()
	r.mu.Unlock()
	return
}

func (r *mathRand) Float32() (n float32) {
	r.mu.Lock()
	n = r.mRand.Float32()
	r.mu.Unlock()
	return
}

func (r *mathRand) Int63n(i int64) (n int64) {
	r.mu.Lock()
	n = r.mRand.Int63n(i)
	r.mu.Unlock()
	return
}

func (r *mathRand) Uint64() (n uint64) {
	r.mu.Lock()
	n = r.mRand.Uint64()
	r.mu.Unlock()
	return
}

func (r *mathRand) Intn(n int) (i int) {
	r.mu.Lock()
	i = r.mRand.Intn(n)
	r.mu.Unlock()
	return
}
