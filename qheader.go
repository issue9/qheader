// SPDX-License-Identifier: MIT

// Package qheader 用于处理 quality factor 报头
package qheader

import (
	"net/http"
	"strings"
	"sync"
)

const destroyMaxHeadersLength = 10

var qheaderPool = sync.Pool{New: func() interface{} {
	return &QHeader{Items: make([]*Item, 0, 3)}
}}

type QHeader struct {
	Raw string // 完整的报头内容

	// 以下为解析并排序之后第一个元素的值。

	Value  string
	Params map[string]string
	Q      float64

	// 完整的元素列表
	Items []*Item
}

// Destroy 回收内存
//
// 这是一个可选操作，如果 QHeader 对象操作频繁，调用此方法在一定程序上可以增加性能。
func (q *QHeader) Destroy() {
	if len(q.Items) < destroyMaxHeadersLength {
		for _, vv := range q.Items {
			itemPool.Put(vv)
		}
		qheaderPool.Put(q)
	}
}

// Accept 返回报头 Accept 处理后的内容列表
//
// */* 会被排在最后。
func Accept(r *http.Request) *QHeader {
	return Parse(r.Header.Get("Accept"), "*/*")
}

// AcceptLanguage 返回报头 Accept-Language 处理后的内容列表
//
// 并不会将 * 排序在最后，* 表示匹配任意非列表中的字段。
func AcceptLanguage(r *http.Request) *QHeader {
	return Parse(r.Header.Get("Accept-Language"), "*")
}

// AcceptCharset 返回报头 Accept-Charset 处理后的内容列表
//
// 并不会将 * 排序在最后，* 表示匹配任意非列表中的字段。
func AcceptCharset(r *http.Request) *QHeader {
	return Parse(r.Header.Get("Accept-Charset"), "*")
}

// AcceptEncoding 返回报头 Accept-Encoding 处理后的内容列表
//
// 并不会将 * 排序在最后，* 表示匹配任意非列表中的字段。
func AcceptEncoding(r *http.Request) *QHeader {
	return Parse(r.Header.Get("Accept-Encoding"), "*")
}

// Parse 解析报头内容
//
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后，都是全名则按原来顺序返回。
//
// header 表示报头的内容；
// any 表示通配符的值，只能是 */*、* 和空值，其它情况则 panic；
func Parse(header string, any string) *QHeader {
	if any != "*" && any != "*/*" && any != "" {
		panic("any 值错误")
	}

	qh := qheaderPool.Get().(*QHeader)
	qh.Raw = header
	qh.Items = qh.Items[:0]
	qh.Params = nil
	qh.Q = 0
	qh.Value = ""

	items := strings.Split(header, ",")
	for _, v := range items {
		if v != "" {
			qh.Items = append(qh.Items, parseItem(v))
		}
	}

	sortItems(qh.Items, any)

	if len(qh.Items) == 0 || qh.Items[0].Err != nil {
		qh.Destroy()
		return nil
	}

	first := qh.Items[0]
	qh.Value = first.Value
	qh.Params = first.Params
	qh.Q = first.Q
	return qh
}
