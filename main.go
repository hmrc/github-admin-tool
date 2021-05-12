package main

import (
	report "github-admin-tool/cmd"
	"github-admin-tool/loadconfig"
)

func main() {
	var cfg loadconfig.Config
	loadconfig.LoadConfig(&cfg)
	report.Run(cfg)
}
