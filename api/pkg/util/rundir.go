package util

import (
	"os"
	"path/filepath"
	"runtime"
)

var dir = ""

func RunDir() string {
	//dir = "/Users/afan/projects/v9os-private/main/api/bin"
	if dir == "" {
		fp := RunFile()
		if fp == "" {
			dir, _ = os.Getwd()
		} else {
			dir = filepath.Dir(fp)
		}
	}
	return dir
}

var file = "-1"

func RunFile() string {
	if file == "-1" {
		tmp1, _ := os.Executable()
		if runtime.GOOS == "android" {
			tmp2, _ := os.Getwd()
			tmp3 := filepath.Dir(tmp1)
			if tmp2 != tmp3 {
				//termux启动
				file = ""
			}
		}
		if file == "-1" {
			//uniapp启动 或者其他平台
			file = tmp1
		}
	}
	return file
}
