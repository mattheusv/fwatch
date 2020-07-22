package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/msAlcantara/fwatch"
)

var (
	dir = flag.String("dir", ".", "Directory to watch for file changes")

	ignorePattern = flag.String("ignore", "", "Comma separated list of pattern to ignore files")

	pattern = flag.String("pattern", "", "Comma separated list of pattern to watch files")

	debug = flag.Bool("V", false, "Execute in verbose mode")

	logger *log.Logger
)

func main() {
	flag.Parse()

	args := flag.Args()[0:]
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *debug {
		logger = log.New(os.Stderr, "", log.LstdFlags)
		logger.Println("Debug enable")
	} else {
		logger = log.New(ioutil.Discard, "", log.LstdFlags)
	}

	w, err := fwatch.NewWatcher(*dir, args, strings.Split(*pattern, ","), strings.Split(*ignorePattern, ","), logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	go func() {
		if err := w.Watch(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	<-c
	w.Stop()
}
