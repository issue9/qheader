// SPDX-License-Identifier: MIT

package qheader

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func BenchmarkParseHeader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parseHeader("application/xml;q=0.9")
	}
}

func BenchmarkParse_multiple(b *testing.B) {
	a := assert.New(b, false)

	str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
	for i := 0; i < b.N; i++ {
		as := Parse(str, "*/*")
		a.True(len(as) > 0)
	}
}

func BenchmarkParse_one(b *testing.B) {
	a := assert.New(b, false)

	str := "application/json;q=0.9"
	for i := 0; i < b.N; i++ {
		as := Parse(str, "*/*")
		a.True(len(as) > 0)
	}
}
