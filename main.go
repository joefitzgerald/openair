package main

import (
	"flag"
	"log"

	"github.com/joefitzgerald/openair/generator"
	"github.com/kelseyhightower/envconfig"
)

var (
	objectNames  = flag.String("object", "", "comma-separated list of OpenAir XML Datatype names; must be set")
	outputPrefix = flag.String("prefix", "", "prefix to be added to the output file")
	outputSuffix = flag.String("suffix", "_openair", "suffix to be added to the output file")
)

func main() {
	flag.Parse()
	if len(*objectNames) == 0 {
		log.Fatalf("the flag -object must be set")
	}

	// Only one directory at a time can be processed, and the default is ".".
	dir := "."
	if args := flag.Args(); len(args) == 1 {
		dir = args[0]
	} else if len(args) > 1 {
		log.Fatalf("only one directory at a time")
	}

	var c generator.Config
	err := envconfig.Process("openair", &c)
	if err != nil {
		log.Fatal(err)
	}

	g := generator.New(c, *objectNames, dir, *outputPrefix, *outputSuffix)

	g.GenerateCommonFile()
	g.GenerateCommonTestFile()
	g.GenerateModelFiles()
}
