package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"generators/utils"
)

func main() {
	defaultPrefixes := []string{
		"COR",
		"DES",
		"CRU",
		"BAT",
		"TTN",
		"JGN",
		"COL",
		"CON",
		"CLN",
		"SCI",
		"TRN",
		"NODE",
		"HUB",
		"CORE",
		"ION",
	}

	count := flag.Int("count", 10, "number of ship names to generate")
	prefixesArg := flag.String("prefixes", "", "comma-separated list of prefixes")
	quotes := flag.Bool("quotes", true, "enable or disable quotes around ship names")
	newlines := flag.Bool("newlines", true, "enable or disable newlines between ship names")
	flag.Parse()

	prefixes := defaultPrefixes
	if *prefixesArg != "" {
		prefixes = strings.Split(*prefixesArg, ",")
	}

	var names []string
	for i := 0; i < *count; i++ {
		f := prefixes[rand.Intn(len(prefixes))]
		hex := utils.MakeRandomHexIdentifier(2, 4)
		name := fmt.Sprintf("%s::%s", f, hex)
		if *quotes {
			names = append(names, fmt.Sprintf("%q", name))
		} else {
			names = append(names, name)
		}
	}

	if *newlines {
		fmt.Println(strings.Join(names, "\n"))
	} else {
		fmt.Println(strings.Join(names, " "))
	}
}
