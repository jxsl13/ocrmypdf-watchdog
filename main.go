package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
)

// Config parameters
type Config struct {

	// In input folder
	In string

	// Out output folder
	Out string

	// Executable
	OCRmyPDFExecutable string
	// OCRmyPDFArgs ar the cli arguments
	OCRmyPDFArgs string

	// PUID is the UID of the output file
	PUID int

	// PGID is the GID of the output file
	PGID int
}

var (
	cfg = Config{
		"/in/",
		"/out/",
		"ocrmypdf",
		"--pdf-renderer sandwich --tesseract-timeout 1800 --rotate-pages -l eng+fra+deu --deskew --clean --skip-text",
		1000,
		100,
	}

	knownUIDs = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "13", "34", "38", "39", "41", "100", "65534"}

	knownGIDs = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "12", "13", "15", "20", "21", "22", "24", "25", "26", "27", "29", "30", "33", "34", "37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "50", "60", "100", "65534"}
)

func generateString() string {
	uuid := uuid.New().String()
	return strings.Replace(uuid, "-", "", -1)[:32]
}

func printInfo(filePath string) {
	info, err := os.Stat(filePath)

	if err != nil {
		log.Printf("Failed to open %q, %q", filePath, err)
		return
	}

	isDir := info.IsDir()
	log.Println("============================== Info ==============================")
	if isDir {
		log.Println("Permissions of folder: ", filePath)
	} else {
		log.Println("Permissions of file: ", filePath)
	}

	log.Println("\tPerm:", info.Mode().Perm())
	log.Println("\tSys:", info.Sys())

	var UID int
	var GID int
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	} else {
		// we are not in linux, this won't work anyway in windows,
		// but maybe you want to log warnings
		UID = -1
		GID = -1
	}
	log.Println("\tUID", UID)
	log.Println("\tGID", GID)
}

func init() {

	gid := os.Getenv("PGID")
	igid, err := strconv.Atoi(gid)
	if err != nil {
		gid = "1000"
		igid = 1000
		log.Println("PGID is not an integer:", err)
	}

	if !contains(gid, knownGIDs) {
		cmd := exec.Command("addgroup", "--gid", gid, generateString())
		output, err := cmd.CombinedOutput()

		log.Println(string(output))
		if err != nil {
			log.Println("error creating group:", err)
		}
	}
	cfg.PGID = igid

	uid := os.Getenv("PUID")
	iuid, err := strconv.Atoi(uid)
	if err != nil {
		uid = "1000"
		iuid = 1000
		log.Println("PUID is not an integer:", err)
	}

	if !contains(uid, knownUIDs) {
		cmd := exec.Command("adduser", "--no-create-home", "--disabled-password", "--uid", uid, "--gid", gid, "--gecos", "\"\"", generateString())
		output, err := cmd.CombinedOutput()

		log.Println(string(output))
		if err != nil {
			log.Println("error creating user:", err)
		}
	}
	cfg.PUID = iuid

	if args := os.Getenv("OCRMYPDF_ARGS"); args != "" {
		cfg.OCRmyPDFArgs = args
	}

	log.Println("========== Configuration ==========")
	log.Println("Owner of output file:")
	log.Printf("\tUID: %d\n", cfg.PUID)
	log.Printf("\tGID: %d\n", cfg.PGID)
	log.Println("OCRmyPDF command:")
	log.Printf("\t%s %s\n", cfg.OCRmyPDFExecutable, cfg.OCRmyPDFArgs)
	printInfo("/out/")
	log.Println("===================================")
}

func contains(element string, in []string) bool {
	for _, s := range in {
		if s == element {
			return true
		}
	}
	return false
}

// IsPDF checks the content type of the file.
func IsPDF(filePath string) bool {
	contentType, err := GetFileContentType(filePath)
	if err != nil {
		log.Println("file type error:", err)
		return false
	}

	log.Printf("content type of %q is %q", filePath, contentType)
	if contentType != "application/pdf" {
		return false
	}
	return true
}

// GetFileContentType checks the file type.
func GetFileContentType(file string) (string, error) {

	// Open File
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = f.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

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

		printInfo(final)
		printInfo("/out/visible.jpg")

		// set external
		err = os.Chown(final, cfg.PUID, cfg.PGID)
		if err != nil {
			log.Println("Failed to change owner:", err)
			return
		}

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
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					filePath := event.Name
					log.Println("modified file:", filePath)

					if !IsPDF(filePath) {
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
