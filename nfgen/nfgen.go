/*************************************************************************
 * Copyright 2017 Gravwell, Inc. All rights reserved.
 * Contact: <legal@gravwell.io>
 *
 * This software may be modified and distributed under the terms of the
 * BSD 2-clause license. See the LICENSE file for details.
 **************************************************************************/

package nfgen

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const (
	version       int = 2
	defaultDelim      = "\t"
	defaultEngine     = "fields"
)

var (
	empty struct{}
)

// Sub represents a name to index translation.  The DataType field is reserved for future use
// The Index field must be > 0 and must be positive
type Sub struct {
	Name     string
	DataType string `json:",omitempty"`
	Index    int
}

// Group represents a set of sane index to name transitions
// A group also defines an optional delimeter for pure fields extraction or an engine.
// The default engine is "fields" and requires a Delim, the "csv" engine type does not use the delim
type Group struct {
	Delim    string `json:",omitempty"`
	Engine   string `json:",omitempty"`
	Name     string
	Subs     []Sub
	subNames map[string]struct{}
}

// Resource is the global resource that is version checked and contains an N number of group definitions
type Resource struct {
	Version int `json:",omitempty"`
	Set     []Group
}

// NFGen is a structure used to safely create a sane named fields resource
type NFGen struct {
	Resource
	groups map[string]struct{}
}

// NewGen creates a new named fields generator
func NewGen() *NFGen {
	return &NFGen{
		Resource: Resource{
			Version: version,
		},
		groups: map[string]struct{}{},
	}
}

// AddGroup adds a group with subs into the named fields generator
func (n *NFGen) AddGroup(g Group) error {
	if g.Name == `` {
		return errors.New("Group name is empty")
	} else if _, ok := n.groups[g.Name]; ok {
		return fmt.Errorf("Group %s already present in the resource", g.Name)
	} else if len(g.Subs) == 0 {
		return errors.New("Group contains no extraction definitions")
	}
	n.Set = append(n.Set, g)
	n.groups[g.Name] = empty
	return nil
}

// Export encodes a given set of groups into a resource file for use in Gravwell
func (n *NFGen) Export(p string) error {
	if n.Version <= 0 {
		return errors.New("Invalid version")
	} else if len(n.Set) == 0 {
		return errors.New("The set is empty, no groups present")
	}
	fout, err := os.Create(p)
	if err != nil {
		return err
	}
	if err = json.NewEncoder(fout).Encode(n); err != nil {
		fout.Close()
		return err
	}
	return fout.Close()
}

// NewGroup is a convienence function for creating a new group and adding
func NewGroup(name, engine, delim string) (g Group, err error) {
	if name == `` {
		err = errors.New("Name is empty")
		return
	}
	if engine == `` {
		engine = defaultEngine
	}
	if engine == `fields` && delim == `` {
		delim = defaultDelim
	}
	g = Group{
		Name:     name,
		Engine:   engine,
		Delim:    delim,
		subNames: map[string]struct{}{},
	}
	return
}

// AddSub adds a name to index extraction to a group and performs sanity checking
// The dt is currently ignored, but will be used in the future
func (g *Group) AddSub(name, dt string, idx int) error {
	if name == `` {
		return errors.New("Extraction name is empty")
	}
	if idx < 0 {
		return errors.New("Index must be positive")
	}
	if _, ok := g.subNames[name]; ok {
		return fmt.Errorf("%s extraction name already exists", name)
	}
	g.Subs = append(g.Subs, Sub{
		Name:  name,
		Index: idx,
	})
	g.subNames[name] = empty
	return nil
}
