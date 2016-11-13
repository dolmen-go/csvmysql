package main

import (
	"bufio"
	"compress/bzip2"
	"io"
	//"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/dolmen-go/csvmysql"
)

func filterFile(filename string, out io.Writer) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	r := io.Reader(f)
	if strings.HasSuffix(filename, ".bz2") {
		bufR := bufio.NewReader(bzip2.NewReader(r))
		// Verify that the header is correct
		_, err := bufR.Peek(50)
		if err != nil && err != io.EOF {
			return err
		}
		r = bufR
	}

	r = csvmysql.NewUnescapeReader(r)

	_, err = io.Copy(out, r)
	return err
}

func main() {
	file := os.Args[1]

	err := filterFile(file, os.Stdout)
	//err := filterFile(file, ioutil.Discard)
	if err != nil {
		log.Fatal(err)
	}
}
