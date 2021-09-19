package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/jxsl13/ocrmypdf-watchdog/config"
)

func main() {
	log.Println("starting ocrmypdf-watchdog...")
	cfg := config.New()
	defer config.Close()

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

			log.Println("event:", event)

			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create {

				filePath := event.Name
				log.Println("file:", filePath)

				if !IsExist(filePath) || !IsPDF(filePath) {
					continue
				}
				// process file
				processFile(filePath)
			}

		}
	}

}
