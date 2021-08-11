// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-11
package main

import (
	"os"

	"github.com/teocci/go-samples/internal/core"
)

func main() {
	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	s.Wait()
}
