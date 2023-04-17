package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/geneseven/etcdutils/config"

	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

func Run() {
	// write to etcd
	pillar := GetConf("/Users/gene/srv/saltstack/pillar/all/init.sls")

	WriteToEtcd(pillar)

	// get key from etcd by hostname
	// myhost := GetMyhost("cacti01.idc1.fsi")
	// fmt.Println(myhost)
	// fmt.Println(config.GetEnvList())

	GetDataByKey("cacti-pkg", "cacti01.dev1.fsi")

}

func GetDataByKey(key string, host string) {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.EtcdServer},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	resp, err := cli.Get(ctx, config.EtcdRoot+"/"+key)

	cancel()
	if err != nil {
		log.Fatal(err)
	}
	myValue := map[string]interface{}{}
	for _, ev := range resp.Kvs {
		// fmt.Printf("%s : %s\n", ev.Key, ev.Value)
		json.Unmarshal(ev.Value, &myValue)
	}

	myhost := GetMyhost(host)
	// myhost_meta := reflect.ValueOf(myhost).Elem()
	rtnMapValue := map[string]interface{}{}
	var rtnListValue []interface{}
	rtnListValueCheckMap := map[string]bool{}
	rtnType := ""
	for _, env := range config.GetEnvList() {
		if v, ok := myValue[env]; ok {
			if env == "Default" {

				switch data := v.(type) {
				case map[string]interface{}:
					// fmt.Println(data)
					rtnType = "map"
					for k2, v2 := range data {
						rtnMapValue[k2] = v2
					}
				case []interface{}:
					rtnType = "list"
					rtnListValue = data
					for _, val := range data {
						switch renderData := val.(type) {
						case string:
							if config.Debug == true {
								fmt.Println("save to rtnListValueCheckMap if default List of String => ", renderData)
							}
							rtnListValueCheckMap[renderData] = true
						}
					}
				default:
					if config.Debug == true {
						fmt.Println("Default case data to defatul:", data)
					}
				}
			} else {
				if rtnType == "" {
					switch data := v.(type) {
					case map[string]interface{}:
						rtnType = "map"
						for _, v2 := range data {
							switch v2.(type) {
							case map[string]interface{}:
								rtnType = "map"
								if config.Debug == true {
									fmt.Println("Non Default Block, check Data Type is Map")
								}
							case []interface{}:
								rtnType = "list"
								if config.Debug == true {
									fmt.Println("Non Default Block, check Data Type is List")
								}
							}
						}
					}
				}
				if rtnType == "map" {
					switch {
					case env == "Hosttype":
						UpdateMapData(v, myhost.Hosttype, rtnMapValue)
					case env == "Domain":
						UpdateMapData(v, myhost.Domain, rtnMapValue)
					case env == "Hostname":
						UpdateMapData(v, myhost.Hostname, rtnMapValue)
					}
				} else {
					switch {
					case env == "Hosttype":
						rtnListValue, err = UpdateListData(v, myhost.Hosttype, rtnListValueCheckMap, rtnListValue)
					case env == "Domain":
						rtnListValue, err = UpdateListData(v, myhost.Domain, rtnListValueCheckMap, rtnListValue)
					case env == "Hostname":
						rtnListValue, err = UpdateListData(v, myhost.Hostname, rtnListValueCheckMap, rtnListValue)
					}
				}
			}
		}
	}

	if rtnType == "map" {
		// fmt.Println(rtnMapValue)
		jsondata, _ := json.Marshal(rtnMapValue)
		fmt.Println(string(jsondata))
	} else {
		rtnListValue, _ = UpdateListWithMap(rtnListValue)
		// fmt.Println(rtnListValue)
		jsondata, _ := json.Marshal(rtnListValue)
		fmt.Println(string(jsondata))
	}

}

func UpdateMapData(v interface{}, myhostValue string, finalValue map[string]interface{}) {
	switch data := v.(type) {
	case map[string]interface{}:
		if localValue, ok := data[myhostValue]; ok {
			switch subData := localValue.(type) {
			case map[string]interface{}:
				for k2, v2 := range subData {
					if config.Debug == true {
						fmt.Println(k2, v2)
					}
					finalValue[k2] = v2
				}
			}
		}
	}
}

