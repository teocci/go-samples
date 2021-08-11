// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-10
package main

import (
	"fmt"
	"time"
)

const msInSeconds int64 = 1e3
const nsInSeconds int64 = 1e6

// FromUnixMilli Converts Unix Epoch from milliseconds to time.Time
func FromUnixMilli(ms int64) time.Time {
	return time.Unix(ms/msInSeconds, (ms%msInSeconds)*nsInSeconds)
}

func printTimestampFormats() {
	t := time.Now()
	fmt.Printf("Nano: %d \n", t.Nanosecond())

	fmt.Println(t.Format(time.RFC3339))

	t1, e := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	fmt.Println(t1)

	fmt.Println(t.Format("3:04PM"))
	fmt.Println(t.Format("Mon Jan _2 15:04:05 2006"))
	fmt.Println(t.Format("2006-01-02T15:04:05.999999-07:00"))
	form := "3 04 PM"
	t2, e := time.Parse(form, "8 41 PM")
	fmt.Println(t2)
	fmt.Printf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	ansic := "Mon Jan _2 15:04:05 2006"
	_, e = time.Parse(ansic, "8:41PM")
	fmt.Println(e)
}

func main() {
	unixTimes := [...]int64{758991688, 758992188, 758992690, 758993186}
	var unixUTCTimes []time.Time
	for index, unixTime := range unixTimes {
		unixUTCTimes = append(unixUTCTimes, FromUnixMilli(unixTime))
		if index > 0 {
			timeDifference := unixUTCTimes[index].Sub(unixUTCTimes[index-1])
			fmt.Println("Time difference in ms :--->", timeDifference)
		}
	}

	//printTimestampFormats()
	//unixTimeA := time.Unix(758991688, 0) //gives unix time stamp in utc
	//unixTimeB := time.Unix(758991688, 0) //gives unix time stamp in utc
	//
	//unitTimeInRFC3339 := unixTimeA.Format(time.RFC3339) // converts utc time to RFC3339 format
	//
	//fmt.Println("unix time stamp in UTC :--->", unixTimeA)
	//fmt.Println("unix time stamp in unitTimeInRFC3339 format :->", unitTimeInRFC3339)
	//a := makeTimestamp()
	//
	//fmt.Printf("%d \n", a)
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
