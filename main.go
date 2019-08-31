package main

import (
	"flag"
	"log"
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
		files, err := FilePathWalkDir(site.LocalPath, site.Exclusions)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			syncFile(f, timeout, site)
		}
	}
}
