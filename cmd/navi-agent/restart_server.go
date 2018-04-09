package main

import (
	"kuaishangtong/common/utils/log"
	"os/exec"
)

func restart_server_in_docker() error {
	if _, err := exec.Command("bash", "-c", "/rpc/run.sh", "restart").Output(); err != nil {
		log.Errorf("Command: %s Error: %v", "bash -c /rpc/run.sh restart", err)
		return err
	}
	return nil
}
