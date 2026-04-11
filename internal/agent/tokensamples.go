package agent

// MaxTokenSamples bounds the rolling window of per-agent token count samples
// surfaced to the frontend for sparkline rendering. At 20 samples with a
// minimum-delta throttle of 50 tokens, the window spans roughly 20+ updates
// of meaningful history without collapsing into noise.
const MaxTokenSamples = 20

// tokenSampleMinDelta is the minimum absolute token-count delta required to
// append a new sample. Token counts update on every streaming chunk (potentially
// 10/sec), so a pure per-tick push would cause the 20-sample window to collapse
// to a ~2-second span and make the sparkline visually meaningless.
const tokenSampleMinDelta = 50

// MaybeAppendTokenSample returns a new slice containing the existing samples
// plus newCount if it passes the throttle, trimmed to MaxTokenSamples.
//
// The first sample is always appended. Subsequent samples are appended only
// when the absolute delta from the previous sample is >= tokenSampleMinDelta.
//
// This function is pure with respect to its input slice — it never mutates
// existing. Callers holding the slice in shared state must serialize calls
// under their own mutex.
func MaybeAppendTokenSample(existing []int, newCount int64) []int {
	sample := int(newCount)

	if len(existing) == 0 {
		return []int{sample}
	}

	last := existing[len(existing)-1]
	delta := sample - last
	if delta < 0 {
		delta = -delta
	}
	if delta < tokenSampleMinDelta {
		return existing
	}

	updated := append(existing, sample)
	if len(updated) > MaxTokenSamples {
		updated = updated[len(updated)-MaxTokenSamples:]
	}
	return updated
}
