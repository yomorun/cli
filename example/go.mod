module example

go 1.16

require (
	github.com/yomorun/cli v0.1.0
	github.com/yomorun/yomo v1.4.1
)

replace (
	// replace github.com/yomorun/yomo => /Users/xiaojianhong/Git/yomo/yomo
	// replace github.com/yomorun/yomo => /Users/fanweixiao/_wrk/yomo
	github.com/yomorun/cli => /Users/venjiang/gopath/src/github.com/yomorun/cli
	github.com/yomorun/yomo => /Users/venjiang/gopath/src/github.com/yomorun/yomo
)
