package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/anaseto/gruid"
)

// Width is the maximum width of the UI.
const Width = 80

// These variables contain the values provided by options on the command line.
var (
	OptTextWidth int
	OptWords     int
	OptTTF       string
)

func main() {
	opttw := flag.Int("c", 30, "maximum number of characters per line")
	optw := flag.Int("w", 2, "maximum number of words per line")
	var optttf *string
	if TTF {
		optttf = flag.String("f", "", "opentype font file with custom font")
	}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s file\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	OptTextWidth = *opttw
	OptWords = *optw
	if OptTextWidth > Width-2 {
		OptTextWidth = Width - 2
		fmt.Fprint(os.Stderr, "gospeedr:argument to -c is too big")
	}
	if OptWords < 1 {
		OptWords = 1
		fmt.Fprint(os.Stderr, "gospeedr:argument to -w should be at least 1")
	}
	if optttf != nil {
		OptTTF = *optttf
	}

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "gospeedr:filename required\n")
		flag.Usage()
		os.Exit(1)
	}
	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "gospeedr:%s\n", err)
		os.Exit(1)
	}
	words := strings.FieldsFunc(string(data), func(r rune) bool {
		return unicode.IsSpace(r) && r != ' '
	})
	if len(words) <= 0 {
		fmt.Fprintf(os.Stderr, "gospeedr:file %s has no words\n", flag.Arg(0))
		os.Exit(1)
	}

	initDriver()
	app := gruid.NewApp(gruid.AppConfig{
		Driver: driver,
		Model:  newModel(words),
	})
	if err := app.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
}
