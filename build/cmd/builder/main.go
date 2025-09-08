package main

import (
	"github.com/outofforest/build/v2"
	"github.com/outofforest/build/v2/pkg/tools/git"
	"github.com/outofforest/tools"
	"github.com/outofforest/tools/pkg/tools/golang"
)

func main() {
	build.RegisterCommands(
		build.Commands,
		git.Commands,
		golang.Commands,
	)
	tools.Main()
}
