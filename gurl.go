package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create a wait group to report to and a work channel that takes URLs
	wg := new(sync.WaitGroup)
	work := make(chan string)

	// Create workers and start goroutines listening to work channel
	numWorkers := 100
	wg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go worker(work, wg)
	}

	// Add URLs to work channel from CSV
	addToWork(work)

	// Wait for workers to finish
	close(work)
	wg.Wait()

	fmt.Println("Done.")
}

func addToWork(work chan string) {
	// Open the file
	file, err := os.Open("urls.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	csvReader := csv.NewReader(file)

	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			break
		}

		work <- line[0]
	}
}

func worker(work chan string, wg *sync.WaitGroup) {
	for workUrl := range work {
		u, err := url.Parse(workUrl)
		if err != nil {
			fmt.Println("Invalid URL " + workUrl + ", skipping.")
			continue
		}

		ext := path.Ext(u.Path)
		name := sanitizeUrl(u.Scheme + u.Host + strings.TrimSuffix(u.Path, ext) + "?" + u.RawQuery)
		filePath := "./output/" + u.Host + "/"
		fileName := filePath + name + ext

		err = os.MkdirAll(filePath, 0755)
		if err != nil {
			fmt.Println("Error creating directory, skipping:", err)
			continue
		}

		out, err := os.Create(fileName)
		if err != nil {
			fmt.Println("Error creating file, skipping:", err)
			continue
		}
		defer out.Close()

		resp, err := http.Get(workUrl)
		if err != nil {
			fmt.Println("Error getting URL, skipping:", err)
			continue
		}
		defer resp.Body.Close()

		bytes, err := io.Copy(out, resp.Body)
		if err != nil {
			fmt.Println("Error writing file, skipping:", err)
			continue
		}

		fmt.Printf("%d bytes: %s\n", bytes, fileName)
	}

	wg.Done()
}

func sanitizeUrl(url string) string {
	sanitized := stripChars(url)

	whitespace := regexp.MustCompile("\\s+")
	sanitized = whitespace.ReplaceAllLiteralString(sanitized, "-")

	return sanitized
}

func stripChars(str string) string {
	toStrip := [...]string{"~", "`", "!", "@", "#", "$", "%", "^", "&", "*", "(", ")", "_", "=", "+", "[", "{", "]", "}", "\\", "|", ";", ":", "\"", "'", "&#8216;", "&#8217;", "&#8220;", "&#8221;", "&#8211;", "&#8212;", "—", "–", ",", "<", ".", ">", "/", "?"}

	for i := range toStrip {
		str = strings.Replace(str, toStrip[i], "", -1)
	}

	return str
}
