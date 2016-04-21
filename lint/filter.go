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
	"strings"
)

type filter func(<-chan *line, chan<- nit) <-chan *line

func filterFence(lch <-chan *line, nch chan<- nit) <-chan *line {
	ch := make(chan *line)
	go func() {
		defer close(ch)
		i := 0
		lastBlankLine := 0
		lastBlockEndLine := 0
		for l := range lch {
			if l.isBlank() {
				lastBlankLine = l.n
			} else if lastBlockEndLine != 0 && lastBlockEndLine == l.n-1 {
				nch <- nit{l: *l, rule: "spacing: fence with newline after ``` block"}
			}

			if strings.HasPrefix(l.s, "```") {
				i++
				if i&1 == 0 {
					// unmuted the stream
					lastBlockEndLine = l.n
				} else if lastBlankLine != 0 && lastBlankLine != l.n-1 {
					// muted the stream but not preceeded by blank line
					nch <- nit{l: *l, rule: "spacing: fence with newline before ``` blocks"}
				}
			}
			if i&1 == 1 {
				continue
			}
			ch <- l
		}
	}()
	return ch
}
