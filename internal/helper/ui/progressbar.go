package ui

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

const (
	progressBarWidth   = 60
	refreshRateMs      = 200
	etaSmoothingFactor = 30
)

var errFailedToCopyData = errors.New("failed to copy data")

// ProgressBar sets up a progress bar for downloading and copies data from
// src to dst.
func ProgressBar(
	src io.Reader,
	dst io.Writer,
	total int64,
	filename string,
	currentItem int,
	totalItems int,
) error {
	p := mpb.New(
		mpb.WithWidth(progressBarWidth),
		mpb.WithRefreshRate(refreshRateMs*time.Millisecond),
	)

	bar := p.New(total,
		mpb.BarStyle().Rbound("|"),
		mpb.PrependDecorators(
			decor.Name(
				fmt.Sprintf("[%d/%d] %s ", currentItem, totalItems, filepath.Base(filename)),
			),
			decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.EwmaETA(decor.ET_STYLE_GO, etaSmoothingFactor),
			decor.Name(" ] "),
			decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", etaSmoothingFactor),
		),
	)

	proxyReader := bar.ProxyReader(src)

	defer func() {
		if err := proxyReader.Close(); err != nil {
			fmt.Printf("error waiting for progress bar: %v\n", err)
		}
	}()

	start := time.Now()

	if _, err := io.Copy(dst, proxyReader); err != nil {
		return fmt.Errorf("%w: %w", errFailedToCopyData, err)
	}

	bar.EwmaIncrInt64(total, time.Since(start))

	p.Wait()

	return nil
}