func UpdateListData(v interface{}, myhostValue string, checkMap map[string]bool, metaArray []interface{}) (finalArray []interface{}, err error) {
	switch data := v.(type) {
	case map[string]interface{}:
		switch subdata := data[myhostValue].(type) {
		case []interface{}:
			for _, val := range subdata {
				switch dataWithMap := val.(type) {
				case map[string]interface{}:
					if len(dataWithMap) == 1 {
						for k, _ := range dataWithMap {
							if checkMap[k] == false {
								metaArray = append(metaArray, dataWithMap)
								checkMap[k] = true
							}
						}
					} else {
						return metaArray, fmt.Errorf("List With Map, Map must have only one element")
					}

				case string:
					valStr := fmt.Sprint(val)
					if config.Debug == true {
						fmt.Println(valStr)
					}
					if checkMap[valStr] == false {
						if config.Debug == true {
							fmt.Println("UpdateListData: List not found(append it) => ", valStr)
						}
						metaArray = append(metaArray, val)
						checkMap[valStr] = true
					}
				default:
					fmt.Println("ListData unKnow: ", dataWithMap)
				}

			}
		default:
			if config.Debug == true {
				fmt.Println("# Default:", subdata)
			}
		}

	}
	return metaArray, nil
}
func UpdateListWithMap(metaArray []interface{}) ([]interface{}, error) {
	listWithMap := map[string]interface{}{}
	var finalArray []interface{}

	for _, v := range metaArray {
		switch data := v.(type) {
		case map[string]interface{}:
			if len(data) == 1 {
				for k, val := range data {
					listWithMap[k] = val
				}
				if config.Debug == true {
					fmt.Println("ListWithMap size=1 => ", data)
				}
			}
		case string:
			finalArray = append(finalArray, data)
			if config.Debug == true {
				fmt.Println("ListWithMap string block => ", data)
			}
		default:
			if config.Debug == true {
				fmt.Println("UpdateListWithMap default block => ", data)
			}
		}
	}
	for k, v := range listWithMap {
		finalArray = append(finalArray, map[string]interface{}{k: v})
	}

	return finalArray, nil
}

func GetMyhost(hostname string) config.MyHost {
	myhost := config.MyHost{}
	s := strings.Split(hostname, ".")
	pattern := regexp.MustCompile(`(?m)(?P<hosttype>([a-z]\w+)?)(?:\d{1,2})?`)
	if len(s) > 1 {
		myhost.Hostname = hostname
		myhost.Host = s[0]
		myhost.Hosttype = string(pattern.FindAllSubmatch([]byte(s[0]), -1)[0][1])
		myhost.Domain = strings.Join(s[1:], ".")
	}
	if len(s) == 1 {
		myhost.Hostname = ""
		myhost.Host = hostname
		myhost.Hosttype = string(pattern.FindAllSubmatch([]byte(hostname), -1)[0][1])
		myhost.Domain = ""
	}
	if config.Debug == true {
		fmt.Println(myhost)
	}
	return myhost
}

func WriteToEtcd(data map[string]interface{}) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.EtcdServer},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
	}
	defer cli.Close()

	kv := clientv3.NewKV(cli)
	for k, v := range data {
		jsonString, err := json.Marshal(v)
		putResp, err := kv.Put(context.TODO(), config.EtcdRoot+"/"+k, string(jsonString[:]))
		if config.Debug == true {
			fmt.Println(putResp)
		}
		if err != nil {
			fmt.Println("Error putting")
		}

	}

}

func GetConf(file string) map[string]interface{} {
	m := make(map[string]interface{})
	yamlFile, err := ioutil.ReadFile(file)

	err = yaml.Unmarshal(yamlFile, &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m:\n%v\n\n", m)
	for k := range m {
		fmt.Println(k)
	}
	return m
}

func WalkKey(data map[string]interface{}) {
	for k, v := range data {
		fmt.Println(k)
		walk_value(v)
	}
	// switch v := v.(type) {
	// case []interface{}:
	// 	for i, v := range v {
	// 		fmt.Println("index:", i)
	// 		walk(v)
	// 	}
	// case map[interface{}]interface{}:
	// 	for k, v := range v {
	// 		fmt.Println("key:", k)
	// 		walk(v)
	// 	}
	// default:
	// 	fmt.Println(v)
	// 	fmt.Println("#######")
	// }
}

// func WalkEnvCombine(env string, data interface{}, combine_data *interface{}) {
// 	if reflect.TypeOf(data) == reflect.TypeOf(combine_data) {
// 		switch t := reflect.TypeOf(data); t.Kind() {
// 		case reflect.Map:
// 			for k, v := range data {
// 				combine_data[k] = v
// 			}
// 		case reflect.Slice:
// 			m := make(map[string]bool)
// 			for _, v := range data {
// 				m[v] = true
// 			}
// 			for _, v := range combind_data {
// 				if m[v] != true {
// 					combind_data.append(v)
// 				}
// 			}
// 		}
// 	}

// }

func walk_value(data interface{}) {
	switch v := data.(type) {
	case []interface{}:
		for i, v := range v {
			fmt.Println("index:", i)
			walk_value(v)
		}
	case map[string]interface{}:
		for k, v := range v {
			fmt.Println("key:", k)
			walk_value(v)
		}
	default:
		fmt.Println("default=>")
		fmt.Println(v)
	}

}
