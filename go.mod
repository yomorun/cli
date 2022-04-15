module github.com/yomorun/cli

go 1.16

require (
	github.com/briandowns/spinner v1.18.1
	github.com/dop251/goja v0.0.0-20220110113543-261677941f3c
	github.com/fatih/color v1.13.0
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	github.com/yomorun/yomo v1.7.2
	golang.org/x/tools v0.1.10
	gopkg.in/ini.v1 v1.66.4 // indirect
)

replace github.com/yomorun/yomo => ../yomo
