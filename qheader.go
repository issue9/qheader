// SPDX-License-Identifier: MIT

// Package qheader 用于处理 quality factor 报头
package qheader

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Header 表示报头内容的单个元素内容
//
// 比如 zh-cmt;q=0.8, zh-cmn;q=1 分根据 , 拆分成两个 Header 对象。
type Header struct {
	Content string // 完整的内容

	// 解析之后的内容
	Value string
	Q     float32
	Err   error
}

func (header *Header) hasWildcard() bool {
	return strings.HasSuffix(header.Value, "/*")
}

// 将 Content 的内容解析到 Value 和 Q 中
func parseHeader(content string) *Header {
	index := strings.IndexByte(content, ';')
	if index < 0 { // 不包含 ; 表示除主内容不包含其它任何附加信息
		return &Header{Content: content, Value: content, Q: 1}
	}

	val := content[:index]
	if index = strings.LastIndex(content, ";q="); index >= 0 {
		q, err := strconv.ParseFloat(content[index+3:], 32)
		if err != nil {
			return &Header{Content: content, Value: val, Err: err}
		}
		return &Header{Content: content, Value: val, Q: float32(q)}
	}
	return &Header{Content: content, Value: val, Q: 1}
}

// Accept 返回报头 Accept 处理后的内容列表
func Accept(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept"), "*/*")
}

// AcceptLanguage 返回报头 Accept-Language 处理后的内容列表
func AcceptLanguage(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept-Language"), "*")
}

// AcceptCharset 返回报头 Accept-Charset 处理后的内容列表
func AcceptCharset(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept-Charset"), "*")
}

// AcceptEncoding 返回报头 Accept-Encoding 处理后的内容列表
func AcceptEncoding(r *http.Request) []*Header {
	return Parse(r.Header.Get("Accept-Encoding"), "*")
}

// Parse 将报头内容解析为 []*Header，并对内容进行排序之后返回
//
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后。
//
// q 值为 0 的数据将被过滤，比如：
//  application/*;q=0.1,application/xml;q=0.1,text/html;q=0
// 其中的 text/html 不会被返回，application/xml 的优先级会高于 application/*
//
// header 表示报头的内容；
// any 表示通配符的值，只能是 */* 或是 *，其它情况则 panic；
func Parse(header string, any string) []*Header {
	if any != "*" && any != "*/*" {
		panic("any 值错误")
	}

	accepts := make([]*Header, 0, strings.Count(header, ",")+1)

	for {
		index := strings.IndexByte(header, ',')
		if index == 0 { // 过滤掉空值
			header = header[1:]
			continue
		}

		if index == -1 { // 最后一条数据
			if header != "" {
				accepts = append(accepts, parseHeader(header))
			}
			break
		}

		// 由上面的两个 if 保证，此处 v 肯定不为空
		accepts = append(accepts, parseHeader(header[:index]))

		header = header[index+1:]
	}

	sortHeaders(accepts, any)

	return accepts
}

func sortHeaders(accepts []*Header, any string) {
	sort.SliceStable(accepts, func(i, j int) bool {
		ii := accepts[i]
		jj := accepts[j]

		if ii.Q != jj.Q {
			return ii.Q > jj.Q
		}

		switch {
		case ii.Value == any:
			return false
		case jj.Value == any:
			return true
		case ii.hasWildcard(): // 如果 any == * 则此判断不启作用
			return false
		default: // !ii.hasWildcard()
			return jj.hasWildcard()
		}
	})
}
