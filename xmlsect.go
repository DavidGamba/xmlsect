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
			nsid = "ns"
			fmt.Fprintf(os.Stderr, "Found default namespace: %s\nUse 'ns' as the namespace ID for xpath: /ns:a/ns:b/ns:c\n\n", ns)
		} else {
			fmt.Fprintf(os.Stderr, "Found '%s' namespace: %s\n\n", nsid, ns)
		}
		// TODO: Check for duplicate nsid
		xmlnsMap[nsid] = ns
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
	xsect <file> [<xpath>] [<relative_xpath>]

	xsect [--help]
`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	opt := getoptions.New()
	opt.Bool("help", false)
	opt.Bool("debug", false)
	opt.Bool("version", false)
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
		version := semver.Version{Major: 0, Minor: 1, Patch: 0}
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
