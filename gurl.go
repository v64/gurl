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
	"strings"
	"sync"
)

func main() {
	// Create a wait group to report to and a work channel that takes URLs
	wg := new(sync.WaitGroup)
	work := make(chan string)

	// Create workers and start goroutines listening to work channel
	numWorkers := 25
	wg.Add(numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go worker(work, wg)
	}

	// Add URLs to work channel from CSV
	addWorkFromCsv(work)

	// Wait for workers to finish
	close(work)
	wg.Wait()

	fmt.Println("Done.")
}

func worker(work chan string, wg *sync.WaitGroup) {
	for workUrl := range work {
		downloadUrl(workUrl)
	}

	wg.Done()
}

func downloadUrl(workUrl string) {
	u, err := url.Parse(workUrl)
	if err != nil {
		fmt.Printf("Invalid URL %s, skipping.\n", workUrl)
		return
	}

	parentDir := "./output/"
	ext := path.Ext(u.Path)
	name := sanitizeUrl(u.Scheme + u.Host + trimSuffix(u.Path, ext) + "?" + u.RawQuery)
	filePath := u.Host + "/"
	fileName := filePath + name + ext
	fullPath := parentDir + filePath
	fullName := parentDir + fileName

	err = os.MkdirAll(fullPath, 0755)
	if err != nil {
		fmt.Println("Error creating directory, skipping:", err)
		return
	}

	_, err = os.Stat(fullName)
	if err == nil {
		fmt.Printf("File %s already exists, skipping\n", fileName)
		return
	}

	out, err := os.Create(fullName)
	if err != nil {
		fmt.Println("Error creating file, skipping:", err)
		return
	}
	defer out.Close()

	resp, err := http.Get(workUrl)
	if err != nil {
		fmt.Println("Error getting URL, skipping:", err)
		return
	}
	defer resp.Body.Close()

	bytes, err := io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error writing file, skipping:", err)
		return
	}

	fmt.Printf("%d bytes: %s\n", bytes, fileName)
}

func sanitizeUrl(url string) string {
	return regexp.MustCompile("\\s+").ReplaceAllLiteralString(stripChars(url), "-")
}

func stripChars(str string) string {
	toStrip := [...]string{
		"~", "`", "!", "@", "#", "$", "%", "^", "&", "*",
		"(", ")", "_", "=", "+", "[", "{", "]", "}", "\\",
		"|", ";", ":", "\"", "'", "&#8216;", "&#8217;",
		"&#8220;", "&#8221;", "&#8211;", "&#8212;", "—",
		"–", ",", "<", ".", ">", "/", "?",
	}

	for _, char := range toStrip {
		str = strings.Replace(str, char, "", -1)
	}

	return str
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}

	return s
}

func addWorkFromCsv(work chan string) {
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
			continue
		}

		work <- line[0]
	}
}
