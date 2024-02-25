package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	pdq "github.com/MTRNord/pdqhash-go"
	"github.com/MTRNord/pdqhash-go/types"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/h2non/filetype"
)

// This is an example. It is not meant to be run in prod.

func processFile(filename string, detailed bool) {
	pdqhasher := pdq.NewPDQHasher()

	numPDQHash := 0
	var prevHash *types.Hash256
	filepath.Walk(filename, func(fullPath string, item os.FileInfo, err error) error {
		if !item.IsDir() {
			// Check if file is an image
			filetypeRef, err := filetype.MatchFile(fullPath)
			if err != nil {
				log.Fatal(err)
			}
			if filetypeRef.MIME.Type == "image" {
				hashAndQuality := pdqhasher.FromFile(fullPath)
				delta := 0
				if numPDQHash == 0 {
					delta = 0
				} else {
					delta = hashAndQuality.Hash.HammingDistance(prevHash)
				}

				if detailed {
					log.Printf("hash=%s,norm=%d,delta=%d,quality=%d,filename=%s", hashAndQuality.Hash.String(), hashAndQuality.Hash.HammingNorm(), delta, hashAndQuality.Quality, fullPath)
				} else {
					log.Printf("%s,%d,%s", hashAndQuality.Hash.String(), hashAndQuality.Quality, fullPath)
				}

				prevHash = hashAndQuality.Hash
				numPDQHash++
			}
		}
		return nil
	})
}

func main() {
	var folder string
	var detailedOutput bool

	flag.StringVar(&folder, "folder", "", "Folder to scan")
	flag.BoolVar(&detailedOutput, "detailed", false, "Detailed output")

	flag.Parse()

	// Check if folder exists and is a folder
	fileInfo, err := os.Stat(folder)
	if err != nil {
		log.Fatal(err)
	}
	if !fileInfo.IsDir() {
		log.Fatalf("'%s' is not a folder", folder)
	}

	vips.LoggingSettings(nil, vips.LogLevelMessage)
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 0,
		MaxCacheFiles:    5,
		MaxCacheMem:      50 * 1024 * 1024,
		MaxCacheSize:     100,
		ReportLeaks:      false,
		CacheTrace:       false,
		CollectStats:     false,
	})
	defer vips.Shutdown()

	processFile(folder, detailedOutput)
}
