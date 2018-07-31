/* CC0 - free software.
To the extent possible under law, all copyright and related or neighboring
rights to this work are waived. See the LICENSE file for more information. */

package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/grandchild/base32k"
)

func main() {
	log.SetFlags(0)
	decode := flag.Bool("d", false, "Decode the standard input")
	decodeLong := flag.Bool("decode", false, "Decode the standard input")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	if !scanner.Scan() {
		log.Fatal("error reading stdin")
	}

	if *decode || *decodeLong {
		result, err := base32k.Decode(scanner.Bytes())
		if err != nil {
			log.Fatal(err)
		}
		writer.Write(result)
	} else {
		writer.Write(base32k.Encode(scanner.Bytes()))
	}
	writer.Write([]byte("\x0a"))
	writer.Flush()
	os.Exit(0)
}
