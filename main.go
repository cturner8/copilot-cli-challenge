/*
Copyright © 2026 cturner8
*/
package main

import "cturner8/binmate/cmd"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd.SetBuildMetadata(version, commit, date)
	cmd.Execute()
}
