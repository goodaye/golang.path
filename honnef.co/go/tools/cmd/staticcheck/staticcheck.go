// staticcheck analyses Go code and makes it better.
package main

import (
	"log"
	"os"

	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/lintcmd"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"honnef.co/go/tools/unused"
)

func main() {
	fs := lintcmd.FlagSet("staticcheck")
	debug := fs.String("debug.unused-graph", "", "Write unused's object graph to `file`")
	qf := fs.Bool("debug.run-quickfix-analyzers", false, "Run quickfix analyzers")
	fs.Parse(os.Args[1:])

	var cs []*lint.Analyzer
	for _, v := range simple.Analyzers {
		cs = append(cs, v)
	}
	for _, v := range staticcheck.Analyzers {
		cs = append(cs, v)
	}
	for _, v := range stylecheck.Analyzers {
		cs = append(cs, v)
	}

	if *qf {
		for _, v := range quickfix.Analyzers {
			cs = append(cs, v)
		}
	}

	cs = append(cs, unused.Analyzer)
	if *debug != "" {
		f, err := os.OpenFile(*debug, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatal(err)
		}
		unused.Debug = f
	}

	lintcmd.ProcessFlagSet(cs, fs)
}
