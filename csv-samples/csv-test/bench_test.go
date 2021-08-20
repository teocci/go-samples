// Package csv_test
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-18
package csv_test

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/gocarina/gocsv"
	"github.com/jszwec/csvutil"
	"github.com/yunabe/easycsv"
)

type A struct {
	A int     `json:"a" csv:"a" name:"a"`
	B float64 `json:"b" csv:"b" name:"b"`
	C string  `json:"c" csv:"c" name:"c"`
	D int64   `json:"d" csv:"d" name:"d"`
	E int8    `json:"e" csv:"e" name:"e"`
	F float32 `json:"f" csv:"f" name:"f"`
	G float32 `json:"g" csv:"g" name:"g"`
	H float32 `json:"h" csv:"h" name:"h"`
	I string  `json:"i" csv:"i" name:"i"`
	J int     `json:"j" csv:"j" name:"j"`
}

func BenchmarkUnmarshal(b *testing.B) {
	fixture := []struct {
		desc    string
		records int
	}{
		{
			desc:    "1 record",
			records: 1,
		},
		{
			desc:    "10 records",
			records: 10,
		},
		{
			desc:    "100 records",
			records: 100,
		},
		{
			desc:    "1000 records",
			records: 1000,
		},
		{
			desc:    "10000 records",
			records: 10000,
		},
		{
			desc:    "100000 records",
			records: 100000,
		},
	}

	tests := []struct {
		desc string
		fn   func([]byte, *testing.B)
	}{
		{
			desc: "csvutil.Unmarshal",
			fn: func(data []byte, b *testing.B) {
				var a []A
				if err := csvutil.Unmarshal(data, &a); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "gocsv.Unmarshal",
			fn: func(data []byte, b *testing.B) {
				var a []A
				if err := gocsv.UnmarshalBytes(data, &a); err != nil {
					b.Error(err)
				}
			},
		},
		{
			desc: "easycsv.ReadAll",
			fn: func(data []byte, b *testing.B) {
				r := easycsv.NewReader(bytes.NewReader(data))
				var a []A
				if err := r.ReadAll(&a); err != nil {
					b.Error(err)
				}
			},
		},
	}

	for _, t := range tests {
		b.Run(t.desc, func(b *testing.B) {
			for _, f := range fixture {
				b.Run(f.desc, func(b *testing.B) {
					data := genData(f.records)
					for i := 0; i < b.N; i++ {
						t.fn(data, b)
					}
				})
			}
		})
	}
}

func genData(records int) []byte {
	header := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	record := []string{"1", "2.5", "foo", "6", "7", "8", "9", "10", "bar", "10"}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Write(header)

	for i := 0; i < records; i++ {
		w.Write(record)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
