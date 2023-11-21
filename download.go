package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/google/go-github/v56/github"
)

var githubClient = github.NewClient(nil)

func downloadBinary() {
	releases, _, err := githubClient.Repositories.ListReleases(
		context.Background(),
		"SagerNet",
		"sing-box",
		nil,
	)

	if err != nil {
		log.Fatalf("Error getting releases: %s", err)
	}

	latestRelease := releases[0]
	for _, release := range releases {
		if release.GetPublishedAt().After(latestRelease.GetPublishedAt().Time) {
			latestRelease = release
		}
	}

	if len(latestRelease.Assets) < 1 {
		log.Fatalf("No assets found for release %s", latestRelease.GetTagName())
	}

	log.Println("Found latest release: " + latestRelease.GetTagName())

	assets := latestRelease.Assets
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var extension string
	if goos == "windows" {
		extension = ".zip"
	} else {
		extension = ".tar.gz"
	}
	
	var target *github.ReleaseAsset

	for _, asset := range assets {
		name := asset.GetName()
		if strings.HasSuffix(name, goos + "-" + goarch + extension) {
			target = asset
			break
		}
	}

	if target == nil {
		log.Fatalf("No asset found for %s-%s", goos, goarch)
	}

	// download the asset
	resp, err := http.Get(target.GetBrowserDownloadURL())
	if err != nil {
		log.Fatalf("Error downloading asset: %s", err)
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading asset: %s", err)
	}
	log.Println("Downloaded asset")
	
	extractCompressedFile(content)
}

func extractCompressedFile(fileByte []byte) {
	if runtime.GOOS == "windows" {
		zipReader,err  := zip.NewReader(bytes.NewReader(fileByte), int64(len(fileByte)))
		if err != nil {
			log.Fatalf("Error reading zip: %s", err)
		}

		for _, file := range zipReader.File {
			if strings.HasPrefix(file.Name, "sing-box"){
				fo,err := file.Open()
				if err != nil {
					log.Fatalf("Error opening file: %s", err)
				}
				defer fo.Close()

				file, err := os.Create(BINARY_FILE)
				if err != nil {
					log.Fatalf("Error creating file: %s", err)
				}
				defer file.Close()

				_, err = io.Copy(file, fo)
				if err != nil {
					log.Fatalf("Error copying file: %s", err)
				}
				
				err = file.Chmod(0755)
				if err != nil {
					log.Fatalf("Error chmod file: %s", err)
				}
			}
		}
	} else {
		rawTarStream, err := gzip.NewReader(bytes.NewReader(fileByte))
		if err != nil {
			log.Fatalf("Error reading tar: %s", err)
		}

		tarReader := tar.NewReader(rawTarStream)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error reading tar: %s", err)
			}

			if strings.HasPrefix(header.Name, "sing-box"){
				file, err := os.Create(BINARY_FILE)
				if err != nil {
					log.Fatalf("Error creating file: %s", err)
				}
				defer file.Close()

				_, err = io.Copy(file, tarReader)
				if err != nil {
					log.Fatalf("Error copying file: %s", err)
				}
				
				err = file.Chmod(0755)
				if err != nil {
					log.Fatalf("Error chmod file: %s", err)
				}
			}
		}
	}	
	log.Println("Extracted asset")
}

func downloadConfig(u string) {
	resp, err := http.Get(u)
	if err != nil {
		log.Fatalf("Error downloading config: %s", err)
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
	}
	log.Println("Downloaded config")

	err = os.WriteFile(CONFIG_FILE, content, 0644)
	if err != nil {
		log.Fatalf("Error writing config: %s", err)
	}
	log.Println("Wrote config")
}