package progress

import (
	"github.com/jedib0t/go-pretty/v6/progress"
	"time"
)

type Progress struct {
	header string
	pw     progress.Writer
	t      *progress.Tracker
}

func NewProgress(header string) *Progress {
	return &Progress{
		header: header,
		pw:     progress.NewWriter(),
	}
}

func (p *Progress) Start(total int) {
	tracker := progress.Tracker{
		Message: p.header,
		Total:   int64(total),
		Units:   progress.UnitsDefault,
	}

	p.t = &tracker
	p.pw.SetMessageLength(50)
	p.pw.SetSortBy(progress.SortByPercentDsc)
	p.pw.SetStyle(progress.StyleDefault)
	p.pw.SetTrackerLength(50)
	p.pw.SetTrackerPosition(progress.PositionRight)
	p.pw.AppendTracker(p.t)

	go p.pw.Render()

	// adding small delay to make sure the progress bar is rendered
	time.Sleep(100 * time.Millisecond)
}

func (p *Progress) Increment() {
	p.t.Increment(1)
}

func (p *Progress) Finish() {
	p.t.MarkAsDone()
}
