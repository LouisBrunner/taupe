package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/LouisBrunner/taupe"
)

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] url\n", os.Args[0])
	flag.PrintDefaults()
}

type args struct {
	address string
}

func parseArgs() *args {
	requiredArgs := 1

	flag.Parse()

	if flag.NArg() != requiredArgs {
		flag.Usage()
		os.Exit(1)
	}

	return &args{address: flag.Arg(0)}
}

func main() {
	flag.Usage = usage
	args := parseArgs()

	app := taupe.NewApplication()
	app.Run(args.address)
}
