package main

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/fsnotify/fsnotify"
)

var (
	cfg             = NewConfig()
	errFileNotFound = errors.New("File not found")
)

func init() {

	// allowed flags
	if flags, err := strconv.Atoi(os.Getenv("LOG_FLAGS")); err == nil && 0 <= flags && flags < 64 {
		log.SetFlags(flags)
	} else {
		log.SetFlags(0)
	}

}

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
