package network

import (
	"net"
	"net/http"

	"github.com/dvkhr/metrix.git/internal/logging"
)

// checkTrustedSubnet проверяет, принадлежит ли IP-адрес доверенной подсети.
func checkTrustedSubnet(ipStr, trustedSubnet string) bool {
	// Если подсеть не указана, разрешаем все IP-адреса
	if trustedSubnet == "" {
		return true
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		logging.Logg.Error("Invalid CIDR format for trusted subnet: %v", err)
		return false
	}

	return subnet.Contains(ip)
}

// HandleRequestWithTrustedSubnet middleware для проверки доверенной подсети.
func HandleRequestWithTrustedSubnet(trustedSubnet string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.Header.Get("X-Real-IP")
		logging.Logg.Debug("real", "ip", clientIP)
		logging.Logg.Debug("trustedSubnet", "ip", trustedSubnet)

		if clientIP == "" {
			http.Error(w, "X-Real-IP header is missing", http.StatusBadRequest)
			return
		}

		if !checkTrustedSubnet(clientIP, trustedSubnet) {
			http.Error(w, "Forbidden: IP address is not in trusted subnet", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
