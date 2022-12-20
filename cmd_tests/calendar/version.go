package main

import (
	"fmt"
)

var (
	release   = "1"
	buildDate = "1"
	gitHash   = "1"
)

type Version struct {
	Release   string
	BuildDate string
	GitHash   string
}

func (v *Version) String() string {
	return fmt.Sprintf(
		"Release: %q;\nBuildDate: %q;\nGitHash: %q;\n",
		release,
		buildDate,
		gitHash,
	)
}

func printVersion() {
	version := Version{
		Release:   release,
		BuildDate: buildDate,
		GitHash:   gitHash,
	}
	fmt.Print(version.String())
}
