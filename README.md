![CrowdStrike Go Metrics Sliding Window](/img/cs-logo.png?raw=true)
# Sliding Window Sampling for go-metrics

## Overview
This project contains a simple implementation of a sliding window sampling approach implementing the [rcrowley/go-metrics](https://github.com/rcrowley/go-metrics) Sample interface.
The user can define a reservoir size, much like the default sampling approach, but also define a window size that is used to
phase out sample values once they fall outside the window.

## Rationale
The problem with the default exponential decay sampling approach is that it only expires old data as new data comes in.
While this approach works perfectly fine for high-volume metrics, it is problematic for low-volume metrics and can lead to 
misleading graphs where the data is stale, but looks like it is recent. This issue is [particularly troublesome when measuring
metrics with a histogram](http://taint.org/2014/01/16/145944a.html) (e.g., endpoint latency).

Ideally this contribution would be made to the go-metrics project itself, but the author has [made it clear](https://github.com/rcrowley/go-metrics/pull/99)
he would rather have others implement this themselves and leverage the exported interfaces. 

## Examples
```go
package main

import (
	"time"
	
	"github.com/crowdstrike/go-metrics-sliding-window"
	"github.com/rcrowley/go-metrics"
)

func main() {
	// Creates a histogram using the sliding window sampling approach with a reservoir size of 1024 and a sampling window of 2 minutes.
	sample := slidingwindow.NewSample(1024, time.Minute*2)
	histogram := metrics.GetOrRegisterHistogram("histogram.latency", metrics.DefaultRegistry, sample)
	histogram.Update(1)

	// Uses the histogram with the sliding window sampling to create a timer.
	timer := metrics.NewCustomTimer(histogram, metrics.NewMeter())
	timer.Time(myFunc())
}
```
