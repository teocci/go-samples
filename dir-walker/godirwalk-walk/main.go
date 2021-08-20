// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-20
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/karrick/godirwalk"
)

func main() {
	optVerbose := flag.Bool("verbose", false, "Print file system entries.")
	flag.Parse()

	dirname, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dirname += "/tmp/5st/5st-logger"

	if flag.NArg() > 0 {
		dirname = flag.Arg(0)
	}

	fmt.Printf("%s\n", dirname)


	err = godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if *optVerbose {
				fmt.Printf("%s %s\n", de.ModeType(), osPathname)
			}
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			if *optVerbose {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			}

			// For the purposes of this example, a simple SkipNode will suffice,
			// although in reality perhaps additional logic might be called for.
			return godirwalk.SkipNode
		},
		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}