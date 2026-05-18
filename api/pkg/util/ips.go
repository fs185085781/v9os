package util

import "net"

func GetAllHostIP() []string {
	var ipList []string

	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return ipList
	}

	for _, iface := range interfaces {
		// 关键优化1：检查网卡状态，必须是处于活动状态(up)且不是回环接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取该网卡上的地址列表
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// 类型断言，判断地址是否为 *net.IPNet 类型
			if ipNet, ok := addr.(*net.IPNet); ok {
				// 关键优化2：排除回环地址和链路本地地址(如169.254.x.x)，只保留IPv4
				if !ipNet.IP.IsLoopback() && !ipNet.IP.IsLinkLocalUnicast() && ipNet.IP.To4() != nil {
					ipList = append(ipList, ipNet.IP.String())
				}
			}
		}
	}
	return ipList
}
