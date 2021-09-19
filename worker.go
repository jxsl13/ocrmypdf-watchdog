package main

import (
	"log"

	"github.com/jxsl13/ocrmypdf-watchdog/config"
)

func startWorkers(size int, jobs <-chan string) {
	for i := 0; i < size; i++ {
		go worker(i, jobs)
	}
	log.Printf("started %d workers\n", size)
}

func worker(id int, jobs <-chan string) {
	ctx := config.New().Context()
	for {
		select {
		case <-ctx.Done():
			log.Printf("closing worker %d...\n", id)
			return
		case job := <-jobs:
			log.Printf("worker %d, new job: %s\n", id, job)
			processFile(job)
		}
	}
}
