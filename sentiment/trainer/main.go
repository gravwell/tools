/*************************************************************************
 * Copyright 2017 Gravwell, Inc. All rights reserved.
 * Contact: <legal@gravwell.io>
 *
 * This software may be modified and distributed under the terms of the
 * BSD 2-clause license. See the LICENSE file for details.
 **************************************************************************/

package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/text"
)

const (
	Pos uint8 = 1
	Neg uint8 = 0
)

var (
	posDir = flag.String("pos", "", "Postive sentiment dataset")
	negDir = flag.String("neg", "", "Postive sentiment dataset")
	output = flag.String("output", "", "Output file for trained model")
)

func main() {
	flag.Parse()
	if err := checkTrainingDir(*posDir); err != nil {
		log.Fatal("Positive training dir failure: ", err)
	}
	if err := checkTrainingDir(*negDir); err != nil {
		log.Fatal("Negative training dir failure: ", err)
	}
	if *output == `` {
		log.Fatal("No output file specified")
	}

	nb, err := Train(*posDir, *negDir)
	if err != nil {
		log.Fatal("Failed to train: ", err)
	}

	if err := nb.PersistToFile(*output); err != nil {
		log.Fatal("Failed to save model", err)
	}
}

func checkTrainingDir(p string) (err error) {
	var fi os.FileInfo
	if p == `` {
		return errors.New("Empty path")
	}
	if fi, err = os.Stat(p); err != nil {
		return
	} else if !fi.Mode().IsDir() {
		err = errors.New("Not a directory")
	}
	return
}

func Train(pos, neg string) (model *text.NaiveBayes, err error) {
	//check the positive data set
	if pos, err = filepath.Abs(pos); err != nil {
		return
	}
	//check the negative data set
	if neg, err = filepath.Abs(neg); err != nil {
		return
	}
	var ct int
	var files int

	stream := make(chan base.TextDatapoint, 1024)
	errCh := make(chan error, 128)
	model = text.NewNaiveBayes(stream, 2, base.OnlyWords)

	go model.OnlineLearn(errCh)

	walk := func(dirpath string, class uint8) error {
		return filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			files++

			// read file
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			scanner.Split(bufio.ScanWords)

			for scanner.Scan() {
				line := strings.TrimSuffix(scanner.Text(), "\n")
				if len(line) == 0 {
					continue
				}

				ct++
				if ct%500 == 0 {
					print(".")
				}

				stream <- base.TextDatapoint{
					X: line,
					Y: class,
				}
			}

			return nil
		})
	}

	if err = walk(pos, Pos); err != nil {
		return
	}
	if err = walk(neg, Neg); err != nil {
		return
	}
	close(stream)

	for lerr := range errCh {
		if lerr != nil {
			err = lerr
		}
	}
	log.Printf("Handled %d words over %d files", ct, files)
	return
}
