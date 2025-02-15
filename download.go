package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v56/github"
	"github.com/melbahja/got"
	"github.com/minio/selfupdate"
	"github.com/schollz/progressbar/v3"
)

var githubClient = github.NewClient(nil)

const MirrorUrl = "https://acc.zxilly.dev/"

func applySelfUpdate(mirror bool) {
	log.Println("Getting releases")
	releases, _, err := githubClient.Repositories.ListReleases(
		context.Background(),
		"Zxilly",
		"xxInstall",
		nil,
	)
	if err != nil {
		log.Fatalf("Error getting releases: %s", err)
	}

	log.Println("Got releases")

	var latestRelease *github.RepositoryRelease

	for _, release := range releases {
		if latestRelease == nil {
			latestRelease = release
		} else if release.GetPublishedAt().After(latestRelease.GetPublishedAt().Time) {
			latestRelease = release
		}
	}

	if latestRelease == nil {
		log.Fatalf("No release found")
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

	suffix := goos + "_" + goarch + extension

	var target *github.ReleaseAsset
	for _, asset := range assets {
		name := asset.GetName()
		if strings.HasSuffix(name, suffix) {
			target = asset
			break
		}
	}

	if target == nil {
		log.Fatalf("No asset found for %s", suffix)
	}

	log.Println("Downloading asset", target.GetName())

	var dlUrl string
	if mirror {
		dlUrl = MirrorUrl + target.GetBrowserDownloadURL()
	} else {
		dlUrl = target.GetBrowserDownloadURL()
	}

	buf, err := downloadWithProgressBar(dlUrl)
	if err != nil {
		log.Fatalf("Error downloading asset: %s", err)
	}

	r, err := extractCompressedFile(buf, "xx")
	if err != nil {
		log.Fatalf("Error extracting file: %s", err)
	}

	err = selfupdate.Apply(r, selfupdate.Options{})
	if err != nil {
		err = selfupdate.RollbackError(err)
		if err != nil {
			log.Fatalf("Error rolling back: %s", err)
		}
	}

	log.Println("Updated to latest version " + latestRelease.GetTagName())
}

func downloadWithProgressBar(url string) ([]byte, error) {
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

	var bar *progressbar.ProgressBar

	g := got.New()
	g.ProgressFunc = func(d *got.Download) {
		if bar == nil {
			total := d.TotalSize()
			if total == 0 {
				bar = progressbar.NewOptions64(-1, barOptions...)
			} else {
				bar = progressbar.NewOptions64(int64(total), barOptions...)
			}
		} else {
			bar.Set(int(d.Size()))
		}
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "xx")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	err = g.Do(&got.Download{
		Concurrency: 16,
		URL:         url,
		Dest:        tmpFile.Name(),
	})
	if err != nil {
		return nil, err
	}
	err = bar.Finish()
	if err != nil {
		return nil, err
	}

	return os.ReadFile(tmpFile.Name())
}

func downloadBinary(prerelease bool, version string, mirror bool) {
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

	log.Println("Downloading asset", target.GetName())

	var dlUrl string
	if mirror {
		dlUrl = MirrorUrl + target.GetBrowserDownloadURL()
	} else {
		dlUrl = target.GetBrowserDownloadURL()
	}

	buf, err := downloadWithProgressBar(dlUrl)
	if err != nil {
		log.Fatalf("Error downloading asset: %s", err)
	}

	writeToBinaryFile := func(r io.Reader) {
		targetFile, err := os.Create(BinaryFile)
		defer targetFile.Close()
		if err != nil {
			log.Fatalf("Error creating file: %s", err)
		}

		_, err = io.Copy(targetFile, r)
		if err != nil {
			log.Fatalf("Error copying file: %s", err)
		}

		err = targetFile.Chmod(0755)
		if err != nil {
			log.Fatalf("Error chmod file: %s", err)
		}

		log.Println("Wrote binary to " + BinaryFile)
	}

	r, err := extractCompressedFile(buf, "sing-box")
	if err != nil {
		log.Fatalf("Error extracting file: %s", err)
	}

	writeToBinaryFile(r)

	if rc, ok := r.(io.ReadCloser); ok {
		rc.Close()
	}
}

//goland:noinspection GoBoolExpressions
func extractCompressedFile(fileByte []byte, target string) (io.Reader, error) {
	defer log.Println("Extracted asset")
	if runtime.GOOS == "windows" {
		zipReader, err := zip.NewReader(bytes.NewReader(fileByte), int64(len(fileByte)))
		if err != nil {
			log.Fatalf("Error reading zip: %s", err)
		}

		for _, file := range zipReader.File {
			if strings.HasSuffix(file.Name, target+".exe") {
				fo, err := file.Open()
				if err != nil {
					log.Fatalf("Error opening file: %s", err)
				}
				log.Println("Found file " + file.Name)

				return fo, nil
			}
		}
		return nil, fmt.Errorf("no file found")
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

			if strings.HasSuffix(header.Name, target) {
				log.Println("Found file " + header.Name)
				return tarReader, nil
			}
		}
		return nil, fmt.Errorf("no file found")
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
