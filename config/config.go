package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
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
	knownUIDs = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 13, 34, 38, 39, 41, 100, 65534}
	knownGIDs = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 15, 20, 21, 22, 24, 25, 26, 27, 29, 30, 33, 34, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 50, 60, 100, 65534}
)

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
	InDir              string
	OutDir             string
	UID                int // uid
	GID                int // gid
	Chmod              uint32
	LogFlags           int
	OCRMyPDFExecutable string
	OCRMyPDFArgs       []string
	OCRMyPDFArgsString string
	ctx                context.Context

	watcher *fsnotify.Watcher
}

func (c *config) Context() context.Context {
	return c.ctx
}

func (c *config) Watcher() *fsnotify.Watcher {
	return c.watcher
}

func (c *config) String() string {
	sb := strings.Builder{}

	sb.WriteString("========== Configuration ==========\n")
	sb.WriteString("OCRmyPDF command:\n")
	sb.WriteString(fmt.Sprintf("\t%s %s\n", c.OCRMyPDFExecutable, c.OCRMyPDFArgsString))

	inInfo, _ := internal.FileInfo("/in/")
	outInfo, _ := internal.FileInfo("/out")
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
			Key:           "PGID",
			DefaultValue:  "100",
			ParseFunction: parsers.Int(&c.GID),
			PostParseAction: func() error {
				if contains(c.GID, knownGIDs) {
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
			DefaultValue:  "1000",
			ParseFunction: parsers.Int(&c.UID),
			PostParseAction: func() error {
				if contains(c.UID, knownUIDs) {
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
			DefaultValue:  "0600",
			ParseFunction: OctalInt(&c.Chmod),
		},
		{
			Key: "initialize context which is closed upon an yof the signals",
			PostParseAction: func() error {
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
	}
}
