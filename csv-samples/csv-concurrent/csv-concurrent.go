// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-18
package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const largeCSVFile = "../1000k.csv"

var (
	// list of channels to communicate with workers
	// Those will be accessed synchronously no mutex required
	workers = make(map[string]chan []string)

	// wg is to make sure all workers done before exiting main
	wg = sync.WaitGroup{}

	// mu used only for sequential printing, not relevant for program logic
	mu = sync.Mutex{}
)

func main() {
	initProcess()
}

func initProcess() {
	// wait for all workers to finish up before exit
	defer waitTilEnd()()

	f, err := os.Open(largeCSVFile)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile()(f)

	r := csv.NewReader(f)

	for {
		rec, err := r.Read()
		if err != nil {
			if err == io.EOF {
				savePartitions()
				return
			}
			log.Fatal(err) // sorry for the panic
		}
		process(rec, true)
	}
}

func process(rec []string, first bool) {
	l := len(rec)
	part := rec[l-1]

	if c, ok := workers[part]; ok {
		// send rec to worker
		c <- rec
	} else {
		// if no worker for the partition

		// make a chan
		nc := make(chan []string)
		workers[part] = nc

		// start worker with this chan
		go worker(nc, first)

		// send rec to worker via chan
		nc <- rec
	}
}

func worker(c chan []string, first bool) {
	// wg.Done signals to main worker completion
	wg.Add(1)
	defer wg.Done()

	var part [][]string
	for {
		// wait for a rec or close(chan)
		rec, ok := <-c
		if ok {
			// save the rec
			// instead of accumulation in memory
			// this can be saved to file directly
			part = append(part, rec)
		} else {
			// channel closed on EOF

			// dump partition
			// locks ensures sequential printing
			// not a required for independent files
			mu.Lock()
			for _, p := range part {
				if first {
					fmt.Printf("%+v\n", p)
				}
			}
			mu.Unlock()

			return
		}
	}
}

// simply signals to workers to stop
func savePartitions() {
	for _, c := range workers {
		// signal to all workers to exit
		close(c)
	}
}

func waitTilEnd() func() {
	return func() {
		wg.Wait()
		fmt.Println("File processed.")
	}
}

func closeFile() func(f *os.File) {
	return func(f *os.File) {
		fmt.Println("Defer: closing file.")
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}
