package main

import (
	"fmt"
	"os"

	"github.com/geneseven/etcdutils/pkg/utils"
)

func main() {

	if len(os.Args) == 3 {
		utils.GetDataByKey(os.Args[1], os.Args[2])
	} else {
		fmt.Println("getpillar: option requires two argument -- <Pillar_Key> <Hostname>")
		fmt.Println("Usage: getpillar <pillar_key> <Hostname>")
	}
}
