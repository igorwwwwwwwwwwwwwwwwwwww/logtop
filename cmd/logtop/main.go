package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/gops/agent"
	"github.com/igorwwwwwwwwwwwwwwwwwwww/logtop"
)

func consumeStdin(ch chan<- string, done chan<- interface{}) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		ch <- scanner.Text()
	}
	close(ch)
	close(done)

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error: reading standard input:", err)
		os.Exit(1)
	}
}

func termUI(top *logtop.TopNTree, mon *logtop.RateMonitor, ch <-chan string, done <-chan interface{}) {
	err := ui.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer ui.Close()

	width, height := ui.TerminalDimensions()

	l := widgets.NewList()
	l.Title = "top k"
	l.Rows = []string{"waiting..."}
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(0, 0, width, height)

	ui.Render(l)

	pruneIntervalSeconds := 30 * time.Second

	events := ui.PollEvents()
	ticker := time.NewTicker(time.Millisecond * 1000 / 60)
	rateTicker := time.NewTicker(1 * time.Second)
	pruneTicker := time.NewTicker(pruneIntervalSeconds)

	rates := mon.Snapshot()

	for {
		select {
		case line, ok := <-ch:
			if !ok {
				continue
			}

			top.Increment(line, time.Now())
			top.Increment("total", time.Now())

			mon.Record(line)
			mon.Record("total")
		case e := <-events:
			if e.Type == ui.KeyboardEvent && e.ID == "q" {
				return
			}
			if e.Type == ui.KeyboardEvent && e.ID == "<C-c>" {
				return
			}
			if e.ID == "<Resize>" {
				payload := e.Payload.(ui.Resize)
				width, height = payload.Width, payload.Height
				l.SetRect(0, 0, width, height)
				ui.Render(l)
				continue
			}
			// fmt.Printf("%+v\n", e)
		case <-ticker.C:
			strs := []string{}
			for _, tup := range top.TopN(uint64(height)) {
				rate, ok := rates[tup.ID]
				if !ok {
					rate = 0.0
				}
				strs = append(strs, fmt.Sprintf("%d %s (%0.2f/s)", tup.Count, tup.ID, rate))
			}

			l.Rows = strs
			ui.Render(l)
		case <-done:
			ticker.Stop()
			pruneTicker.Stop()

			strs := []string{}
			for _, tup := range top.TopN(uint64(height)) {
				strs = append(strs, fmt.Sprintf("%d %s", tup.Count, tup.ID))
			}

			l.Rows = strs
			ui.Render(l)
		case <-rateTicker.C:
			rates = mon.Snapshot()
		case <-pruneTicker.C:
			top.PruneBefore(time.Now().Add(-pruneIntervalSeconds))
		}
	}
}

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	ch := make(chan string, 1024)
	done := make(chan interface{})
	go consumeStdin(ch, done)

	top := logtop.NewTopNTree()
	mon := logtop.NewRateMonitor()

	termUI(top, mon, ch, done)
}
