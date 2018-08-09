/*************************************************************************
 * Copyright 2017 Gravwell, Inc. All rights reserved.
 * Contact: <legal@gravwell.io>
 *
 * This software may be modified and distributed under the terms of the
 * BSD 2-clause license. See the LICENSE file for details.
 **************************************************************************/

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cdipaolo/goml/text"
)

var (
	trainData = flag.String("td", "", "Path to trained model file")
)

func main() {
	flag.Parse()
	if *trainData == `` {
		log.Fatal("Missing training data file")
	}
	if len(flag.Args()) < 1 {
		log.Fatal("No sentences provided")
	}
	var nb text.NaiveBayes
	if err := nb.RestoreFromFile(*trainData); err != nil {
		log.Fatal("Failed to import", err)
	}
	for _, v := range flag.Args() {
		t := time.Now()
		p := nb.Predict(v)
		dur := time.Since(t)
		fmt.Printf("%d] \"%s\" %v\n", p, v, dur)
	}
}
