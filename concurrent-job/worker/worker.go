// Package worker
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-18
package worker

import (
	"time"
)

// Job represents a single entity that should be processed.
// For example a struct that should be saved to database
type Job struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type JobChannel chan Job

type JobQueue chan chan Job

type TaskFunc func(id int, job Job) error

// Worker is a single processor. Typically, it is possible to
// start multiple workers for better throughput
type Worker struct {
	ID      int           // id of the worker
	JobChan JobChannel    // a channel to receive single unit of work
	Queue   JobQueue      // shared between all workers.
	Quit    chan struct{} // a channel to quit working
}

func New(ID int, JobChan JobChannel, Queue JobQueue, Quit chan struct{}) *Worker {
	return &Worker{
		ID:      ID,
		JobChan: JobChan,
		Queue:   Queue,
		Quit:    Quit,
	}
}

func (wr *Worker) Start(fn TaskFunc) {
	go func() {
		for {
			// when available, put the JobChan again on the JobPool
			// and wait to receive a job
			wr.Queue <- wr.JobChan
			select {
			case job := <-wr.JobChan:
				// when a job is received, process
				err := fn(wr.ID, job)
				if err != nil {
					return
				}
			case <-wr.Quit:
				// a signal on this channel means someone triggered
				// a shutdown for this worker
				close(wr.JobChan)
				return
			}
		}
	}()
}

// Stop closes the Quit channel on the worker.
func (wr *Worker) Stop() {
	close(wr.Quit)
}
