package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jxsl13/ocrmypdf-watchdog/config"
	"github.com/jxsl13/ocrmypdf-watchdog/internal"
)

func processFile(filePath string) {

	cfg := config.New()

	log.Println("Processing file: " + filePath)
	internal.PrintInfo(filePath)

	// first get the parts of the path: (dir)+(filename)+(.ext)
	directory := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	extension := filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(extension)]

	// try to create temp file that can be used
	tmpFile, err := ioutil.TempFile(directory, filename+".*"+extension)
	if err != nil {
		log.Printf("Unable to create temp file: %v", err)
		return
	}

	// close file and delete it
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	// move pdf to that tempfile's name
	err = os.Rename(filePath, tmpFile.Name())
	if err != nil {
		log.Printf("Cannot rename file: %v", err)
		return
	}
	// delete temp file at the end
	defer os.Remove(tmpFile.Name())

	target := cfg.OutDir
	if !strings.HasSuffix(target, "/") {
		target = target + "/"
	}

	// OCR
	targetWithoutExtension := target + filename
	target = targetWithoutExtension + ".tmp"
	log.Printf("Run command: %s %s %s %s\n", cfg.OCRMyPDFExecutable, cfg.OCRMyPDFArgsString, tmpFile.Name(), target)

	// add tmp file input, target output
	args := append(cfg.OCRMyPDFArgs, tmpFile.Name(), target)

	// execute OCR
	cmd := exec.Command(cfg.OCRMyPDFExecutable, args...)
	out, err := cmd.CombinedOutput()
	log.Println(string(out))

	if err != nil {
		// error: tmp back to original name
		log.Printf("Job finished failed: %v\n", err)
		os.Rename(tmpFile.Name(), filePath)
	} else {
		log.Printf("Job finished successfully.")

		// ok: rename tmp target to final target
		final := targetWithoutExtension + ".pdf"

		// OCR'ed target file is renamed to final file
		os.Rename(target, final)

		// set external permissions to user's UID and GID
		// the problem that also the Synology support was not able to solve, is that
		// the Synology Drive app cannot properly work with docker files.
		// this means that a scanner will neccessarily have to use the specific
		// user's access rights in order to properly copy those pemrissions over.
		err = os.Chown(final, cfg.UID, cfg.GID)
		if err != nil {
			log.Println("Failed to change owner:", err)
			return
		}

		// safe chmod
		err = os.Chmod(final, 0600)
		if err != nil {
			log.Println("Failed to change file permissions:", err)
			return
		}
		internal.PrintInfo(final)
	}
}
