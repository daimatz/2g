package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"code.google.com/p/go.text/encoding/japanese"
	"code.google.com/p/go.text/transform"
)

func Dat(url string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("failed to fetch: %v, err: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	reader := transform.NewReader(resp.Body, japanese.ShiftJIS.NewDecoder())
	bufr := bufio.NewReader(reader)
	for {
		lineb, err := bufr.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("failed to read content. err: %v\n", err)
			return
		}
		line := string(lineb)
		fmt.Printf("line: %v\n", string(line))
	}
}

func DatQueue(dat chan string, done chan bool) {
	for {
		url, more := <-dat
		if more && url != "" {
			Dat(url)
		} else {
			done <- true
		}
	}
}

func main() {
	dat := make(chan string)
	done := make(chan bool)

	go DatQueue(dat, done)

	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			dat <- os.Args[i]
		}
	} else {
		in := bufio.NewReader(os.Stdin)
		for {
			input, err := in.ReadString('\n')
			if err != nil {
				break
			}
			dat <- input
		}
	}
	close(dat)

	<-done
}
