// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
)

func main() {
	for _, path := range os.Args[1:] {
		checkPath(path)
	}
}

var rules = []rule{ruleLines, ruleSpacing, ruleYou, ruleSpell, rulePuncSpace}

func checkPath(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	nch := make(chan nit)
	go func() {
		defer close(nch)
		lch := filterFence(ReadLines(f), nch)
		f := rules[0]
		for _, r := range rules[1:] {
			f = chain(f, r)
		}
		f(lch, nch)
	}()
	for n := range nch {
		fmt.Printf("%s.%d: %s ", n.l.path, n.l.n, n.rule)
		if len(n.badSubstr) > 0 {
			fmt.Printf("(%q in %q)\n", n.badSubstr, n.l.s)
		} else {
			fmt.Printf("(%q)\n", n.l.s)
		}
	}
}
