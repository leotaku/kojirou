package util

import "github.com/cheggaaa/pb/v3"

const (
	tmpl        = `{{ string . "prefix" | printf "%-10v" }} {{ bar . "|" "█" "▌" " " "|" }} {{ counters . | printf "%-15v" }} {{ "|" }}`
	tmplSpecial = `{{ string . "prefix" | printf "%-10v" }} {{ bar . "|" "█" "▌" " " "|" }} {{ string . "suffix" | printf "%-15v" }} {{ "|" }}`
)

type Bar struct {
	*pb.ProgressBar
}

func NewBar() *Bar {
	pb := pb.New(0).SetTemplate(tmpl).Start()

	return &Bar{
		ProgressBar: pb,
	}
}

func (b *Bar) Message(msg string) *Bar {
	b.Set("prefix", msg)

	return b
}

func (b *Bar) Fail(msg string) *Bar {
	b.SetTemplate(tmplSpecial)
	b.Set("suffix", msg)

	return b
}

func (b *Bar) Succeed(msg string) *Bar {
	b.SetTemplate(tmplSpecial)
	b.Set("suffix", msg)
	b.SetTotal(1).SetCurrent(1)

	return b
}

func (b *Bar) AddTotal(value int64) *Bar {
	b.SetTotal(b.Total() + value)

	return b
}
