package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"code.google.com/p/go.text/encoding/japanese"
	"code.google.com/p/go.text/transform"
)

// FIXME: brash up regexp
var ImgUrlRegexp *regexp.Regexp = regexp.MustCompile("h?ttp://[0-9a-zA-Z/\\-.%]+?\\.(jpg|jpeg|gif|png)")

var Fetched = struct {
	sync.RWMutex
	m map[string]bool
}{m: make(map[string]bool)}

func Img(url string) {
	Fetched.Lock()
	if Fetched.m[url] {
		Fetched.Unlock()
		return
	}
	Fetched.m[url] = true
	Fetched.Unlock()
	log.Printf("downloading %v...\n", url)

	out, err := os.Create(filepath.Base(url))
	if err != nil {
		log.Printf("failed to create download file: %v\n", err)
		return
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("failed to download img: %v, err: \n", url, err)
		return
	}
	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("failed to copy img: %v\n", err)
		return
	}
	log.Printf("saved %v (%v bytes)\n", url, n)
}

func ImgQueue(ch chan string, done chan bool) {
	for {
		url, more := <-ch
		if more && url != "" {
			Img(url)
		} else {
			done <- true
		}
	}
}

func Dat(url string) {
	Fetched.Lock()
	if Fetched.m[url] {
		Fetched.Unlock()
		return
	}
	Fetched.m[url] = true
	Fetched.Unlock()
	log.Printf("reading %v...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("failed to fetch: %v, err: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	imgCh := make(chan string)
	done := make(chan bool)
	go ImgQueue(imgCh, done)

	reader := transform.NewReader(resp.Body, japanese.ShiftJIS.NewDecoder())
	bufr := bufio.NewReader(reader)
	for {
		lineb, err := bufr.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("failed to read content. err: %v\n", err)
			return
		}
		line := strings.Split(string(lineb), "<>")
		if len(line) >= 4 {
			body := line[3]
			matched := ImgUrlRegexp.FindAllString(body, -1)
			for i := 0; i < len(matched); i++ {
				imgCh <- matched[i]
			}
		}
	}
	close(imgCh)

	<-done
}

func DatQueue(ch chan string, done chan bool) {
	for {
		url, more := <-ch
		if more && url != "" {
			Dat(url)
		} else {
			done <- true
		}
	}
}

func main() {
	datCh := make(chan string)
	done := make(chan bool)

	go DatQueue(datCh, done)

	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			datCh <- os.Args[i]
		}
	} else {
		in := bufio.NewReader(os.Stdin)
		for {
			input, err := in.ReadString('\n')
			if err != nil {
				break
			}
			datCh <- input
		}
	}
	close(datCh)

	<-done

	fmt.Println("OK")
}
