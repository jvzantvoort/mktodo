package main

import (
	"os"

	"github.com/jvzantvoort/mktodo/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	os.Exit(cmd.Execute())
}
