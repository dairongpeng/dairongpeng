## 背景
在多人协作开发的项目中，如果review前期欠缺，那么风格和规范一般来说是多样的，随着项目体量的膨胀，越来越不好控制。当前存在一种场景，是如何发现项目中各个文件使用到了中文并且整理出来。找到项目使用到的中文，无论是日志还是常量，都有助于进一步优化项目，也可以针对性的抽取出来做国际化使用。

## FileFind使用
FileFind采用命令行开发的模式，符合工具使用起来简单的特点。核心命令为`filectl`

### 获取命令
这里直接演示`go install的方式`，有条件的小伙伴可以直接进行源码编译。

```shell
Last login: Wed Mar 15 11:06:52 on ttys002
 ~/ go install github.com/dairongpeng/filefind@v0.0.1
go: downloading github.com/dairongpeng/filefind v0.0.1
 ~/ cd go/bin
 ~/go/bin/ ls | grep filefind
filefind
 ~/go/bin/ mv filefind filectl
 ~/go/bin/ ls | grep filectl
filectl
```

### 获取帮助
```shell
 ~/go/bin/ filectl -h
Usage:
  filectl [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  scan        Run file find run cmd

Flags:
  -f, --folder string   need find files from folder
  -h, --help            help for filectl

Use "filectl [command] --help" for more information about a command.
 ~/go/bin/ filectl scan -h
read and filter all files from folder

Usage:
  filectl scan [flags]

Aliases:
  scan, run, find

Flags:
  -e, --export    if export result to json file
  -h, --help      help for scan
  -v, --version   version for scan

Global Flags:
  -f, --folder string   need find files from folder
 ~/go/bin/
```

### 准备待查找的目标文件夹
```shell
 ~/go/bin/ cd ~/workspace/go-workspace/filefind/example
 ~/workspace/go-workspace/filefind/example/ [master] cat main.go
package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("您好，这里我使用了中文;")

	str := "这里不, 确定能不能被, 工具发现呢。"
	if len(strings.Split(str, ",")) == 3 {
		fmt.Println("This is as expected")
	}

	fmt.Println("Hello filectl.")
}
 ~/workspace/go-workspace/filefind/example/ [master]
```

### 进行文件查找并输出报告

报告存在目标文件夹下，名称为`result.json`

```shell
 ~/workspace/go-workspace/filefind/example/ [master] filectl -f . scan -e true
[FILE-FIND] read source folder: .
[FILE-FIND] {
    "path": "./main.go",
    "values": [
        {
            "line": 9,
            "val": "您好这里我使用了中文"
        },
        {
            "line": 11,
            "val": "这里不确定能不能被工具发现呢"
        }
    ]
}
 ~/workspace/go-workspace/filefind/example/ [master] ls
main.go     result.json
 ~/workspace/go-workspace/filefind/example/ [master]
```

## 总结
该工具用来扫描项目的根目录，拿到根目录下文件所有存在不规范或者有待重构的中文句子，汇总出来，整理成文档。后面不管是抽取常量，还是进行国际化改造，就简单很多了。