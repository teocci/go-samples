// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-18
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/teocci/go-samples/concurrent-job/dispatcher"
	"github.com/teocci/go-samples/concurrent-job/worker"
)

var c = &http.Client{Timeout: time.Millisecond * 15000}

func main() {
	start := time.Now()
	poolNumber := runtime.NumCPU()
	fmt.Println(poolNumber)

	dd := dispatcher.New(poolNumber).Start(CallApi)

	terms := []int{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3,
		4, 1, 2, 3, 4}

	for i := range terms {
		dd.Submit(worker.Job{
			ID:        i,
			Name:      fmt.Sprintf("JobID::%d", i),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	end := time.Now()
	log.Print(end.Sub(start).Seconds())
}

func CallApi(id int, job worker.Job) error {
	baseURL := "https://age-of-empires-2-api.herokuapp.com/api/v1/civilization/%d"

	ur := fmt.Sprintf(baseURL, job.ID)
	req, err := http.NewRequest(http.MethodGet, ur, nil)
	if err != nil {
		//log.Printf("error creating a request for term %d :: error is %+v", num, err)
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		//log.Printf("error querying for term %d :: error is %+v", num, err)
		return err
	}
	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		//log.Printf("error reading response body :: error is %+v", err)
		return err
	}

	log.Printf("%d  :: ok", id)
	return nil
}
