package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func processFile(filePath string) {
	// big part taken from https://github.com/bernmic/ocrmypdf-watchdog/blob/5437c1827298c7223d754819d644b95dd9ddd605/main.go#L92

	log.Println("Processing file " + filePath)
	printInfo(filePath)

	// first get the parts of the path: dir+filename+ext
	directory := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	extension := filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(extension)]
	// try to rename file
	tmpFile, err := ioutil.TempFile(directory, filename+".*"+extension)
	if err != nil {
		log.Printf("Unable to create temp file: %v", err)
		return
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())
	err = os.Rename(filePath, tmpFile.Name())
	if err != nil {
		log.Printf("Cannot rename file. Stopping here: %v", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	target := cfg.Out
	if !strings.HasSuffix(target, "/") {
		target = target + "/"
	}
	targetWithoutExtension := target + filename
	target = targetWithoutExtension + ".tmp"
	log.Printf("Run command: %s %s %s %s\n", cfg.OCRmyPDFExecutable, cfg.OCRmyPDFArgs, tmpFile.Name(), target)
	args := strings.Split(cfg.OCRmyPDFArgs, " ")
	args = append(args, tmpFile.Name(), target)
	cmd := exec.Command(cfg.OCRmyPDFExecutable, args...)

	out, err := cmd.CombinedOutput()

	log.Println(string(out))

	if err != nil {
		// error: tmp back to original name
		log.Printf("Job finished with result %v\n", err)
		os.Rename(tmpFile.Name(), filePath)
	} else {
		log.Printf("Job finished with result successfully.")

		// ok: rename tmp target to final target
		final := targetWithoutExtension + ".pdf"
		os.Rename(target, final)

		printInfo("/out/visible.jpg")
		printInfo(final)
		// set external
		err = os.Chown(final, cfg.UID, cfg.GID)
		if err != nil {
			log.Println("Failed to change owner:", err)
			return
		}
		printInfo(final)
		os.Chmod(final, 0002)
		printInfo(final)

	}
}
