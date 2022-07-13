// SPDX-License-Identifier: MIT

package qheader

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestParseHeader(t *testing.T) {
	a := assert.New(t, false)

	h := parseHeader("application/xml")
	a.Equal(h.Value, "application/xml").
		Equal(h.Q, 1.0).
		NotError(h.Err)

	h = parseHeader("application/xml;")
	a.Equal(h.Value, "application/xml").
		Equal(h.Q, 1.0).
		NotError(h.Err)

	h = parseHeader("application/xml;q=0.9")
	a.Equal(h.Value, "application/xml").
		Equal(h.Q, float32(0.9)).
		NotError(h.Err)

	h = parseHeader(";application/xml;q=0.9")
	a.Error(h.Err).
		Equal(h.Raw, ";application/xml;q=0.9").
		Empty(h.Value).
		Empty(h.Params)

	h = parseHeader("application/xml;qq=xx;q=0.9")
	a.Equal(h.Value, "application/xml").
		Equal(h.Q, float32(0.9)).
		NotError(h.Err)

	h = parseHeader("text/html;format=xx;q=0.9")
	a.Equal(h.Value, "text/html").
		Equal(h.Q, float32(0.9)).
		NotError(h.Err)

	// 要求 q 必须在最后，否则被录作 q 值的一部分
	h = parseHeader("text/html;q=0.9;format=xx")
	a.NotError(h.Err).
		Equal(h.Raw, "text/html;q=0.9;format=xx").
		Equal(h.Q, float32(0.9)).
		Equal(h.Value, "text/html").
		Equal(h.Params, map[string]string{"format": "xx", "q": "0.9"})

	h = parseHeader("text/html;format=xx;q=x.9")
	a.Error(h.Err)

	h = parseHeader("text/html;format=xx;q=0.9x")
	a.Error(h.Err)
}

func TestSortHeaders(t *testing.T) {
	a := assert.New(t, false)

	as := []*Header{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
	}
	sortHeaders(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "*/*")

	as = []*Header{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
	}
	sortHeaders(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "b/*")
	a.Equal(as[2].Value, "*/*")

	as = []*Header{
		{Value: "*/*", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "c/c", Q: 0.7},
		{Value: "b/*", Q: 0.7},
	}
	sortHeaders(as, "*/*")
	a.Equal(as[0].Value, "c/c")
	a.Equal(as[1].Value, "a/*")
	a.Equal(as[2].Value, "b/*")
	a.Equal(as[3].Value, "*/*")

	as = []*Header{
		{Value: "d/d", Q: 0.7},
		{Value: "a/*", Q: 0.7},
		{Value: "*/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
		{Value: "c/c", Q: 0.7},
	}
	sortHeaders(as, "*/*")
	a.Equal(as[0].Value, "d/d")
	a.Equal(as[1].Value, "c/c")
	a.Equal(as[2].Value, "a/*")
	a.Equal(as[3].Value, "b/*")
	a.Equal(as[4].Value, "*/*")

	// Q 值不一样
	as = []*Header{
		{Value: "d/d", Q: 0.7},
		{Value: "a/*", Q: 0.8},
		{Value: "*/*", Q: 0.7},
		{Value: "b/*", Q: 0.7},
		{Value: "c/c", Q: 0.7},
	}
	sortHeaders(as, "*/*")
	a.Equal(as[0].Value, "a/*")
	a.Equal(as[1].Value, "d/d")
	a.Equal(as[2].Value, "c/c")
	a.Equal(as[3].Value, "b/*")
	a.Equal(as[4].Value, "*/*")

	// 相同 Q 值，保持原样
	as = []*Header{
		{Value: "zh-cn", Q: 0.7},
		{Value: "zh-tw", Q: 0.8},
		{Value: "*", Q: 0.7},
		{Value: "en", Q: 0.7},
		{Value: "en-us", Q: 0.7},
	}
	sortHeaders(as, "*")
	a.Equal(as[0].Value, "zh-tw")
	a.Equal(as[1].Value, "zh-cn")
	a.Equal(as[2].Value, "en")
	a.Equal(as[3].Value, "en-us")
	a.Equal(as[4].Value, "*")

	// 相同 Q 值，Err 不同
	as = []*Header{
		{Value: "zh-cn", Q: 0.7, Err: errors.New("zh-cn")},
		{Value: "zh-tw", Q: 0.8},
		{Value: "*", Q: 0.7},
		{Value: "en", Q: 0.7, Err: errors.New("en")},
		{Value: "en-us", Q: 0.7},
	}
	sortHeaders(as, "*")
	a.Equal(as[0].Value, "zh-tw")
	a.Equal(as[1].Value, "en-us")
	a.Equal(as[2].Value, "*")
	a.Equal(as[3].Value, "zh-cn")
	a.Equal(as[4].Value, "en")

	// Params 不一样
	as = []*Header{
		{Raw: "1", Value: "zh-cn", Q: 0.7},
		{Raw: "2", Value: "zh-cn", Q: 0.7, Params: map[string]string{"level": "1"}},
	}
	sortHeaders(as, "*")
	a.Equal(as[0].Raw, "2")
	a.Equal(as[1].Raw, "1")
}
