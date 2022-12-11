package helper

import (
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	rootPath = "" // 用于记录获取到的根目录，多次获取时减少函数调用
)

// RootPath 获取当前可执行文件所在目录绝对路径
//  - go run开发模式时获取的是源码所在根目录绝对路径
//  - go build编译后执行时获取的是可执行二进制文件所在目录的绝对路径
//  - 返回值无后缀斜杠
func RootPath() string {
	if rootPath != "" {
		return rootPath
	}
	// +++++++
	// os.Executable() 获取当前可执行文件的绝对路径，开发时（go run xxx）是临时目录
	// os.Args[0] 当前可执行文件的绝对路径，等价于os.Executable()
	// runtime.Caller(0) 获取的是编译时当前文件（helper.go）所在目录，编译后获取到的一直都是编译前的路径
	// +++++++
	ePath, _ := os.Executable()
	if !strings.Contains(ePath, os.TempDir()) {
		rootPath = path.Dir(ePath)
		return rootPath
	}

	_, filename, _, _ := runtime.Caller(0)
	rootPath = path.Dir(filename)
	return rootPath
}

// Path 获取相对于项目根目录下的指定 dir 的绝对路径
//  - 获取项目根目录绝对路径请使用 RootPath
//  - 获取项目根目录下的名称为 dir 的子目录的绝对路径
//  - 返回值无后缀斜杠
//  - 并不会检查指定的子目录是否存在，用于快捷获取项目根目录下的某个目录的绝对路径
func Path(dir string) string {
	return RootPath() + "/" + strings.Trim(strings.Trim(dir, "."), "/")
}
