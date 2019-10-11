// +build ignore

package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/vrischmann/logfmt"
)

func dictionaryWordGenerator() <-chan string {
	f, err := os.Open("/usr/share/dict/words")
	if err != nil {
		log.Fatalf("can't find words. if you're on a Debian derivative install wfrench. err=%v", err)
	}

	words := make([]string, 0, 65536)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := scanner.Text()
		if strings.Contains(word, `'`) {
			continue
		}

		word = strings.TrimSpace(word)
		word = strings.ToLower(word)

		words = append(words, word)
	}

	ch := make(chan string)

	go func(words []string) {
		for {
			pos := rand.Intn(len(words))
			ch <- words[pos]
		}
	}(words)

	return ch
}

func makeOutputFile(filename string) *os.File {
	if filename != "" {
		f, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		return f
	}

	f, err := ioutil.TempFile("", "logfmt")
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	var (
		flSize   = flag.Int("size", 1024, "Size of the test file in megabytes")
		flOutput = flag.String("output", "", "Output file")
	)

	flag.Parse()

	maxSize := *flSize * 1024 * 1024

	//

	output := makeOutputFile(*flOutput)
	defer output.Close()

	log.Printf("output at %s", output.Name())

	//

	wordGenerator := dictionaryWordGenerator()

	var (
		buf          = make([]byte, 0, 4096)
		totalWritten int
	)
	for totalWritten < maxSize {
		nbPairs := rand.Intn(30)
		if nbPairs < 4 {
			nbPairs = 4
		}

		pairs := make(logfmt.Pairs, nbPairs)
		for i := 0; i < nbPairs; i++ {
			pairs[i].Key = <-wordGenerator
			pairs[i].Value = <-wordGenerator
		}

		buf = pairs.AppendFormat(buf)
		buf = append(buf, '\n')

		n, err := output.Write(buf)
		if err != nil {
			log.Fatal(err)
		}
		totalWritten += n

		buf = buf[:0]
	}
}
