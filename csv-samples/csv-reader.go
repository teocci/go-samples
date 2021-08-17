// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-17
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("./test.csv")
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	// csv reader producer
	rdr := csv.NewReader(file)

	// read all the csv content
	rows, err := rdr.ReadAll()
	if err != nil {
		log.Fatalln("Error: ", err)
	}

	// read row and column
	for i, row := range rows {
		for j := range row {
			fmt.Printf("%s ", rows[i][j])
		}
		fmt.Println()
	}
}