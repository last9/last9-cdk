package httpmetrics

import "math"

var LatencyBins = latencyBins(0.25, 1.5, 31)

// With latency, keeping track of individual measurements is a good, first-pass
// solution. It gives you the flexibility to perform aggregations at a later
// time and experiment with different ways of constructing histograms and
// explore metrics in specific time windows, among other tests. However, as
// your scale increases, memory restrictions can make it impossible to record
// every individual measurement. At this point, you’ll need techniques that
// allow you to keep an increasing amount of metrics in a constrained amount of
// memory.
//
// What is the ideal bucket size for the histogram? Two sanely known
// bucketization approaches:
//
// **1. Lay Out the Bucket Sizes as an Arithmetic Sequence**
// This concept is best illustrated with an example:
// [1ms, 3ms, 5ms, 7ms, 9ms, 11ms, … 2000ms]
//
// In the example above, bucket 0 contains the count of latencies that were <=
// 1ms. Bucket 1 contains the count of latencies > 1ms and <= 3ms. The
// difference in buckets in this arithmetic sequence is 2ms. This approach is
// better than measuring each latency separately, but with fixed-width buckets,
// it will take roughly 1,000 buckets to measure latencies up to 2000ms. Most
// of the buckets will be empty, which isn’t ideal.
//
// **2. Lay Out the Bucket Sizes as a Geometric Sequence**
// Again, we’ll start with an example:
// [1ms, 1.5ms, 2.25ms, 3.75ms, 5.07ms, 7.6ms, … 2000ms]
//
// Here, each bucket boundary is 150% of the previous bucket’s boundary. The
// advantage of this style of bucketization is that we get very granular data
// at low latencies, which is the interesting part of the distribution, and the
// data becomes less granular for larger buckets. With 19 buckets, you can
// capture latencies up to 2000ms and each bucket will be used well.
//
// Fun Fact
// A geometric sequence is the default bucketization used at Google to measure
// latency histograms.
//
// Reference Reading
// - https://twitter.com/el_bhs/status/993819711107485697?lang=en
// - https://www.circonus.com/2018/08/latency-slos-done-right/
func latencyBins(stepStart, rate float64, limit int) []float64 {
	var out = []float64{0}
	for i := 1; i < limit; i++ {
		out = append(out, math.Floor((out[len(out)-1]+(stepStart*math.Pow(rate, float64(i-1))))*4)/4)
	}

	return append(out, float64(3600000))[1:]
}