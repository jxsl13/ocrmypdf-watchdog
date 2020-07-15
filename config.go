package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	knownUIDs = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "13", "34", "38", "39", "41", "100", "65534"}

	knownGIDs = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "12", "13", "15", "20", "21", "22", "24", "25", "26", "27", "29", "30", "33", "34", "37", "38", "39", "40", "41", "42", "43", "44", "45", "46", "50", "60", "100", "65534"}
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
	UID int

	// PGID is the GID of the output file
	GID int
}

// NewConfig creates a new Configuration from the container environment
func NewConfig() Config {

	// default values
	cfg := Config{
		"/in/",
		"/out/",
		"ocrmypdf",
		"--pdf-renderer sandwich --tesseract-timeout 1800 --rotate-pages -l eng+fra+deu --deskew --clean --skip-text",
		1000,
		100,
	}

	// fetch PGID from environment variables
	gid := os.Getenv("PGID")
	igid, err := strconv.Atoi(gid)
	if err != nil {

		// fallback if invalid GID
		gid = "100"
		igid = 100
		log.Println("PGID is not an integer:", err)
	}

	// if GID does not yet exist, try to create a random new group with that GID
	if !contains(gid, knownGIDs) {

		// generate groupname from a UUID (nearly impossible to have two groups with the same name)
		groupName := generateString()

		// execute linux command to add a new group with given GID
		cmd := exec.Command("addgroup", "--gid", gid, groupName)
		output, err := cmd.CombinedOutput()

		log.Println(string(output))

		// print error
		if err != nil {
			log.Println("error creating group:", groupName, "error:", err)
		}
	}

	// set config's GID to the newly created ID
	cfg.GID = igid

	// get environment variable PUID & convert to an integer
	uid := os.Getenv("PUID")
	iuid, err := strconv.Atoi(uid)
	if err != nil {

		// fallback to 1000
		uid = "1000"
		iuid = 1000
		log.Println("PUID is not an integer:", err)
	}

	// user does not yet exist
	if !contains(uid, knownUIDs) {

		// create unique username
		userName := generateString()

		// create user & add to previously created group
		cmd := exec.Command("adduser", "--no-create-home", "--disabled-password", "--uid", uid, "--gid", gid, "--gecos", "\"\"", userName)
		output, err := cmd.CombinedOutput()

		// print output
		log.Println(string(output))
		if err != nil {
			// print error
			log.Println("error creating user:", userName, "error:", err)
		}
	}

	// set UID to either fallback or to the newly created user's UID
	cfg.UID = iuid

	// is args not empty, pass them to the configuration file.
	if args := os.Getenv("OCRMYPDF_ARGS"); args != "" {
		cfg.OCRmyPDFArgs = args
	}

	return cfg
}

func (cfg *Config) String() string {
	sb := strings.Builder{}

	sb.WriteString("========== Configuration ==========\n")
	sb.WriteString("OCRmyPDF command:\n")
	sb.WriteString(fmt.Sprintf("\t%s %s\n", cfg.OCRmyPDFExecutable, cfg.OCRmyPDFArgs))

	inInfo, _ := fileInfo("/in/")
	outInfo, _ := fileInfo("/out")
	sb.WriteString(inInfo)
	sb.WriteString(outInfo)

	sb.WriteString("===================================")

	return sb.String()
}
