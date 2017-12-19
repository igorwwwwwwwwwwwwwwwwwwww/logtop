package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	ui "github.com/gizak/termui"
	"github.com/igorwwwwwwwwwwwwwwwwwwww/logtop"
)

func consumeStdin(top *logtop.TopNTree, mon *logtop.RateMonitor) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		if err := top.Increment(line); err != nil {
			fmt.Fprintln(os.Stderr, "error: incrementing counter:", err)
			os.Exit(1)
		}
		if err := top.Increment("total"); err != nil {
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

	ls := ui.NewList()
	ls.Items = []string{"waiting..."}
	ls.ItemFgColor = ui.ColorYellow
	ls.BorderLabel = "top n"
	ls.Height = ui.TermHeight()

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, ls),
		),
	)

	ui.Body.Width = ui.TermWidth()

	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/sys/wnd/resize", func(ui.Event) {
		ls.Height = ui.TermHeight()

		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()

		ui.Render(ui.Body)
	})

	ui.Handle("/timer/1s", func(e ui.Event) {
		// update every 2s
		t := e.Data.(ui.EvtTimer)

		if t.Count%2 != 0 {
			return
		}

		n := ui.TermHeight()
		rates := mon.Snapshot()

		strs := []string{}
		for _, tup := range top.TopN(uint64(n)) {
			rate, ok := rates[tup.ID]
			if !ok {
				rate = 0.0
			}
			strs = append(strs, fmt.Sprintf("%d %s (%0.2f/s)\n", tup.Count, tup.ID, rate))
		}

		ls.Items = strs
		ui.Render(ls)
	})

	ui.Loop()
}

func main() {
	top := logtop.NewTopNTree()
	mon := logtop.NewRateMonitor()

	go consumeStdin(top, mon)
	go pruneOld(top)

	// sleepUI(top, mon)
	termUI(top, mon)
}
