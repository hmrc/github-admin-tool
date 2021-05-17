package csv

import "testing"

func TestRun(t *testing.T) {
	type args struct {
		parsed [][]string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "1 row", args: args{parsed: [][]string{{"hello", "hello2"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Run(tt.args.parsed)
		})
	}
}
