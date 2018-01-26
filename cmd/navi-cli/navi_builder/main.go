package main

import (
	"fmt"
	"git.oschina.net/kuaishangtong/navi/cmd/navi-cli/navi_builder/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
