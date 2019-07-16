package transfer

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/stackrox/rox/pkg/concurrency"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	progressBarWidth       = 120
	nonTerminalRefreshRate = 1 * time.Minute
)

type singleCounterDecorator struct {
	decor.WC
	fmt string
}

func newSingleCounterDecorator(fmt string) decor.Decorator {
	wc := decor.WC{}
	wc.Init()
	return &singleCounterDecorator{
		WC:  wc,
		fmt: fmt,
	}
}

func (d *singleCounterDecorator) Decor(st *decor.Statistics) string {
	str := fmt.Sprintf(d.fmt, decor.CounterKiB(st.Current))
	return d.FormatMsg(str)
}

type unknownTotalSizeFiller struct {
	tick int
}

func (f *unknownTotalSizeFiller) Fill(w io.Writer, width int, stat *decor.Statistics) {
	f.tick++

	effectiveWidth := width - 2

	arrowWidth := 5
	arrowSpace := 10
	total := arrowWidth + arrowSpace

	bar := strings.Builder{}
	_, _ = bar.WriteRune('[')
	for i := 0; i < effectiveWidth; i++ {
		if i > f.tick || (f.tick-i)%total >= arrowWidth {
			_, _ = bar.WriteRune('-')
		} else {
			_, _ = bar.WriteRune('>')
		}
	}
	_, _ = bar.WriteRune(']')
	_, _ = w.Write([]byte(bar.String()))
}

func createProgressBars(ctx context.Context, name string, totalSize int64) (*mpb.Bar, func()) {
	outFile := os.Stderr

	opts := []mpb.ProgressOption{
		mpb.WithWidth(progressBarWidth),
		mpb.WithOutput(outFile),
	}

	shutdownSig := concurrency.NewSignal()

	if !terminal.IsTerminal(int(outFile.Fd())) {
		refreshC := make(chan time.Time, 1)
		refreshC <- time.Now() // first tick right away
		shutdownNotifyC := make(chan struct{})
		shutdownC := shutdownSig.Done()

		go func() {
			t := time.NewTicker(nonTerminalRefreshRate)
			defer t.Stop()

			for {
				select {
				case tick := <-t.C:
					refreshC <- tick
				case <-shutdownC:
					shutdownC = nil
					refreshC <- time.Now()
					t = time.NewTicker(100 * time.Millisecond)
					defer t.Stop()
				case <-shutdownNotifyC:
					return
				}
			}
		}()

		opts = append(opts, mpb.WithManualRefresh(refreshC), mpb.WithShutdownNotifier(shutdownNotifyC))
	}

	progressBars := mpb.New(opts...)

	var filler mpb.Filler
	var counterDecorator decor.Decorator
	if totalSize == 0 {
		filler = &unknownTotalSizeFiller{}
		counterDecorator = newSingleCounterDecorator("% 10.1f")
	} else {
		counterDecorator = decor.CountersKibiByte("% 10.1f / % 10.1f")
	}

	appendedDecorators := []decor.Decorator{
		decor.AverageSpeed(decor.UnitKiB, "% 11.1f"),
		decor.Name(fmt.Sprintf(" %s ", name)),
	}
	if totalSize != 0 {
		appendedDecorators = append(appendedDecorators, decor.AverageETA(decor.ET_STYLE_MMSS))
	}

	progressBar := progressBars.Add(totalSize, filler,
		mpb.PrependDecorators(counterDecorator),
		mpb.AppendDecorators(appendedDecorators...),
	)

	shutdownFunc := func() {
		progressBars.Abort(progressBar, false)
		shutdownSig.Signal()
		progressBars.Wait()
	}

	return progressBar, shutdownFunc
}
