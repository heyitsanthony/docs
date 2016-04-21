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
	"math/rand"
	"regexp"
	"strings"
	"sync"
)

type nit struct {
	// substring to highlight
	badSubstr  string
	suggestion string
	// violated rule
	rule string
	// line as it is in the file
	l line
}

type rule func(<-chan *line, chan<- nit)

func chain(a, b rule) rule {
	if rand.Intn(2) == 1 {
		// randomize to keep chain tree "bushy"
		t := a
		a = b
		b = t
	}
	return func(lch <-chan *line, nch chan<- nit) {
		var wg sync.WaitGroup
		ach, bch := make(chan *line), make(chan *line)
		go func() {
			a(ach, nch)
			wg.Done()
		}()
		go func() {
			b(bch, nch)
			wg.Done()
		}()
		wg.Add(2)
		for l := range lch {
			ach <- l
			bch <- l
		}
		close(ach)
		close(bch)
		wg.Wait()
	}
}

func rulePuncSpace(lch <-chan *line, nch chan<- nit) {
	re := regexp.MustCompile("[.,;!\"]  ")
	for l := range lch {
		if re.MatchString(l.s) {
			nch <- nit{l: *l, rule: "spacing: punctuation"}
		}
	}
}

func ruleSpell(lch <-chan *line, nch chan<- nit) {
	asp, err := NewASpeller()
	if err != nil {
		panic(err)
	}
	sp := NewMultiSpeller(asp, NewHunSpeller())
	defer sp.Close()

	re := regexp.MustCompile("(http(s|)://[^ ]*[ )\\]?]|[#.,:;()?]|\\[|\\]|\\`[^\\`].*`)")

	for l := range lch {
		s := re.ReplaceAllString(l.s, " ")
		for _, w := range strings.Split(s, " ") {
			if sp.Check(w) || isAcronym(s) {
				continue
			}
			nch <- nit{l: *l, rule: "spelling", badSubstr: w}
		}
	}
}

func ruleYou(lch <-chan *line, nch chan<- nit) {
	for l := range lch {
		for _, w := range strings.Split(l.s, " ") {
			if strings.ToLower(w) == "you" {
				nch <- nit{l: *l, rule: "eschew you"}
			}
		}
	}
}

func ruleLines(lch <-chan *line, nch chan<- nit) {
	for l := range lch {
		for _, n := range l.check() {
			nch <- n
		}
	}
}

func ruleSpacing(lch <-chan *line, nch chan<- nit) {
	lastBlankLine := 0
	lastHeaderLine := 0
	for l := range lch {
		if l.isHeader() {
			if lastBlankLine != l.n-1 {
				nch <- nit{l: *l, rule: "spacing: missing blank line before header"}
			}
			lastHeaderLine = l.n
		}
		if l.isBlank() {
			if lastBlankLine != 0 && lastBlankLine == l.n-1 {
				nch <- nit{l: *l, rule: "spacing: double blank line"}
			}
			lastBlankLine = l.n
		} else if lastHeaderLine != 0 && lastHeaderLine == l.n-1 {
			nch <- nit{l: *l, rule: "spacing: missing blank line after header"}
		}
	}
}
