// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-20
package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	dir, err := os.Getwd()
	if err != nil || dir == "" {
		log.Fatalf("Error: %v\n", err.Error())
	}

	fmt.Printf("Current Working Dir: %v\n", dir)
}
