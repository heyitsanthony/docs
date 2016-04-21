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
	"bufio"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

type line struct {
	path string
	n    int
	s    string
}

func NewLine(p string, n int, s string) *line { return &line{path: p, n: n, s: s} }

func ReadLines(f *os.File) <-chan *line {
	ch := make(chan *line)
	p := f.Name()
	go func() {
		defer close(ch)
		r := bufio.NewReader(f)
		i := 1
		for {
			s, _, err := r.ReadLine()
			if err != nil {
				return
			}
			ch <- NewLine(p, i, string(s))
			i++
		}
	}()
	return ch
}

func (l *line) isHeader() bool { return !l.isBlank() && l.s[0] == '#' }
func (l *line) isBlank() bool  { return len(l.s) == 0 }

func (l *line) check() (ret []nit) {
	return l.checkHeader()
}

func (l *line) checkHeader() (ret []nit) {
	if !l.isHeader() {
		return
	}
	words := strings.Split(l.s, " ")[1:]
	if len(words) == 0 {
		ret = append(ret, nit{l: *l, rule: "header: empty header"})
		return
	}
	r, _ := utf8.DecodeRuneInString(words[0])
	if unicode.IsLower(r) {
		n := nit{l: *l, badSubstr: words[0], rule: "header: first word not capitalized"}
		ret = append(ret, n)
		return
	}
	for _, w := range words[1:] {
		if strings.ToLower(w) == w {
			continue
		}
		if isAcronym(w) || isStrictCamelCase(w) {
			continue
		}
		n := nit{l: *l, badSubstr: w, rule: "header: word capitalized"}
		ret = append(ret, n)
	}
	return
}

func isAcronym(s string) bool { return strings.ToUpper(s) == s }

// isStrictCamelCase in the sense you know it's {c,C}amelCase
func isStrictCamelCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	n := 0
	for _, c := range s {
		if unicode.IsUpper(c) {
			n++
		}
	}
	if n > 2 {
		return true
	}
	if n != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return !unicode.IsUpper(r)
}
