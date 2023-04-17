package config

import "time"

type MyHost struct {
	Host     string
	Hostname string
	Hosttype string
	Domain   string
}

type PillarEntity struct {
	Default  struct{}
	Hosttype struct{}
	Domain   struct{}
	Hostname struct{}
}

const (
	EtcdServer     = "etcd:2379"
	EtcdRoot       = "/pillar_ext"
	RequestTimeout = 3 * time.Second
	Debug          = false
)

func GetEnvList() [5]string {
	return [5]string{"Default", "Hosttype", "Domain", "Hostname"}
}
