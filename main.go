package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func main() {
	var config Config
	var configpath string
	var timeout time.Duration

	flag.StringVar(&configpath, "c", "config.yml", "Path to the config.yml")
	flag.DurationVar(&timeout, "t", 0, "Upload timeout.")
	flag.Parse()

	readFile(&config, configpath)

	wg.Add(len(config.Config))
	for _, site := range config.Config {
		go syncSite(timeout, site)
	}
	wg.Wait()
}
