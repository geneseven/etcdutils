package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/geneseven/etcdutils/pkg/utils"
)

func main() {
	var pillar map[string]interface{}
	if len(os.Args) == 2 {
		if _, err := os.Stat(os.Args[1]); errors.Is(err, os.ErrNotExist) {
			fmt.Println(os.Args[1], " File not exit.")
			os.Exit(1)
		}
		pillar = utils.GetConf(os.Args[1])
	} else {
		fmt.Println("setpillar: option requires one argument -- <yaml_file>")
		fmt.Println("Usage: setpillar <yaml_file>")
	}
	utils.WriteToEtcd(pillar)
}
