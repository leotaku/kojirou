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

type Reporter interface {
	Increase(int)
	Add(int)
	NewProxyWriter(io.Writer) io.Writer
}

type BarReporter struct {
	bar       *pb.ProgressBar
	firstCall bool
}

func (r BarReporter) Increase(n int) {
	r.bar.AddTotal(int64(n))
}

func (r BarReporter) Add(n int) {
	r.bar.Add(n)
}

func (r BarReporter) NewProxyWriter(w io.Writer) io.Writer  {
	return r.bar.NewProxyWriter(w)
}

func (r BarReporter) Done() {
	r.bar.Finish()
}

func (r *BarReporter) Cancel(message string) {
	r.bar.Set("message", message)
	r.bar.SetTotal(1).SetCurrent(1)
	r.Done()
}

func TitledProgress(title string) BarReporter {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Start()

	return BarReporter{bar, true}
}

func VanishingProgress(title string) BarReporter {
	bar := pb.New(0).SetTemplate(progressTemplate)
	bar.Set("prefix", title)
	bar.Set(pb.CleanOnFinish, true)
	bar.Start()

	return BarReporter{bar, true}
}
