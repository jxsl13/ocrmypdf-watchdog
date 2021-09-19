package main

import (
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/jxsl13/ocrmypdf-watchdog/config"
	"github.com/jxsl13/ocrmypdf-watchdog/internal"
)

func processFile(filePath string) {

	log.Println("Processing file: " + filePath)

	cfg := config.New()
	uid, gid, perm, err := cfg.TargetFilePermissions(filePath)
	if err != nil {
		log.Printf("failed to fetch file permissions: %s\n", err)
		return
	}
	internal.PrintInfo(filePath)

	// first get the parts of the path: (dir)+(filename)+(.ext)
	//directory := filepath.Dir(filePath)
	filename := filepath.Base(filePath)                  // file name without directories
	failedFilePath := path.Join(cfg.FailedDir, filename) // failed dir + original file name
	targetFileName := path.Join(cfg.OutDir, filename)

	// add tmp file input, target output

	// execute OCR
	args := append(cfg.OCRMyPDFArgs, filePath, targetFileName)
	cmd := exec.Command(cfg.OCRMyPDFExecutable, args...)
	out, err := cmd.CombinedOutput()
	log.Println(string(out))

	if err != nil {
		// error: tmp back to original name
		log.Printf("Job failed: %v\n", err)
		// move not OCR'ed file back inpo failed folder

		if !internal.IsExistingDir(cfg.FailedDir) {
			err := os.MkdirAll(cfg.FailedDir, fs.FileMode(cfg.Chmod))
			if err != nil {
				log.Println("failed to create directory: ", cfg.FailedDir)
			}
			return
		}
		err = internal.Move(filePath, failedFilePath)
		if err != nil {
			log.Printf("failed to move file from %s to %s\n", filePath, failedFilePath)
		}
	} else {
		log.Println("Job finished successfully.")

		// set external permissions to user's UID and GID
		// the problem that also the Synology support was not able to solve, is that
		// the Synology Drive app cannot properly work with docker files.
		// this means that a scanner will neccessarily have to use the specific
		// user's access rights in order to properly copy those pemrissions over.
		err = os.Chown(targetFileName, uid, gid)
		if err != nil {
			log.Println("Failed to change owner:", err)
			return
		}

		// safe chmod
		err = os.Chmod(targetFileName, perm)
		if err != nil {
			log.Println("Failed to change file permissions:", err)
			return
		}
		internal.PrintInfo(targetFileName)
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("failed to remove: %s: %v", filePath, err)
		}
	}
}
