package main

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

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
	filename := filepath.Base(filePath)                   // file name without directories
	failedFilePath := path.Join(cfg.FailedDir, filename)  // failed dir + original file name
	extension := filepath.Ext(filename)                   // file extension
	filename = filename[0 : len(filename)-len(extension)] // remove file extension

	target := cfg.OutDir
	if !strings.HasSuffix(target, "/") {
		target = target + "/"
	}

	// try to create temp file that can be used
	tmpFile, err := ioutil.TempFile(target, filename+".*"+extension)
	if err != nil {
		log.Printf("Unable to create temp file: %v", err)
		return
	}

	// close file and delete it
	tmpFile.Close()
	err = os.Remove(tmpFile.Name())
	if err != nil {
		log.Printf("failed to remove file %s\n", tmpFile.Name())
	}

	// move pdf to that tempfile's location
	err = internal.Move(filePath, tmpFile.Name())
	if err != nil {
		log.Printf("Cannot move file: %v", err)
		return
	}
	// delete temp file at the end
	defer os.Remove(tmpFile.Name())

	// OCR
	targetWithoutExtension := target + filename
	target = targetWithoutExtension + ".tmp"

	// add tmp file input, target output
	args := append(cfg.OCRMyPDFArgs, tmpFile.Name(), target)

	// execute OCR
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
		err = internal.Move(tmpFile.Name(), failedFilePath)
		if err != nil {
			log.Printf("failed to move file from %s to %s\n", tmpFile.Name(), failedFilePath)
		}
	} else {
		log.Println("Job finished successfully.")

		// ok: rename tmp target to final target
		final := targetWithoutExtension + ".pdf"

		// OCR'ed target file is renamed to final file
		err := os.Rename(target, final)
		if err != nil {
			log.Printf("failed to move file from %s to %s\n", target, final)
			return
		}

		// set external permissions to user's UID and GID
		// the problem that also the Synology support was not able to solve, is that
		// the Synology Drive app cannot properly work with docker files.
		// this means that a scanner will neccessarily have to use the specific
		// user's access rights in order to properly copy those pemrissions over.
		err = os.Chown(final, uid, gid)
		if err != nil {
			log.Println("Failed to change owner:", err)
			return
		}

		// safe chmod
		err = os.Chmod(final, perm)
		if err != nil {
			log.Println("Failed to change file permissions:", err)
			return
		}
		internal.PrintInfo(final)
	}
}
