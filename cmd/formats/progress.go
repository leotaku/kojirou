package formats

import (
	"io"

	"github.com/cheggaaa/pb/v3"
)

const (
	progressTemplate = `` +
		`{{ string . "prefix" | printf "%-10v" }}` +
		`{{ bar . "|" "█" "▌" " " "|" }}` + `{{ " " }}` +
		`{{ if string . "message" }}` +
		`{{   string . "message" | printf "%-15v" }}` +
		`{{ else }}` +
		`{{   counters . | printf "%-15v" }}` +
		`{{ end }}` + `{{ " |" }}`
)

type Progress interface {
	Increase(int)
	Add(int)
	NewProxyWriter(io.Writer) io.Writer
}

type CliProgress struct {
	bar       *pb.ProgressBar
	firstCall bool
}

func (p CliProgress) Increase(n int) {
	p.bar.AddTotal(int64(n))
}

func (p CliProgress) Add(n int) {
	p.bar.Add(n)
}

func (p CliProgress) NewProxyWriter(w io.Writer) io.Writer {
	return p.bar.NewProxyWriter(w)
}

func (p CliProgress) Done() {
	p.bar.Finish()
}

func (p *CliProgress) Cancel(message string) {
	p.bar.Set("message", message)
	p.bar.SetTotal(1).SetCurrent(1)
	p.Done()
}

func TitledProgress(title string) CliProgress {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Start()

	return CliProgress{bar, true}
}

func VanishingProgress(title string) CliProgress {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()

	return CliProgress{bar, true}
}
