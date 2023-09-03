package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	// 定义一个启动新进程的参数结构体
	cmd := exec.Command("ls", "-l", "/")

	// 启动新进程并获取进程状态
	process, err := os.StartProcess(cmd.Path, cmd.Args, &os.ProcAttr{
		Dir:   "/",                                        // 设置工作目录
		Env:   os.Environ(),                               // 设置环境变量
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, // 设置标准输入、输出、错误流
	})
	if err != nil {
		fmt.Println("Error starting process:", err)
		return
	}

	// 等待新进程退出
	state, err := process.Wait()
	if err != nil {
		fmt.Println("Error waiting for process:", err)
		return
	}

	fmt.Println("Process exited with status:", state.ExitCode())
}
