// SPDX-FileCopyrightText: 2019-2024 caixw
//
// SPDX-License-Identifier: MIT

package qheader

import (
	"mime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var itemPool = &sync.Pool{New: func() interface{} { return &Item{} }}

// Item 表示报头内容的单个元素内容
//
// 比如 zh-cmt;q=0.8, zh-cmn;q=1, 拆分成两个 Item 对象。
type Item struct {
	Raw string // 原始值

	// 以下为解析之后的内容

	// 主值，比如 application/json;q=0.9，Value 的值为 application/json
	Value string

	// 其它参数，如果参数数只有名称，没有值，则键值为空。q 参数也在其中。
	// 比如以下值 application/json;q=0.9;level=1;p 将被解析为以下内容：
	//  map[string]string {
	//      "q": "0.9",
	//      "level": "1",
	//      "p": "",
	//  }
	Params map[string]string

	// 为 q 参数的转换后的 float64 类型值
	Q float64

	// 如果 Q 解析失败，则会将错误信息保存在 Err 上。
	// 此值不为空，在排序时将排在最后。
	Err error
}

func (header *Item) hasWildcard() bool {
	return strings.HasSuffix(header.Value, "/*")
}

func parseItem(content string) *Item {
	h := itemPool.Get().(*Item)
	h.Q = 1
	h.Raw = content

	h.Value, h.Params, h.Err = mime.ParseMediaType(content)
	if h.Err != nil {
		h.Q = 0
	} else if len(h.Params) > 0 && h.Params["q"] != "" {
		h.Q, h.Err = strconv.ParseFloat(h.Params["q"], 32)
	}

	return h
}

func sortItems(items []*Item, any string) {
	sort.SliceStable(items, func(i, j int) bool {
		ii := items[i]
		jj := items[j]

		if ii.Err != nil {
			return false
		}
		if jj.Err != nil {
			return true
		}

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
