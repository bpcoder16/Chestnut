package utils

import "net"

// CheckIPv4Valid 判断字符串是否是有效的 IPv4 地址
func CheckIPv4Valid(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// CheckPortValid 判断端口是否有效 (1 - 65535)
func CheckPortValid(port uint) bool {
	return port >= 1 && port <= 65535
}
