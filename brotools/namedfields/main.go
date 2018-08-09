/*************************************************************************
 * Copyright 2017 Gravwell, Inc. All rights reserved.
 * Contact: <legal@gravwell.io>
 *
 * This software may be modified and distributed under the terms of the
 * BSD 2-clause license. See the LICENSE file for details.
 **************************************************************************/

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	defDelim = "\t"
	version  = 1
)

var (
	ErrNotFound = errors.New("not found")

	ofile   = flag.String("o", "/tmp/namedfields.json", "Output file")
	verbose = flag.Bool("v", false, "Verbose")
)

func init() {
	flag.Parse()
	if *ofile == `` {
		log.Fatal("Output file required")
	}
}

type Type struct {
	Name     string
	DataType string `json:",omitempty"`
	Index    int
}

type Group struct {
	Delim string
	Name string
	Subs  []Type
}

type Resource struct {
	Version int `json:",omitempty"`
	Set     []Group
}

func main() {
	if len(flag.Args()) != 1 {
		log.Fatal("I need an input folder")
	}
	var bset []Group

	err := filepath.Walk(flag.Args()[0], func(p string, info os.FileInfo, err error) error {
		var bs Group
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if filepath.Base(p) != `main.bro` {
			logf("Skipping non main %s\n", p)
			return nil
		}
		if bs, err = processFile(p); err == nil {
			bset = append(bset, bs)
			logf("Extracted %d names from %s\n", len(bs.Subs), p)
		} else {
			logln("Failed to process", p, err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	r := Resource{
		Version: version,
		Set:     bset,
	}
	bts, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		log.Fatal("failed to generate resource", err)
	}
	if err := ioutil.WriteFile(*ofile, bts, 0660); err != nil {
		log.Fatal("Failed to write resource file", err)
	}
}

func processFile(f string) (br Group, err error) {
	var bts []byte
	if bts, err = ioutil.ReadFile(f); err != nil {
		return
	}
	var module string
	if module, err = getModule(bts); err != nil {
		err = fmt.Errorf("Failed to get module %s", err)
		return
	}
	var tr []byte
	if tr, _, err = getTypeRecord(bts); err != nil {
		err = fmt.Errorf("Failed to get type record %s", err)
		return
	}
	lines := getLines(tr)

	if br.Subs, err = getTypes(lines); err != nil {
		err = fmt.Errorf("Failed to get type record %s", err)
		return
	}
	br.Name = fmt.Sprintf("%s", module)
	br.Delim = defDelim
	return
}

func getModule(bts []byte) (m string, err error) {
	if bts, err = getRegexField(bts, `module\s([a-zA-Z_0-9]+);\n`); err != nil {
		return
	}
	m = string(bts)
	return
}

func getExport(bts []byte) (v []byte, err error) {
	v, err = getRegexField(bts, `\nexport\s\{(.+)\n}`)
	return
}

func getTypeRecord(bts []byte) (v []byte, name string, err error) {
	var x []byte
	if x, err = getRegexField(bts, `type\s([a-zA-Z]+)\:\srecord\s\{`); err != nil {
		return
	}
	name = string(x)
	v, err = getRegexField(bts, `type\s[a-zA-Z]+\:\srecord\s\{([^\}]+)\};`)
	return
}

func getLines(bts []byte) (v []string) {
	vv := bytes.Split(bts, []byte("\n"))
	for i := range vv {
		vv[i] = bytes.TrimSpace(vv[i])
		if bytes.HasPrefix(vv[i], []byte("#")) || len(vv[i]) == 0 {
			continue
		}
		v = append(v, string(vv[i]))
	}
	return
}

func getField(bts, s, e []byte) (v []byte, vidx int, err error) {
	var x []byte
	var idx int
	if vidx = bytes.Index(bts, s); vidx < 0 {
		err = ErrNotFound
		return
	}
	vidx += len(s)
	x = bts[vidx:]
	if idx = bytes.Index(x, e); idx < 0 {
		err = ErrNotFound
		return
	}
	v = x[:idx]
	return
}

func getRegexField(bts []byte, r string) (v []byte, err error) {
	rxp := regexp.MustCompile(r)
	subs := rxp.FindSubmatch(bts)
	if len(subs) <= 1 {
		err = ErrNotFound
	} else {
		v = subs[1]
	}
	return
}

func getTypes(lines []string) (tps []Type, err error) {
	var tp Type
	var idx int
	for _, v := range lines {
		if tp, err = getType(v); err != nil {
			return
		}
		if tp.Name == `id` && tp.DataType == `conn_id` {
			tps = append(tps, idSet(idx)...)
			idx += 4
		} else {
			tp.Index = idx
			idx++
			tps = append(tps, tp)
		}
	}
	return
}

func getType(l string) (tp Type, err error) {
	if flds := strings.Fields(l); len(flds) < 2 {
		err = ErrNotFound
		return
	} else {
		tp.Name = strings.TrimRight(flds[0], ":")
		tp.DataType = flds[1]
	}
	return
}

func idSet(idx int) (tp []Type) {
	tp = []Type{
		{
			Name:     `src`,
			DataType: `ip`,
			Index:    idx,
		},
		{
			Name:     `src_port`,
			DataType: `port`,
			Index:    idx + 1,
		},
		{
			Name:     `dst`,
			DataType: `ip`,
			Index:    idx + 2,
		},
		{
			Name:     `dst_port`,
			DataType: `port`,
			Index:    idx + 3,
		},
	}
	return
}

func logf(f string, args ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Printf(f, args...)
}

func logln(args ...interface{}) {
	if !*verbose {
		return
	}
	fmt.Println(args...)
}
