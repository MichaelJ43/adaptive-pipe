package eta

// SimpleMovingAverageMs returns the arithmetic mean of the last n samples in ms.
// If samples is empty, returns 0 and ok false.
func SimpleMovingAverageMs(samples []int64) (avgMs int64, ok bool) {
	if len(samples) == 0 {
		return 0, false
	}
	var sum int64
	for _, s := range samples {
		sum += s
	}
	return sum / int64(len(samples)), true
}
