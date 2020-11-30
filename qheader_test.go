// SPDX-License-Identifier: MIT

package qheader

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestAccept(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Add("Accept", "text/json;q=0.5,text/xml;q=0.8,application/xml;q=0.8")
	accepts := Accept(r)
	a.NotNil(accepts)
	a.Equal(accepts[0].Value, "text/xml")
	a.Equal(accepts[1].Value, "application/xml")
	a.Equal(accepts[2].Value, "text/json")
}

func TestAcceptLanguage(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Add("Accept-Language", "zh-tw;q=0.5,zh-cn;q=0.8,en;q=0.8")
	accepts := AcceptLanguage(r)
	a.Equal(3, len(accepts))
	a.Equal(accepts[0].Value, "zh-cn")
	a.Equal(accepts[1].Value, "en")
	a.Equal(accepts[2].Value, "zh-tw")
}

func TestAcceptEncoding(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Add("Accept-Encoding", "gzip;q=0.5,compress;q=0.8,*;q=0.6,br")
	accepts := AcceptEncoding(r)
	a.Equal(4, len(accepts))
	a.Equal(accepts[0].Value, "br")
	a.Equal(accepts[1].Value, "compress")
	a.Equal(accepts[2].Value, "*")
	a.Equal(accepts[3].Value, "gzip")
}

func TestAcceptCharset(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Add("Accept-Charset", "utf8;q=0.5,abc;q=0.5,defg;q=0.5,*;q=0.5,cp936,utf32;q=0.4")
	accepts := AcceptCharset(r)
	a.Equal(len(accepts), 6)
	a.Equal(accepts[0].Value, "cp936")
	a.Equal(accepts[1].Value, "utf8")
	a.Equal(accepts[2].Value, "abc")
	a.Equal(accepts[3].Value, "defg")
	a.Equal(accepts[4].Value, "*")
	a.Equal(accepts[5].Value, "utf32")
}

func TestParseHeader(t *testing.T) {
	a := assert.New(t)

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
	a.Equal(h.Value, "").
		Equal(h.Q, float32(0.9)).
		NotError(h.Err)

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

func TestParse(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		Parse(",a1", "not-allow")
	})

	as := Parse(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.8", "*/*")
	a.NotEmpty(as)
	a.Equal(len(as), 6)
	// 确定排序是否正常
	a.Equal(as[0].Q, float32(1.0))
	a.Equal(as[5].Q, float32(.5))

	as = Parse(",a1,a2,a3;q=0.5,a4,a5;q=0.9,a6;a61;q=0.0", "*/*")
	a.NotEmpty(as)
	a.Equal(len(as), 6)
	a.Equal(as[0].Q, float32(1.0))

	// xx/* 的权限低于相同 Q 值的其它权限
	as = Parse("x/*;q=0.1,b/*;q=0.1,a/*;q=0.1,t/*;q=0.1,text/plain;q=0.1", "*/*")
	a.NotEmpty(as)
	a.Equal(len(as), 5)
	a.Equal(as[0].Value, "text/plain").Equal(as[0].Q, float32(0.1))
	a.Equal(as[1].Value, "x/*").Equal(as[1].Q, float32(0.1))
	a.Equal(as[2].Value, "b/*").Equal(as[2].Q, float32(0.1))
	a.Equal(as[3].Value, "a/*").Equal(as[3].Q, float32(0.1))
	a.Equal(as[4].Value, "t/*").Equal(as[4].Q, float32(0.1))

	// xx/* 的权限低于相同 Q 值的其它权限
	as = Parse("text/*;q=0.1,xx/*;q=0.1,text/html;q=0.1", "*/*")
	a.NotEmpty(as)
	a.Equal(len(as), 3)
	a.Equal(as[0].Value, "text/html").Equal(as[0].Q, float32(0.1))
	a.Equal(as[1].Value, "text/*").Equal(as[1].Q, float32(0.1))

	// */* 的权限最底
	as = Parse("text/html;q=0.1,text/*;q=0.1,xx/*;q=0.1,*/*;q=0.1", "*/*")
	a.NotEmpty(as)
	a.Equal(len(as), 4)
	a.Equal(as[0].Value, "text/html").Equal(as[0].Q, float32(0.1))
	a.Equal(as[1].Value, "text/*").Equal(as[1].Q, float32(0.1))

	as = Parse("utf-8;q=x.9,gbk;q=0.8", "*/*")
	a.Equal(2, len(as)).
		Equal(as[0].Value, "gbk").NotError(as[0].Err).
		Equal(as[1].Value, "utf-8").Error(as[1].Err)
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

	// Params 不一样
	as = []*Header{
		{Raw: "1", Value: "zh-cn", Q: 0.7},
		{Raw: "2", Value: "zh-cn", Q: 0.7, Params: map[string]string{"level": "1"}},
	}
	sortHeaders(as, "*")
	a.Equal(as[0].Raw, "2")
	a.Equal(as[1].Raw, "1")
}
