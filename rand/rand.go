package pkgrand

// Rand is an interface for random number generators.
type Rand interface {
	Float64() float64
	Float32() float32
	Int63n(n int64) int64
	Uint64() uint64
	Intn(n int) int
	Read(p []byte) (n int, err error)
}

// IntRange generates an integer in range of min and max.
// It never panics. It will swap min and max if min is bigger
// than max. [min, max]
func IntRange(r Rand, min, max int64) int64 {
	// Return min if 'min' == 'max'
	if min == max {
		return min
	}

	// Swap min and max if 'min' > 'max'.
	if min > max {
		originMin := min
		min = max
		max = originMin
	}

	// Figure out if the min/max numbers calculation
	// would cause a panic in the Int63() function.
	// For example if math.MinInt64 and math.MaxInt64 are
	// passed as min and max then the result of the calculation
	// bellow will be equal to 0.
	if max-min+1 > 0 {
		return min + r.Int63n(max-min+1)
	}

	// Loop through the range until we find a number that fits.
	for {
		v := int64(r.Uint64())
		if (v >= min) && (v <= max) {
			return v
		}
	}
}

func IntRangeArray(r Rand, count, min, max int64) []int64 {
	res := make([]int64, count)
	for i := int64(0); i < count; i++ {
		res[i] = IntRange(r, min, max)
	}
	return res
}

func UintRange(r Rand, min, max uint) uint {
	if min == max {
		return min
	}

	if min > max {
		originMin := min
		min = max
		max = originMin
	}

	// Figure out if the min/max numbers calculation
	// would cause a panic in the Int63() function.
	if int(max)-int(min)+1 > 0 {
		return uint(r.Intn(int(max)-int(min)+1) + int(min))
	}

	// Loop through the range until we find a number that fits
	for {
		v := uint(r.Uint64())
		if (v >= min) && (v <= max) {
			return v
		}
	}
}

func Float64Range(r Rand, min, max float64) float64 {
	if min == max {
		return min
	}
	return r.Float64()*(max-min) + min
}
