package slidingwindow

import (
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

// Sample is a sliding window sample with a maximum reservoir size in the window.
type Sample struct {
	values        []sample
	reservoirSize int
	mu            sync.Mutex
	window        time.Duration
	c             clock
}

type sample struct {
	v int64
	t time.Time
}

type clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}

// NewSample constructs a new sliding-window sample over the given time window. At the end of each window
// the metrics outside the window are removed. This is an ideal sampling method for low-volume metrics where sample
// values would stick around indefinitely because there was not a steady stream of new metrics. Please note that the
// reservoir size is the size of two in-memory slices of int64s and time. Times used to hold the sample values and times.
// Therefore, if a large reservoir size is used it can be an extremely inefficient use of memory.
func NewSample(reservoirSize int, window time.Duration) metrics.Sample {
	return &Sample{
		values:        []sample{},
		reservoirSize: reservoirSize,
		window:        window,
		c:             realClock{},
	}
}

// slideWindow removes the samples that are older than the window and preserves the samples that are less than or equal
// to the window. Note, it is the responsibility of the caller of this function to ensure that a lock on s.values is obtained.
func (s *Sample) slideWindow() {
	currentTime := s.c.Now()

	// If there's no values or the oldest value (this is the first element because values are stored chronologically) is
	// within the window sliding is not necessary. This is an optimization to reduce complexity so that the slice is not
	// copied on each call when it is not necessary.
	if len(s.values) == 0 || s.isInWindow(currentTime, s.values[0].t) {
		return
	}

	preserved := make([]sample, 0, len(s.values))
	for i, existing := range s.values {
		if s.isInWindow(currentTime, existing.t) {
			// Since the values are stored in chronological order it is safe to assume if the ith element is in the window
			// then all the elements that follow it are also in the window. This is an optimization to reduce the time
			// complexity spent looping over all the elements.
			preserved = append(preserved, s.values[i:]...)
			break
		}
	}
	s.values = preserved
}

// isInWindow computes the difference between the current time and time the metric was collected to determine if it is
// less than or equal to the window size which means the metric should be preserved.
func (s *Sample) isInWindow(current, metric time.Time) bool {
	return current.Sub(metric) <= s.window
}

// Clear deletes all the sample values.
func (s *Sample) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = []sample{}
}

// Count returns the number of sample values.
func (s *Sample) Count() int64 {
	return int64(s.Size())
}

// Max returns the maximum sample v.
func (s *Sample) Max() int64 {
	return metrics.SampleMax(s.Values())
}

// Mean returns the mean of the values in the sample.
func (s *Sample) Mean() float64 {
	return metrics.SampleMean(s.Values())
}

// Min returns the minimum of the values in the sample.
func (s *Sample) Min() int64 {
	return metrics.SampleMin(s.Values())
}

// Percentile returns an arbitrary percentile of values in the sample.
func (s *Sample) Percentile(p float64) float64 {
	return metrics.SamplePercentile(s.Values(), p)
}

// Percentiles returns a slice of arbitrary percentiles of values in the sample.
func (s *Sample) Percentiles(ps []float64) []float64 {
	return metrics.SamplePercentiles(s.Values(), ps)
}

// Size returns the size of the sample, which will not exceed the reservoir size.
func (s *Sample) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.slideWindow()
	return len(s.values)
}

// Snapshot returns a copy of the samples that is flushed to graphite.
func (s *Sample) Snapshot() metrics.Sample {
	v := s.Values()
	return metrics.NewSampleSnapshot(int64(len(v)), v)
}

// StdDev returns the standard deviation of the values in the sample.
func (s *Sample) StdDev() float64 {
	return metrics.SampleStdDev(s.Values())
}

// Sum returns the sum of the values in the sample.
func (s *Sample) Sum() int64 {
	return metrics.SampleSum(s.Values())
}

// Update adds a new value v to the samples.
func (s *Sample) Update(v int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.slideWindow()

	// If the reservoir is met or exceeded skip storing the value to ensure a tight upper bound is kept on memory.
	if len(s.values) >= s.reservoirSize {
		return
	}

	s.values = append(s.values, sample{v: v, t: s.c.Now()})
}

// Values returns a copy of the values in the sample.
func (s *Sample) Values() []int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.slideWindow()
	v2 := make([]int64, len(s.values))
	for i, v := range s.values {
		v2[i] = v.v
	}
	return v2
}

// Variance returns the variance of the values in the sample.
func (s *Sample) Variance() float64 {
	return metrics.SampleVariance(s.Values())
}
