package network

import (
	"fmt"
	"net"
)

// GetOutboundIP возвращает внешний IP-адрес хоста.
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "77.88.55.80:53")
	if err != nil {
		return "", fmt.Errorf("failed to determine outbound IP: %w", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
