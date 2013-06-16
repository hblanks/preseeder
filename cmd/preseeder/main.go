package main

import (
	"flag"
	"fmt"
	"github.com/hblanks/preseeder"
	"io/ioutil"
	"log"
	"os"
)

const usage_summary = `
Serves a templated, Debian preseed.cfg at http://*/preseed
using parameters defined in preseed.yaml. Can serve additional files
such as a late command to execute at the end of preseed and an arbitrary
directory of static files. Tracks which hosts have accessed preseed.cfg
& late_command, including their IP and MAC address.`

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] preseed.yaml\n", os.Args[0])
	fmt.Fprintf(os.Stderr, usage_summary)
	flag.PrintDefaults()
	os.Exit(2)
}

func readFile(path string) string {
	readBytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error reading %s. %v", path, err)
	}
	return string(readBytes)
}

func main() {
	keyPath := flag.String("i",
		"", "Path to id_rsa.pub / authorized_keys")
	preseedTemplatePath := flag.String("p",
		"", "Path to a preseed template file")
	listenAddress := flag.String("l",
		":8080", "Address to listen on")
	lateCommandPath := flag.String("x",
		"", "execute this file during late_command")
	staticRoot := flag.String("s",
		"", "serve the given directory of files as http://0.0.0.0:8080/static/")

	help := flag.Bool("h", false, "Print help")
	flag.Parse()

	if *help || flag.NArg() != 1 {
		usage()
	}

	preseedYamlPath := flag.Arg(0)

	var err error
	var preseed string
	var preseedContext *preseeder.PreseedContext
	var authorizedKeys string
	var lateCommand string

	if *preseedTemplatePath != "" {
		preseed = readFile(*preseedTemplatePath)
	}

	preseedContext, err = preseeder.ParseYaml(preseedYamlPath)
	if err != nil {
		log.Fatalf("Error parsing %s. %v", preseedYamlPath, err)
	}

	if len(*keyPath) != 0 {
		authorizedKeys = readFile(*keyPath)
	}

	if *lateCommandPath != "" {
		lateCommand = readFile(*lateCommandPath)
	}

	server := preseeder.NewPreseedServer(
		preseed,
		preseedContext,
		authorizedKeys,
		lateCommand,
		*staticRoot,
	)
	log.Fatal(server.ListenAndServe(*listenAddress, nil))
}
