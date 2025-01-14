package main

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

// XML structure for parsing S3 ListBucket result
type ListBucketResult struct {
	Contents []struct {
		Key string `xml:"Key"`
	} `xml:"Contents"`
}

var (
	urlFlag     = flag.String("u", "", "S3 bucket URL to retrieve keys from")
	urlFileFlag = flag.String("U", "", "File containing list of S3 bucket URLs")
	threads     = flag.Int("t", 30, "Number of goroutines for downloading")
	limit       = flag.Int("l", 50, "Limit of keys to retrieve from S3 bucket")
	downloadKey = flag.String("d", "", "Download a single key")
	downloadAll = flag.Bool("D", false, "Download all keys found")
	filter      = flag.String("f", "", "Filter keys to display only those containing this substring")
	debug       = flag.Bool("debug", false, "Show detailed error messages")
)

func main() {
	flag.Parse()

	if *urlFlag == "" && *urlFileFlag == "" {
		log.Fatal("Either -u or -U must be specified")
	}

	var keys []string
	if *urlFlag != "" {
		keys = getS3Keys(*urlFlag, *limit, *urlFlag)
	} else if *urlFileFlag != "" {
		urls := readURLsFromFile(*urlFileFlag)
		for _, bucketURL := range urls {
			keys = append(keys, getS3Keys(bucketURL, *limit, bucketURL)...)
		}
	}

	// Only show the list of keys if -d and -D are not used
	if *downloadKey == "" && !*downloadAll {
		for _, key := range keys {
			if *filter == "" || strings.Contains(key, *filter) {
				fmt.Println("Key:", key)
			}
		}
	}

	if *downloadKey != "" {
		downloadSingleKey(*urlFlag, *downloadKey)
	} else if *downloadAll {
		downloadAllKeys(*urlFlag, keys, *threads)
	}
}

// debugLog logs a message only if the --debug flag is set
func debugLog(format string, v ...interface{}) {
	if *debug {
		log.Printf(format, v...)
	}
}

// getS3Keys fetches S3 keys from a bucket URL and parses XML response
// If XML parsing fails, logs the error and skips to the next URL if -U is set.
func getS3Keys(bucketURL string, limit int, prefix string) []string {
	resp, err := http.Get(bucketURL)
	if err != nil {
		debugLog("Failed to retrieve keys from %s: %v", bucketURL, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		debugLog("Failed to retrieve keys from %s, status code: %d", bucketURL, resp.StatusCode)
		return nil
	}

	// Read and parse the XML response to retrieve keys
	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		debugLog("Error reading response body from %s: %v", bucketURL, err)
		return nil
	}

	var result ListBucketResult
	if err := xml.Unmarshal(rawData, &result); err != nil {
		debugLog("Error parsing XML from %s: %v. Skipping to the next URL.", bucketURL, err)
		return nil
	}

	// Extract keys up to the specified limit, prepending with the bucket URL if -U is used
	var keys []string
	for i, content := range result.Contents {
		if i >= limit {
			break
		}
		key := content.Key
		// If -U is set, prepend the bucket URL to each key
		if *urlFileFlag != "" {
			key = fmt.Sprintf("%s/%s", bucketURL, key)
		}
		keys = append(keys, key)
	}

	return keys
}

// downloadSingleKey downloads a single key from the bucket URL
func downloadSingleKey(bucketURL, key string) {
	url := fmt.Sprintf("%s/%s", bucketURL, key)
	downloadAndSave(url, key)
	fmt.Printf("Downloaded %s\n", key)
}

// downloadAllKeys downloads all specified keys concurrently with a progress bar
func downloadAllKeys(bucketURL string, keys []string, threads int) {
	bar := pb.StartNew(len(keys))
	bar.Set(pb.SIBytesPrefix, true)

	sem := make(chan struct{}, threads)
	var wg sync.WaitGroup
	for _, key := range keys {
		wg.Add(1)
		sem <- struct{}{}
		go func(k string) {
			defer wg.Done()
			url := fmt.Sprintf("%s/%s", bucketURL, k)
			downloadAndSave(url, k)
			bar.Increment()
			<-sem
		}(key)
	}
	wg.Wait()
	bar.Finish()
}

// downloadAndSave handles the downloading and saving of a file from a URL
func downloadAndSave(url, key string) {
	resp, err := http.Get(url)
	if err != nil {
		debugLog("Failed to download key %s: %v", key, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		debugLog("Failed to download key %s, status code: %d", key, resp.StatusCode)
		return
	}

	saveToFile(key, resp.Body)
}

// saveToFile saves the downloaded content to a file
func saveToFile(key string, content io.Reader) {
	localFile := filepath.Base(key)
	file, err := os.Create(localFile)
	if err != nil {
		debugLog("Failed to create file %s: %v", localFile, err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		debugLog("Failed to save content for key %s: %v", key, err)
	}
}

// readURLsFromFile reads URLs from a file, one per line
func readURLsFromFile(filename string) []string {
	var urls []string
	file, err := os.Open(filename)
	if err != nil {
		debugLog("Failed to open file: %v", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		debugLog("Failed to read file: %v", err)
	}
	return urls
}
