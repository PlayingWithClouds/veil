package vpn

import (
	"net"
	"strings"
)

var vpnPrefixes = []string{"tun", "wg", "ppp", "proton", "nordlynx", "mullvad"}

// IsConnected reports whether a VPN interface is up and its name.
func IsConnected() (connected bool, iface string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false, ""
	}
	for _, i := range interfaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		for _, prefix := range vpnPrefixes {
			if strings.HasPrefix(i.Name, prefix) {
				return true, i.Name
			}
		}
	}
	return false, ""
}
