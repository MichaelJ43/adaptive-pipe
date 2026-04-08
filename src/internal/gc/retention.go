package gc

// BuildNumbersToExpire returns build numbers that are outside the retention window
// (keep the last `keep` builds, sorted by build number descending).
func BuildNumbersToExpire(sortedDescBuildNumbers []int64, keep int) []int64 {
	if keep < 0 {
		keep = 0
	}
	if len(sortedDescBuildNumbers) <= keep {
		return nil
	}
	out := sortedDescBuildNumbers[keep:]
	// out is ascending order of old builds (since input was desc)
	return out
}
