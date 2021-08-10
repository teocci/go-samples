// Package utils_time Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-10
package utils_time

import "time"

const msInSeconds int64 = 1e3
const nsInSeconds int64 = 1e6

// UnixToMS Converts Unix Epoch from milliseconds to time.Time
func UnixToMS (ms int64) time.Time {
	return time.Unix(ms/msInSeconds, (ms%msInSeconds)*nsInSeconds)
}