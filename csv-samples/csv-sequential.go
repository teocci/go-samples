// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-18
package main
import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strings"
)

// DUMMY
var csvFile1 = strings.NewReader(`
fsdio,abc,def,2017,11,06,01
1sdf9,abc,def,2017,11,06,01
1d243,abc,def,2017,11,06,01
1v2t3,abc,def,2017,11,06,01
a1523,abc,def,2017,11,06,01
1r2r3,abc,def,2017,11,06,02
11213,abc,def,2017,11,06,02
g1253,abc,def,2017,11,06,02
d1e23,abc,def,2017,11,06,02
a1d23,abc,def,2017,11,06,02
12jj3,abc,def,2017,11,06,03
t1r23,abc,def,2017,11,06,03
2123r,abc,def,2017,11,06,03
22123,abc,def,2017,11,06,03
14d23,abc,def,2017,11,06,04
1da23,abc,def,2017,11,06,04
12fy3,abc,def,2017,11,06,04
12453,abc,def,2017,11,06,04`)

func main() {
	// CSV
	r := csv.NewReader(csvFile1)
	partitions := make(map[string][][]string)

	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				err = nil

				savePartitions(partitions)

				return
			}
			log.Fatal(err)
		}

		process(rec, partitions)
	}

}

//savePartitions prints only
func savePartitions(partitions map[string][][]string) {
	for part, recs := range partitions {
		fmt.Println(part)
		for _, rec := range recs {
			fmt.Println(rec)
		}
	}
}

//process this can also write/append directly to a file
func process(rec []string, partitions map[string][][]string) {
	l := len(rec)
	part := rec[l-1]
	if p, ok := partitions[part]; ok {
		partitions[part] = append(p, rec)
	} else {
		partitions[part] = [][]string{rec}
	}
}


