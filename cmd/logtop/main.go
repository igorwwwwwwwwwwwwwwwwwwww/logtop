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

func consumeStdin(top *logtop.TopNTree, mon *logtop.RateMonitor) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		if err := top.Increment(line, time.Now()); err != nil {
			fmt.Fprintln(os.Stderr, "error: incrementing counter:", err)
			os.Exit(1)
		}
		if err := top.Increment("total", time.Now()); err != nil {
			fmt.Fprintln(os.Stderr, "error: incrementing counter:", err)
			os.Exit(1)
		}

		mon.Record(line)
		mon.Record("total")
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error: reading standard input:", err)
		os.Exit(1)
	}
}

func pruneOld(top *logtop.TopNTree) {
	pruneIntervalSeconds := 30 * time.Second

	time.Sleep(pruneIntervalSeconds)
	for {
		top.PruneBefore(time.Now().Add(-pruneIntervalSeconds))
		time.Sleep(pruneIntervalSeconds)
	}
}

func sleepUI(top *logtop.TopNTree, mon *logtop.RateMonitor) {
	for {
		time.Sleep(1 * time.Second)

		rates := mon.Snapshot()

		for _, tup := range top.TopN(6) {
			rateStr := ""
			if rate, ok := rates[tup.ID]; ok {
				rateStr = fmt.Sprintf("(%0.2f/s)", rate)
			}
			fmt.Println(tup.Count, tup.ID, rateStr)
		}
		fmt.Println()
	}

	for _, tup := range top.TopN(5) {
		fmt.Println(tup.Count, tup.ID)
	}
}

func termUI(top *logtop.TopNTree, mon *logtop.RateMonitor) {
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

	events := ui.PollEvents()
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
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
			rates := mon.Snapshot()

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
		}
	}
}

func debugUI(top *logtop.TopNTree) {
	var err error
	if err = top.Increment("foo", time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "error: reading standard input:", err)
	}
	if err = top.Increment("foo", time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "error: reading standard input:", err)
	}
	for _, tup := range top.TopN(5) {
		fmt.Println(tup.Count, tup.ID)
	}
	os.Exit(32)
}

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	top := logtop.NewTopNTree()
	mon := logtop.NewRateMonitor()

	// debugUI(top)

	go consumeStdin(top, mon)
	go pruneOld(top)

	// sleepUI(top, mon)
	termUI(top, mon)
}
