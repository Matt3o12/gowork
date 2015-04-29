package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

var formatDebug = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{module}.%{shortfunc} ▶ %{level}" +
		" %{id:03x}%{color:reset} \"%{message}\" \033[0;37mat %{longpkg}/%{shortfile}%{color:reset}",
)
var formatDefault = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{module}.%{shortfunc} ▶ %{level}" +
		" %{id:03x}%{color:reset} %{message}",
)

func isVerbose() bool {
	args := os.Args
	for _, theArg := range args {
		if theArg == "-v" || theArg == "--verbose" {
			return true
		}
	}

	return false
}

func setupLogging() {
	var backend logging.Backend = logging.NewLogBackend(os.Stderr, "", 0)
	level := logging.WARNING
	format := formatDefault
	if isVerbose() {
		format = formatDebug
		level = logging.DEBUG
	}

	backend = logging.NewBackendFormatter(backend, format)
	leveled := logging.AddModuleLevel(backend)
	leveled.SetLevel(level, "")

	logging.SetBackend(leveled)
}

func main() {
	app := cli.NewApp()
	app.Author = "Matteo Kloiber"
	app.Name = "Gowork Wrapper"
	app.HideVersion = true // TODO: Show version when available.

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Turn on verbose logging.",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "workon",
			Action: workon,
			Usage:  "CD into project to work on it.",
			Flags:  app.Flags,
		},
	}

	app.EnableBashCompletion = true

	setupLogging()
	app.RunAndExitOnError()
}
