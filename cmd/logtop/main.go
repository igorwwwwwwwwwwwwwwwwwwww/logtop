package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/igorwwwwwwwwwwwwwwwwwwww/logtop"

	_ "net/http/pprof"
)

func main() {
	top := logtop.NewTopNTree()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			for _, e := range top.TopN(5) {
				fmt.Println(e.Count, e.Line)
			}
			fmt.Println()
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if err := top.Increment(line); err != nil {
			fmt.Fprintln(os.Stderr, "incrementing counter:", err)
			os.Exit(1)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}

	for _, e := range top.TopN(5) {
		fmt.Println(e.Count, e.Line)
	}
}
