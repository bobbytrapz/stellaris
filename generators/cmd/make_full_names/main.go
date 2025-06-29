package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"

	"generators/utils"
)

func main() {
	defaultFirst := []string{
		"SYS",
		"CORE",
		"MEM",
		"NET",
		"SEC",
	}
	defaultSecond := []string{
		"Assert",
		"Cleanse",
		"Filter",
		"Purge",
		"Resolve",
		"Sanitize",
		"Truncate",
		"Validate",
		"Commit",
		"Index",
		"Iterate",
		"Parse",
		"Quantize",
		"Rebase",
		"Sync",
		"Subtask",
	}

	count := flag.Int("count", 10, "number of names to generate")
	firstPrefixes := flag.String("first", strings.Join(defaultFirst, ","), "comma-separated list of first prefixes")
	secondPrefixes := flag.String("second", strings.Join(defaultSecond, ","), "comma-separated list of second prefixes")
	quotes := flag.Bool("quotes", true, "enable quotes around generated names")
	newlines := flag.Bool("newlines", true, "enable newline after each name")
	flag.Parse()

	first := defaultFirst
	if *firstPrefixes != strings.Join(defaultFirst, ",") {
		first = strings.Split(*firstPrefixes, ",")
	}

	second := defaultSecond
	if *secondPrefixes != strings.Join(defaultSecond, ",") {
		second = strings.Split(*secondPrefixes, ",")
	}

	var names []string
	for i := 0; i < *count; i++ {
		f := first[rand.Intn(len(first))]
		s := second[rand.Intn(len(second))]
		hex := utils.MakeRandomHexIdentifier(2, 4)

		name := fmt.Sprintf("%s::%s:%s", f, s, hex)

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
