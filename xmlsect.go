// This file is part of xmlsect.
//
// Copyright (C) 2017  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/DavidGamba/go-getoptions"
	"github.com/DavidGamba/xmlsect/semver"
	"github.com/santhosh-tekuri/dom"
	"github.com/santhosh-tekuri/xpath"
)

func extractXMLNS(b []byte) map[string]string {
	xmlnsMap := make(map[string]string)
	r := regexp.MustCompile(`xmlns:?([^=]*)=["']?([^"'\s]+)["']?`)
	rm := r.FindAllSubmatch(b, -1)
	for _, m := range rm {
		// 0 - whole match, 1 - nsid, 2 - ns
		nsid := string(m[1])
		ns := string(m[2])
		if nsid == "" {
			nsid = "DEFAULT"
			xmlnsMap[nsid] = ns
			nsid = "_"
			xmlnsMap[nsid] = ns
			fmt.Fprintf(os.Stderr, "Found default namespace: %s\nUse '_' or 'DEFAULT' as the namespace ID for xpath: /_:a/_:b/_:c\n\n", ns)
		} else {
			fmt.Fprintf(os.Stderr, "Found '%s' namespace: %s\n\n", nsid, ns)
			xmlnsMap[nsid] = ns
		}
	}
	return xmlnsMap
}

func printNodeSet(n []dom.Node) {
	var handled bool
	for _, e := range n {
		switch t := e.(type) {
		case *dom.Attr:
			fmt.Printf("%s=%s\n", t.Name, t.Value)
			handled = true
		case *dom.Text:
			fmt.Printf("%s\n", t.Data)
			handled = true
		}
	}
	if !handled {
		doc := &dom.Document{n}
		printDoc(doc)
	}
}

func printNodeSetTree(n []dom.Node, unique bool) {
	var str string
	l := len(n)
	for i, e := range n {
		str += printTreeNode(e, 0, unique)
		if i+1 < l {
			str += "\n"
		}
	}
	fmt.Println(str)
}

// Taken and modified from github.com/santhosh-tekuri/dom/marshal.go
//
//   Copyright 2017 Santhosh Kumar Tekuri. All rights reserved.
//   Use of this source code is governed by a BSD-style
//   license that can be found in the LICENSE file.
//
func printName(n *dom.Name) string {
	var str string
	if n.Prefix != "" {
		str += n.Prefix
		str += ":"
	}
	str += n.Local
	return str
}

// Taken and modified from github.com/santhosh-tekuri/dom/marshal.go
//
//   Copyright 2017 Santhosh Kumar Tekuri. All rights reserved.
//   Use of this source code is governed by a BSD-style
//   license that can be found in the LICENSE file.
//
func printTreeNode(n dom.Node, level int, unique bool) string {
	var str string
	switch n := n.(type) {
	case *dom.Document:
		log.Printf("Document. Children %d\n", len(n.Children()))
		for _, c := range n.Children() {
			str += printTreeNode(c, level+1, unique)
		}
	case *dom.Element:
		str += "/"
		str += printName(n.Name)
		for prefix, _ := range n.NSDecl {
			str += " "
			str += "xmlns"
			if prefix != "" {
				str += ":"
				str += prefix
			}
		}
		for _, attr := range n.Attrs {
			str += " @"
			str += printName(attr.Name)
		}
		if len(n.Children()) != 0 {
			log.Printf("Element '%s' Children %d\n", n.Name, len(n.ChildNodes))
			uniqueMap := make(map[string]int)
			for _, c := range n.Children() {
				tmpStr := printTreeNode(c, level+1, unique)
				// Skip empty results
				if tmpStr == "" {
					continue
				}
				if !unique {
					if strings.HasPrefix(tmpStr, "/") {
						str += "\n"
						str += strings.Repeat("    ", level+1)
					}
					str += tmpStr
				}
				if v, ok := uniqueMap[tmpStr]; ok {
					count := v + 1
					uniqueMap[tmpStr] = count
				} else {
					uniqueMap[tmpStr] = 1
				}
			}
			if !unique {
				return str
			}
			for k, v := range uniqueMap {
				if strings.HasPrefix(k, "/") {
					str += "\n"
					str += strings.Repeat("    ", level+1)
					str += fmt.Sprintf("[%d] %s", v, k)
				} else {
					str += fmt.Sprintf("%s", k)
				}
			}
		}
	case *dom.Text:
		r := regexp.MustCompile(`^\n\s+$|^\s+$`)
		if !r.Match([]byte(n.Data)) {
			str += " text"
		}
	case *dom.ProcInst:
		// TODO
	}
	return str
}

func printDoc(doc *dom.Document) {
	buf := new(bytes.Buffer)
	if err := dom.Marshal(doc, buf); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		return
	}
	fmt.Printf("%s\n", buf.String())
}

func synopsis() {
	synopsis := `# USAGE:
	xsect <file> [<xpath>] [<relative_xpath>] [--tree [--unique]]

	xsect [--help]
`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	var unique bool
	opt := getoptions.New()
	opt.Bool("help", false)
	opt.Bool("debug", false)
	opt.Bool("version", false)
	opt.Bool("tree", false)
	opt.BoolVar(&unique, "unique", false)
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("help") {
		synopsis()
		os.Exit(1)
	}
	if opt.Called("version") {
		version := semver.Version{Major: 0, Minor: 2, Patch: 0}
		fmt.Println(version)
		os.Exit(1)
	}
	if !opt.Called("debug") {
		log.SetOutput(ioutil.Discard)
	}
	log.Println(remaining)
	var file, xpathQuery, xpathRelQuery string
	// var file, xpathQuery string
	l := len(remaining)
	switch {
	case l < 1:
		fmt.Fprintf(os.Stderr, "ERROR: Missing file!\n")
		synopsis()
		os.Exit(1)
	case l == 1:
		file = remaining[0]
		xpathQuery = "/"
	case l == 2:
		file = remaining[0]
		xpathQuery = remaining[1]
	case l >= 3:
		file = remaining[0]
		xpathQuery = remaining[1]
		xpathRelQuery = remaining[2]
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	xmlnsMap := extractXMLNS(b)
	doc, err := dom.Unmarshal(xml.NewDecoder(bytes.NewReader(b)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	compiler := &xpath.Compiler{
		Namespaces: xmlnsMap,
	}
	expr, err := compiler.Compile(xpathQuery)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	log.Printf("xpath %s\n", expr)
	nodeSet, err := expr.EvalNodeSet(doc, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	log.Printf("results: %d\n", len(nodeSet))
	if opt.Called("tree") {
		printNodeSetTree(nodeSet, unique)
		os.Exit(0)
	}
	printNodeSet(nodeSet)
	if xpathRelQuery != "" {
		doc := &dom.Document{nodeSet}
		expr, err := compiler.Compile(xpathRelQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		log.Printf("xpath %s\n", expr)
		nodeSet, err := expr.EvalNodeSet(doc, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		log.Printf("results: %d\n", len(nodeSet))
		printNodeSet(nodeSet)
	}
}
