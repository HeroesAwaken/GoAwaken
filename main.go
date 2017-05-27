package main

import (
	"flag"

	gs "github.com/ReviveNetwork/GoRevive/GameSpy"
	log "github.com/ReviveNetwork/GoRevive/Log"
)

var (
	// BuildTime of the build provided by the build command
	BuildTime = "Not provided"
	// GitHash of build provided by the build command
	GitHash = "Not provided"
	// GitBranch of the build provided by the build command
	GitBranch = "Not provided"
	// compileVersion we are receiving by the build command
	CompileVersion = "0"
	// Version of the Application
	Version = "0.0.0"
)

func main() {
	var (
		logLevel = flag.String("loglevel", "error", "LogLevel [error|warning|note|debug]")
	)
	flag.Parse()

	if CompileVersion != "0" {
		Version = Version + "." + CompileVersion
	}

	log.SetLevel(*logLevel)
	log.Notef("Starting up v%s\n", Version)

	// Startup done

	test := gs.ShortHash("Bla")

	log.Noteln(test)
}
