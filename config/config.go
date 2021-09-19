package config

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/jxsl13/ocrmypdf-watchdog/internal"
	configo "github.com/jxsl13/simple-configo"
	"github.com/jxsl13/simple-configo/parsers"
)

var (
	cfg       = (*config)(nil)
	once      = sync.Once{}
	closeOnce = sync.Once{}

	// found in /ect/passwd inside of the container
	knownUIDs = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13, 34, 38, 39, 41, 100, 65534}
	knownGIDs = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 15, 20, 21, 22, 24, 25, 26, 27, 29, 30, 33, 34, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 50, 60, 100, 65534}
)

// New creates a config singleton that's only created once
func New() *config {
	if cfg != nil {
		return cfg
	}
	once.Do(func() {
		c := &config{}
		err := configo.ParseEnv(c)
		if err != nil {
			log.Fatalln(err)
		}
		cfg = c
	})
	return cfg
}

// Close does the config cleanup, like closing the watcher, etc.
func Close() {
	closeOnce.Do(func() {
		_, err := configo.Unparse(cfg)
		if err != nil {
			log.Println(err)
		}
	})
}

type config struct {
	LogFlags int

	InDir     string
	OutDir    string
	FailedDir string

	OCRMyPDFExecutable string
	OCRMyPDFArgs       []string
	OCRMyPDFArgsString string

	UID   int // uid
	GID   int // gid
	Chmod uint32

	ctx     context.Context
	watcher *fsnotify.Watcher

	NumWorkers int // number of goroutines processing files
}

func (c *config) TargetFilePermissions(sourceFile string) (uid int, gid int, chmod fs.FileMode, err error) {
	info, err := os.Stat(sourceFile)
	if err != nil {
		return 0, 0, 0, err
	}

	var UID, GID int
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	} else {
		// we are not in linux, this won't work anyway in windows,
		// but maybe you want to log warnings
		UID = -1
		GID = -1
	}

	mode := info.Mode()

	if c.UID >= 0 {
		UID = c.UID
	}

	if c.GID >= 0 {
		GID = c.GID
	}

	if c.Chmod > 0 {
		mode = fs.FileMode(c.Chmod)
	}

	return UID, GID, mode, nil
}

// Context may be used to check the moment when the application is closed
func (c *config) Context() context.Context {
	return c.ctx
}

// Watcher returns the initialized watcher that may be used to listen for file system events in the
// input directore (not recursive, so only to level changes in the directory are )
func (c *config) Watcher() *fsnotify.Watcher {
	return c.watcher
}

// String returns a string representation of the
func (c *config) String() string {
	sb := strings.Builder{}

	sb.WriteString("========== Configuration ==========\n")
	sb.WriteString("ocrmypdf command:\n")
	sb.WriteString(fmt.Sprintf("\t%s %s\n", c.OCRMyPDFExecutable, c.OCRMyPDFArgsString))

	inInfo, _ := internal.FileInfo(c.InDir)
	outInfo, _ := internal.FileInfo(c.OutDir)
	sb.WriteString(inInfo)
	sb.WriteString(outInfo)

	sb.WriteString("===================================")

	return sb.String()
}

func (c *config) Options() configo.Options {
	delimiter := " "
	return configo.Options{
		{
			Key:           "LOG_FLAGS",
			DefaultValue:  "3",
			ParseFunction: parsers.RangesInt(&c.LogFlags, 1, 63),
			PostParseAction: func() error {
				// do set the value
				log.SetFlags(c.LogFlags)
				return nil
			},
		},
		{
			Key:           "OCRMYPDF_EXECUTABLE",
			DefaultValue:  "ocrmypdf",
			ParseFunction: parsers.String(&c.OCRMyPDFExecutable),
		},
		{
			Key:          "OCRMYPDF_ARGS",
			DefaultValue: "--pdf-renderer sandwich --tesseract-timeout 1800 --rotate-pages -l eng+fra+deu --deskew --clean --skip-text",
			ParseFunction: parsers.And(
				parsers.List(&c.OCRMyPDFArgs, &delimiter),
				parsers.String(&c.OCRMyPDFArgsString),
			),
		},
		{
			Key:             "IN_DIRECTORY",
			DefaultValue:    "/in",
			ParseFunction:   parsers.String(&c.InDir),
			PostParseAction: dirMustExist(&c.InDir),
		},
		{
			Key:             "OUT_DIRECTORY",
			DefaultValue:    "/out",
			ParseFunction:   parsers.String(&c.OutDir),
			PostParseAction: dirMustExist(&c.OutDir),
		},
		{
			Key:           "FAILED_DIR_NAME",
			DefaultValue:  "failed",
			ParseFunction: parsers.String(&c.FailedDir),
			PostParseAction: func() error {
				// create full path in input directory
				c.FailedDir = path.Join(c.InDir, c.FailedDir)
				return nil
			},
		},
		{
			Key:           "PGID",
			DefaultValue:  "-1",
			Description:   "set this value to >= 0 in order to force a specific user group id for the resulting file",
			ParseFunction: parsers.Int(&c.GID),
			PostParseAction: func() error {
				if c.GID < 0 || contains(c.GID, knownGIDs) {
					return nil
				}

				// we 'need' to create a new group
				group := generateRandomGroup()
				cmd := exec.Command("addgroup", "--gid", strconv.Itoa(c.GID), group)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("%w: %v", err, string(output))
				}
				return nil
			},
		},
		{
			Key:           "PUID",
			DefaultValue:  "-1",
			Description:   "set this value to >= 0 in order to force a user id for the resulting file",
			ParseFunction: parsers.Int(&c.UID),
			PostParseAction: func() error {
				if c.UID < 0 || contains(c.UID, knownUIDs) {
					return nil
				}
				// we 'need' to create a new user in the container with the given group id
				user := generateRandomUser()
				cmd := exec.Command(
					"adduser",
					"--no-create-home",
					"--disabled-password",
					"--uid", strconv.Itoa(c.UID),
					"--gid", strconv.Itoa(c.GID),
					"--gecos",
					"\"\"",
					user)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("%w: %v", err, string(output))
				}
				return nil
			},
		},
		{
			Key:           "CHMOD",
			DefaultValue:  "0000",
			Description:   "set this value to > 0000 in order to force specific file permissions (chmod) for the resulting file",
			ParseFunction: OctalInt(&c.Chmod),
		},
		{
			Key: "initialize context which is closed upon an yof the signals",
			PostParseAction: func() error {
				// in case the application receives one of these signals,
				// the context is closed
				// anything listening on the context.Done() channel will
				// be able to fetch the signal from that channel
				c.ctx, _ = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
				return nil
			},
		},
		{
			Key: "inizialize directory watcher",
			PreParseAction: func() error {
				w, err := fsnotify.NewWatcher()
				if err != nil {
					return err
				}
				err = w.Add(c.InDir)
				if err != nil {
					defer w.Close()
					return err
				}
				c.watcher = w
				return nil
			},
			PreUnparseAction: func() error {
				log.Println("closing watcher...")
				return c.watcher.Close()
			},
		},
		{
			Key:           "NUM_WORKERS",
			DefaultValue:  "1",
			ParseFunction: parsers.Int(&c.NumWorkers),
			PostParseAction: func() error {
				if c.NumWorkers <= 0 {
					c.NumWorkers = 1
				}
				return nil
			},
		},
	}
}
