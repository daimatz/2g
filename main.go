package main

import (
	"bufio"
	"fmt"
	"os"
)

func fetch(dat chan string, done chan bool) {
	for {
		d, more := <-dat
		if more && d != "" {
			fmt.Println(d)
		} else {
			done <- true
		}
	}
}

func main() {
	dat := make(chan string)
	done := make(chan bool)

	go fetch(dat, done)

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
