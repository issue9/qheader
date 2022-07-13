// SPDX-License-Identifier: MIT

// Package qheader 用于处理 quality factor 报头
package qheader

import (
	"net/http"
	"strings"
)

func Destroy(v []*Header) {
	for _, vv := range v {
		pool.Put(vv)
	}
}

// Accept 返回报头 Accept 处理后的内容列表
//
// */* 会被排在最后。
func Accept(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept"), "*/*")
}

// AcceptLanguage 返回报头 Accept-Language 处理后的内容列表
//
// 并不会将 * 排序在最后，* 表示匹配任意非列表中的字段。
func AcceptLanguage(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept-Language"), "*")
}

// AcceptCharset 返回报头 Accept-Charset 处理后的内容列表
//
// 并不会将 * 排序在最后，* 表示匹配任意非列表中的字段。
func AcceptCharset(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept-Charset"), "*")
}

// AcceptEncoding 返回报头 Accept-Encoding 处理后的内容列表
//
// 并不会将 * 排序在最后，* 表示匹配任意非列表中的字段。
func AcceptEncoding(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept-Encoding"), "*")
}

// Parse 将报头内容解析为 []*Header，并对内容进行排序之后返回
//
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后，都是全名则按原来顺序返回。
//
// header 表示报头的内容；
// any 表示通配符的值，只能是 */*、* 和空值，其它情况则 panic；
func Parse(header string, any string) []*Header {
	if any != "*" && any != "*/*" && any != "" {
		panic("any 值错误")
	}

	accepts := make([]*Header, 0, strings.Count(header, ",")+1)

	items := strings.Split(header, ",")
	for _, v := range items {
		if v != "" {
			accepts = append(accepts, parseHeader(v))
		}
	}

	sortHeaders(accepts, any)

	return accepts
}
