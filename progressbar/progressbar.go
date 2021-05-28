package progressbar

import "fmt"

const PercentMultiplier = 50

type Bar struct {
	percent int64  // progress percentage
	cur     int64  // current progress
	total   int64  // total value for progress
	rate    string // the actual progress bar to be printed
	graph   string // the fill value for progress bar
}

func (bar *Bar) NewOption(start, total int64) {
	bar.cur = start
	bar.total = total

	if bar.graph == "" {
		bar.graph = "#"
	}

	bar.percent = bar.getPercent()

	for i := 0; i < int(bar.percent); i++ {
		bar.rate += bar.graph // initial progress position
	}
}

func (bar *Bar) getPercent() int64 {
	return int64((float32(bar.cur) / float32(bar.total)) * PercentMultiplier)
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()

	if bar.percent != last {
		var i int64
		for ; i < bar.percent-last; i++ {
			bar.rate += bar.graph
		}
		fmt.Printf("\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent*2, bar.cur, bar.total)
	}
}

func (bar *Bar) Finish() {
	fmt.Println("")
}
