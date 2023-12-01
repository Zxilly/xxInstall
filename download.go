package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v56/github"
)

var githubClient = github.NewClient(nil)

func downloadBinary(prerelease bool, version string) {
	log.Println("Getting releases")
	releases, _, err := githubClient.Repositories.ListReleases(
		context.Background(),
		"SagerNet",
		"sing-box",
		nil,
	)

	if err != nil {
		log.Fatalf("Error getting releases: %s", err)
	}
	log.Println("Got releases")

	var latestRelease *github.RepositoryRelease

	if version != "" {
		for _, release := range releases {
			if release.GetTagName() == version {
				if release.GetPrerelease() && !prerelease {
					log.Fatalf("Version %s is a prerelease, use --prerelease to download", version)
				}

				latestRelease = release
				break
			}
		}
		if latestRelease == nil {
			log.Fatalf("No release found for version %s", version)
		}
	} else {
		for _, release := range releases {
			if release.GetPrerelease() && !prerelease {
				continue
			}
			if latestRelease == nil {
				latestRelease = release
			} else if release.GetPublishedAt().After(latestRelease.GetPublishedAt().Time) {
				latestRelease = release
			}
		}

		if latestRelease == nil {
			log.Fatalf("No release found")
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
		if strings.HasSuffix(name, goos+"-"+goarch+extension) {
			target = asset
			break
		}
	}

	if target == nil {
		log.Fatalf("No asset found for %s-%s", goos, goarch)
	}

	log.Println("Downloading asset")
	// download the asset
	resp, err := http.Get("https://gp.zxilly.dev/" + target.GetBrowserDownloadURL())
	if err != nil {
		log.Fatalf("Error downloading asset: %s", err)
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)

	barOptions := []progressbar.Option{
		progressbar.OptionSetDescription("downloading"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(10),
		progressbar.OptionThrottle(65 * time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	}
	if shouldSend {
		barOptions = append(barOptions, progressbar.OptionSetWriter(senderConn))
		barOptions = append(barOptions, progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(senderConn, "\n")
		}))
	} else {
		barOptions = append(barOptions, progressbar.OptionSetWriter(os.Stdout))
		barOptions = append(barOptions, progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stdout, "\n")
		}))
	}

	bar := progressbar.NewOptions64(resp.ContentLength, barOptions...)

	_, err = io.Copy(io.MultiWriter(buf, bar), resp.Body)
	if err != nil {
		log.Fatalf("Error reading asset: %s", err)
	}

	extractCompressedFile(buf.Bytes())
}

func extractCompressedFile(fileByte []byte) {
	defer log.Println("Extracted asset")
	if runtime.GOOS == "windows" {
		zipReader, err := zip.NewReader(bytes.NewReader(fileByte), int64(len(fileByte)))
		if err != nil {
			log.Fatalf("Error reading zip: %s", err)
		}

		for _, file := range zipReader.File {
			if strings.HasPrefix(file.Name, "sing-box") {
				fo, err := file.Open()
				if err != nil {
					log.Fatalf("Error opening file: %s", err)
				}
				defer fo.Close()

				file, err := os.Create(BinaryFile)
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
				return
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

			if strings.HasPrefix(header.Name, "sing-box") {
				file, err := os.Create(BinaryFile)
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
				return
			}
		}
	}
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

	err = os.WriteFile(ConfigFile, content, 0644)
	if err != nil {
		log.Fatalf("Error writing config: %s", err)
	}
	log.Println("Wrote config")
}
