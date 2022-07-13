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

func BenchmarkParse(b *testing.B) {
	a := assert.New(b, false)

	b.Run("4", func(b *testing.B) {
		str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
		for i := 0; i < b.N; i++ {
			as := Parse(str, "*/*")
			a.True(len(as) > 0)
		}
	})

	b.Run("pool-4", func(b *testing.B) {
		str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
		for i := 0; i < b.N; i++ {
			as := Parse(str, "*/*")
			a.True(len(as) > 0)
			Destroy(as)
		}
	})

	b.Run("1", func(b *testing.B) {
		str := "application/json;q=0.9"
		for i := 0; i < b.N; i++ {
			as := Parse(str, "*/*")
			a.True(len(as) > 0)
		}
	})

	b.Run("pool-1", func(b *testing.B) {
		str := "application/json;q=0.9"
		for i := 0; i < b.N; i++ {
			as := Parse(str, "*/*")
			a.True(len(as) > 0)
			Destroy(as)
		}
	})
}
