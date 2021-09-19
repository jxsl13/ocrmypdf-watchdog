package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/jxsl13/ocrmypdf-watchdog/config"
)

func main() {
	cfg := config.New()
	defer config.Close()

	watcher := cfg.Watcher()
	ctx := cfg.Context()

	for {
		select {
		case <-ctx.Done():
			// application is closed via sigint/sigterm
			break
		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("errors channel is closed...")
				break
			}
			log.Println("error:", err)

		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("events channel is closed...")
				break
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
