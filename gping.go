package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"

	"github.com/go-ping/ping"
)

var host string

func init() {
	flag.StringVar(&host, "host", "", "to ping remote host")

}

func main() {
	flag.Parse()
	if flag.NFlag() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	pinger, err := ping.NewPinger(host)
	if err != nil {
		panic(err)
	}

	term, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer term.Close()

	length := 30
	data := make([]float64, length)

	const redrawInterval = 250 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	var lc *linechart.LineChart

	if lc, err = linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	); err != nil {
		panic(err)
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		i := pkt.Seq % length
		data[i] = pkt.Rtt.Seconds()
		rotated := append(data[i:], data[:i]...)
		lc.Series("first", rotated)
	}

	go pinger.Run()

	c, err := container.New(
		term,
		container.Border(linestyle.Round),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(lc),
	)

	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
			pinger.Stop()
		}
	}

	if err := termdash.Run(ctx, term, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}
