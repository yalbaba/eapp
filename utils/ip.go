package utils

import (
	"fmt"
	"net"
	"os"
)

func GetRealIp() (string, error) {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		os.Exit(1)
		return "", fmt.Errorf("%v", err)
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//fmt.Println(ipnet.IP.String())
				return ipnet.IP.String(), nil
			}

		}
	}

	return "", fmt.Errorf("获取ip失败")
}
