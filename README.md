qheader
[![Go](https://github.com/issue9/qheader/workflows/Go/badge.svg)](https://github.com/issue9/qheader/actions?query=workflow%3AGo)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/issue9/qheader/branch/master/graph/badge.svg)](https://codecov.io/gh/issue9/qheader)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/issue9/qheader)](https://pkg.go.dev/github.com/issue9/qheader)
======

解析报 quality factor 报头的内容，诸如 Accept、Accept-Charset 等报头。

```go
accepts := qheader.AcceptEncoding("gzip,compress;q=0.9,*;q=0.5,br")
// 返回 br,gzip,compress,* 的顺序
```

安装
----

```shell
go get github.com/issue9/qheader
```

版权
----

本项目源码采用 [MIT](https://opensource.org/licenses/MIT) 开源授权许可证，完整的授权说明可在 [LICENSE](LICENSE) 文件中找到。
