// Package utils_time Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-10
package utils_time

import (
	"testing"
	"time"
)

func BenchmarkMultiplying(b *testing.B) {
	var inp int64
	inp = 1542810446506
	for n := 0; n < b.N; n++ {
		time.Unix(0, inp*int64(time.Nanosecond))
	}
}

func BenchmarkDiv(b *testing.B) {
	var ms int64
	ms = 1542810446506
	for n := 0; n < b.N; n++ {
		time.Unix(ms/msInSec, (ms%msInSec)*nsInMS)
	}
}

func TestUnixToMS(t *testing.T) {
	ms := 1542810446506
	ts := UnixToMS(int64(ms))
	if ts.UTC().String() != "2018-11-21 14:27:26.506 +0000 UTC" {
		t.Errorf("Expected time to be '2018-11-21 14:27:26.506 +0000 UTC' got '%s'", ts.UTC().String())
	}
}