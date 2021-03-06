// SPDX-License-Identifier: MIT

// Package qheader 用于处理 quality factor 报头
package qheader

import (
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Header 表示报头内容的单个元素内容
//
// 比如 zh-cmt;q=0.8, zh-cmn;q=1, 拆分成两个 Header 对象。
type Header struct {
	// 完整的报头内容
	Raw string

	// 解析之后的内容

	// 主值
	// 比如 application/json;q=0.9，Value 的值为 application/json
	Value string

	// 其它参数，q 参数也在其中。如果参数数只有名称，没有值，则键值为空。
	// 比如以下值 application/json;q=0.9;level=1;p 将被解析为以下内容：
	//  map[string]string {
	//      "q": "0.9",
	//      "level": "1",
	//      "p": "",
	//  }
	Params map[string]string

	// 为 q 参数的转换后的 float64 类型值
	Q float64

	// 如果 Q 解析失败，则会将错误信息保存在 Err 上
	Err error
}

func (header *Header) hasWildcard() bool {
	return strings.HasSuffix(header.Value, "/*")
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

// 将 Content 的内容解析到 Value 和 Q 中
func parseHeader(content string) *Header {
	val, params, err := mime.ParseMediaType(content)
	if err != nil {
		return &Header{
			Err: err,
			Raw: content,
		}
	}

	h := &Header{
		Raw:    content,
		Params: params,
		Value:  val,
	}

	if len(params) == 0 {
		h.Q = 1
		return h
	}

	if h.Params["q"] != "" {
		h.Q, h.Err = strconv.ParseFloat(h.Params["q"], 32)
	} else {
		h.Q = 1
	}

	return h
}

func sortHeaders(accepts []*Header, any string) {
	sort.SliceStable(accepts, func(i, j int) bool {
		ii := accepts[i]
		jj := accepts[j]

		if ii.Q != jj.Q {
			return ii.Q > jj.Q
		}

		switch {
		case ii.Value == jj.Value:
			return len(ii.Params) > len(jj.Params)
		case ii.Value == any:
			return false
		case jj.Value == any:
			return true
		case ii.hasWildcard(): // 如果 any == * 则此判断不启作用
			return false
		case jj.hasWildcard():
			return true
		default:
			return false
		}
	})
}
