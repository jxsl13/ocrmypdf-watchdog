package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/jxsl13/ocrmypdf-watchdog/config"
)

var (
	jobs chan string
)

func main() {
	log.Println("starting ocrmypdf-watchdog...")
	cfg := config.New()
	defer config.Close()

	// shared file path channel
	jobs = make(chan string, cfg.NumWorkers)
	// start x worker routines
	startWorkers(cfg.NumWorkers, jobs)

	watcher := cfg.Watcher()
	ctx := cfg.Context()
	log.Println("started ocrmypdf-watchdog")
	for {
		select {
		case <-ctx.Done():
			log.Println("context closed...")
			return
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("errors channel is closed...")
				return
			}
			log.Println("error:", err)

		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("events channel is closed...")
				return
			}
			// do this asynchronously
			go func(e fsnotify.Event) {
				if e.Op&fsnotify.Create == fsnotify.Create {
					filePath := e.Name
					if !IsPDF(filePath) {
						return
					}
					// add file to jobs
					jobs <- filePath
					// print afte rthe job has been added, as the channel above might be blocking
					log.Println(e.Op.String(), ": file:", filePath)
				}
			}(event)
		}
	}

}
