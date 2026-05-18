package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	root          = ""
	hasBuildQd    = false
	androidCcPath = "C:\\software\\devlop\\android-ndk-r27d\\toolchains\\llvm\\prebuilt\\windows-x86_64\\bin\\x86_64-linux-android21-clang.cmd"
)

func main() {
	initData()
	res_console := build(false, []string{
		"windows/amd64/.exe",
		"windows/arm64/.exe",
		"linux/amd64/",
		"linux/arm64/",
		"darwin/amd64/",
		"darwin/arm64/",
		"android/amd64/",
		"android/arm64/",
	})
	if !res_console {
		fmt.Println("Console编译失败")
		return
	}
	res_gui := build(true, []string{
		"windows/amd64/.exe",
	})
	if !res_gui {
		fmt.Println("GUI编译失败")
		return
	}
	fmt.Println("整体编译成功")
}
func initData() {
	wd, _ := os.Getwd()
	root = filepath.Dir(wd)
	os.RemoveAll(filepath.Join(root, "dist"))
	os.MkdirAll(filepath.Join(root, "dist"), 0755)

}
func build(gui bool, list []string) bool {
	apiPath := filepath.Join(root, "api")
	if !hasBuildQd {
		os.RemoveAll(filepath.Join(apiPath, "internal", "server", "web"))
		//先编译前端
		cmdQd := exec.Command(
			"npm", "run", "build",
		)
		webPath := filepath.Join(root, "web")
		cmdQd.Dir = webPath
		// 执行编译
		if err := cmdQd.Run(); err != nil {
			fmt.Printf("编译前端失败: %v\n", err)
			return false
		}
		err := os.Remove(filepath.Join(apiPath, "internal", "server", "web", "service.json"))
		if err != nil {
			fmt.Printf("删除service.json失败: %v\n", err)
			return false
		}
		fmt.Println("前端编译成功")
		hasBuildQd = true
	}
	for _, item := range list {
		strings.Split(item, "/")
		osName := strings.Split(item, "/")[0]
		arch := strings.Split(item, "/")[1]
		ext := strings.Split(item, "/")[2]
		// 定义编译命令（含环境变量和编译参数）
		ldflags := "-s -w"
		rkName := "./cmd/console/main.go"
		cgo := "0"
		gn := "console"
		if osName == "android" {
			gui = false
		}
		if gui {
			rkName = "./cmd/gui/main.go"
			cgo = "1"
			gn = "gui"
			if osName == "windows" {
				ldflags += " -H=windowsgui"
			}
		}
		cc := ""
		if osName == "android" && arch == "amd64" {
			cgo = "1"
			cc = androidCcPath
		}
		ext = fmt.Sprintf("_%s_%s_%s%s", osName, arch, gn, ext)
		filePath := filepath.Join(root, "dist", "v9os"+ext)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			cmd := exec.Command(
				"go", "build",
				"-o", filePath,
				"-ldflags", ldflags, // 移除符号表和调试信息
				"-gcflags", "-l=4",
				rkName,
			)
			cmd.Dir = apiPath
			// 设置关键环境变量（静态编译）
			env := append(
				os.Environ(),       // 继承当前环境
				"CGO_ENABLED="+cgo, // 禁用 CGO
				"GOOS="+osName,     // 目标系统
				"GOARCH="+arch,     // 目标架构
				"SYSTEMROOT="+os.Getenv("SYSTEMROOT"),
			)
			if cc != "" {
				env = append(env, "CC="+cc)
			}
			cmd.Env = env
			// 合并输出并执行命令
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("编译失败: %v\n输出:\n%s\n", err, string(output))
				return false
			}
			fmt.Printf("编译成功！输出路径:%s\n", filePath)
		}
	}
	return true
}
