package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/PastureStack/hosts-file-updater/internal/metadata"
	"github.com/PastureStack/hosts-file-updater/updater"
)

const (
	defaultMetadataURL = "http://metadata/2015-12-19"
)

var (
	VERSION = "dev"
	debug   bool
)

type options struct {
	metadataURL     string
	metadataTimeout time.Duration
	updateInterval  int
	showVersion     bool
	debug           bool
	exit            bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		errorf("%v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	opts, err := parseOptions(args, os.Stdout)
	if err != nil {
		return err
	}
	if opts.exit {
		return nil
	}
	if opts.showVersion {
		fmt.Println(VERSION)
		return nil
	}
	debug = opts.debug

	metadataClient, err := metadata.NewClientAndWait(opts.metadataURL, opts.metadataTimeout)
	if err != nil {
		return err
	}

	u := &updater.Updater{
		MetadataClient: metadataClient,
	}

	metadataClient.OnChange(opts.updateInterval, u.Run)
	// It never exits
	return nil
}

func parseOptions(args []string, output io.Writer) (options, error) {
	opts := options{}
	flags := flag.NewFlagSet("hosts-file-updater", flag.ContinueOnError)
	flags.SetOutput(output)

	flags.IntVar(&opts.updateInterval, "update-interval", 5, "time interval between refreshes of host list, in seconds")
	flags.StringVar(&opts.metadataURL, "metadata-url", defaultMetadataURL, "base URL of the legacy metadata API")
	flags.DurationVar(&opts.metadataTimeout, "metadata-timeout", 10*time.Second, "timeout for each metadata HTTP request")
	flags.BoolVar(&opts.showVersion, "version", false, "print version and exit")
	flags.BoolVar(&opts.debug, "debug", false, "enable debug logging")

	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "PastureStack Hosts File Updater\n\n")
		fmt.Fprintf(flags.Output(), "Populates /etc/hosts from a legacy metadata API.\n\n")
		fmt.Fprintf(flags.Output(), "Usage of %s:\n", flags.Name())
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
		opts.exit = true
		return opts, nil
	} else if err != nil {
		return opts, err
	}
	if flags.NArg() != 0 {
		return opts, fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}
	if opts.showVersion {
		return opts, nil
	}
	if opts.updateInterval <= 0 {
		return opts, fmt.Errorf("update-interval must be greater than 0")
	}
	if opts.metadataURL == "" {
		return opts, fmt.Errorf("metadata-url must not be empty")
	}
	if opts.metadataTimeout < 0 {
		return opts, fmt.Errorf("metadata-timeout must not be negative")
	}

	return opts, nil
}

func debugf(format string, args ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG: "+format+"\n", args...)
	}
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
}
