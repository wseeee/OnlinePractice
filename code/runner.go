package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
)

func main() {
	csd := exec.Command("go", "run", "code-user/main.go")
	var out, stderr bytes.Buffer
	csd.Stderr = &stderr
	csd.Stdout = &out
	stdinPiipe, err := csd.StdinPipe()
	if err != nil {
		log.Print(err)
	}
	io.WriteString(stdinPiipe, "23 11\n")
	//根据测试的输入案例进行运行，拿到输出结果和标准输出结果做出比对

	if err := csd.Run(); err != nil {
		log.Fatalln(err, stderr.String())
	}
	fmt.Println(out.String())

	println(out.String() == "34\n")

}
