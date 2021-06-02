package progressbar

import "fmt"

const PercentMultiplier = 50

type Bar struct {
	percent int    // progress percentage
	cur     int    // current progress
	total   int    // total value for progress
	rate    string // the actual progress bar to be printed
	graph   string // the fill value for progress bar
}

func (bar *Bar) NewOption(start, total int) {
	bar.cur = start
	bar.total = total

	if bar.graph == "" {
		bar.graph = "#"
	}

	bar.percent = bar.getPercent()

	for i := 0; i < bar.percent; i++ {
		bar.rate += bar.graph // initial progress position
	}
}

func (bar *Bar) getPercent() int {
	return int((float32(bar.cur) / float32(bar.total)) * PercentMultiplier)
}

func (bar *Bar) Play(cur int) {
	bar.cur = cur
	last := bar.percent // nolint // needs to be set before conditional for progressbar
	bar.percent = bar.getPercent()

	if bar.percent != last {
		var i int
		for ; i < bar.percent-last; i++ {
			bar.rate += bar.graph
		}

		fmt.Printf("\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent*2, bar.cur, bar.total) //nolint // output for progressbar
	}
}

func (bar *Bar) Finish() {
	fmt.Println("") // nolint // output of last blank line when finished progress
}
