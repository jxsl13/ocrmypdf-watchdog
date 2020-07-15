package main

import (
	"errors"
	"log"

	"github.com/fsnotify/fsnotify"
)

var (
	cfg             = NewConfig()
	errFileNotFound = errors.New("File not found")
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
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

					processFile(filePath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(cfg.In)
	if err != nil {
		log.Fatal(err)
	}
	<-done

}
