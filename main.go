package main

import (
	"flag"
	"time"
)

func main() {
	var config Config
	var configpath string
	var timeout time.Duration

	flag.StringVar(&configpath, "c", "config.yml", "Path to the config.yml")
	flag.DurationVar(&timeout, "t", 0, "Upload timeout.")
	flag.Parse()

	readFile(&config, configpath)

	for _, site := range config.Config {
		syncFile(timeout, site)
	}
}
