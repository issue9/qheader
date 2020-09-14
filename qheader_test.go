// SPDX-License-Identifier: MIT

package qheader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func BenchmarkParseHeader(b *testing.B) {
	a := assert.New(b)

	for i := 0; i < b.N; i++ {
		_, _, err := parseHeader("application/xml;q=0.9")
		a.NotError(err)
	}
}

func BenchmarkParse_multiple(b *testing.B) {
	a := assert.New(b)

	str := "application/json;q=0.9,text/plain;q=0.8,text/html,text/xml,*/*;q=0.1"
	for i := 0; i < b.N; i++ {
		as, err := Parse(str, "*/*")
		a.NotError(err).True(len(as) > 0)
	}
}

func BenchmarkParse_one(b *testing.B) {
	a := assert.New(b)

	str := "application/json;q=0.9"
	for i := 0; i < b.N; i++ {
		as, err := Parse(str, "*/*")
		a.NotError(err).True(len(as) > 0)
	}
}

func TestAccept(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Add("Accept", "text/json;q=0.5,text/xml;q=0.8,application/xml;q=0.8")
	accepts, err := Accept(r)
	a.NotError(err).NotNil(accepts)
	a.Equal(accepts[0].Value, "text/xml")
	a.Equal(accepts[1].Value, "application/xml")
	a.Equal(accepts[2].Value, "text/json")
}

func TestAcceptLanguage(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Add("Accept-Language", "zh-tw;q=0.5,zh-cn;q=0.8,en;q=0.8")
	accepts, err := AcceptLanguage(r)
	a.NotError(err).NotNil(accepts)
	a.Equal(accepts[0].Value, "zh-cn")
	a.Equal(accepts[1].Value, "en")
	a.Equal(accepts[2].Value, "zh-tw")
}

func TestParseHeader(t *testing.T) {
	a := assert.New(t)

	v, q, err := parseHeader("application/xml")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, 1.0)

	v, q, err = parseHeader("application/xml;")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, 1.0)

	v, q, err = parseHeader("application/xml;q=0.9")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, float32(0.9))

	v, q, err = parseHeader(";application/xml;q=0.9")
	a.NotError(err).
		Equal(v, "").
		Equal(q, float32(0.9))

	v, q, err = parseHeader("application/xml;qq=xx;q=0.9")
	a.NotError(err).
		Equal(v, "application/xml").
		Equal(q, float32(0.9))

	v, q, err = parseHeader("text/html;format=xx;q=0.9")
	a.NotError(err).
		Equal(v, "text/html").
		Equal(q, float32(0.9))

	// 要求 q 必须在最后，否则还是会出错
	v, q, err = parseHeader("text/html;q=0.9;format=xx")
	a.Error(err).Empty(v).Empty(q)

	v, q, err = parseHeader("text/html;format=xx;q=x.9")
	a.Error(err).Empty(v).Empty(q)

	v, q, err = parseHeader("text/html;format=xx;q=0.9x")
	a.Error(err).Empty(v).Empty(q)
}

func TestParse(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		Parse(",a1", "not-allow")
	})

	as, err := Parse(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.8", "*/*")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 6)
	// 确定排序是否正常
	a.Equal(as[0].Q, float32(1.0))
	a.Equal(as[5].Q, float32(.5))

	// 0.0 会被过滤
	as, err = Parse(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.0", "*/*")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 5)
	a.Equal(as[0].Q, float32(1.0))

	// xx/* 的权限低于相同 Q 值的其它权限
	as, err = Parse("x/*;q=0.1,b/*;q=0.1,a/*;q=0.1,t/*;q=0.1,text/plain;q=0.1", "*/*")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 5)
	a.Equal(as[0].Value, "text/plain").Equal(as[0].Q, float32(0.1))
	a.Equal(as[1].Value, "x/*").Equal(as[1].Q, float32(0.1))
	a.Equal(as[2].Value, "b/*").Equal(as[2].Q, float32(0.1))
	a.Equal(as[3].Value, "a/*").Equal(as[3].Q, float32(0.1))
	a.Equal(as[4].Value, "t/*").Equal(as[4].Q, float32(0.1))

	// xx/* 的权限低于相同 Q 值的其它权限
	as, err = Parse("text/*;q=0.1,xx/*;q=0.1,text/html;q=0.1", "*/*")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 3)
	a.Equal(as[0].Value, "text/html").Equal(as[0].Q, float32(0.1))
	a.Equal(as[1].Value, "text/*").Equal(as[1].Q, float32(0.1))

	// */* 的权限最底
	as, err = Parse("text/html;q=0.1,text/*;q=0.1,xx/*;q=0.1,*/*;q=0.1", "*/*")
	a.NotError(err).NotEmpty(as)
	a.Equal(len(as), 4)
	a.Equal(as[0].Value, "text/html").Equal(as[0].Q, float32(0.1))
	a.Equal(as[1].Value, "text/*").Equal(as[1].Q, float32(0.1))

	as, err = Parse(",a1,a2,a3;q=5,a4,a5;q=0.9,a6;a61;q=0.x8", "*/*")
	a.Error(err).Empty(as)

	as, err = Parse("utf-8;q=x.9,gbk;q=0.8", "*/*")
	a.Error(err).Empty(as)
}

func TestSortHeaders(t *testing.T) {
	a := assert.New(t)

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

	// Q 值不一样
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
}
