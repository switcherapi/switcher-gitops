package main

import (
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/server"
)

func main() {
	config.InitEnv()
	server.Init()
}
