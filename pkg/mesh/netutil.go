package mesh

import (
	"net"
)

// NetworkInfo contains detected network information
type NetworkInfo struct {
	IP          string
	MAC         string
	HasInternet bool
	Interface   string
}

// DetectNetworkInfo auto-detects the local network configuration
func DetectNetworkInfo() *NetworkInfo {
	info := &NetworkInfo{
		IP:  "127.0.0.1",
		MAC: "00:00:00:00:00:00",
	}

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return info
	}

	// Find the best interface (non-loopback, up, with IP)
	for _, iface := range interfaces {
		// Skip loopback, down interfaces, and virtual interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip IPv6 and loopback
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			// Found a valid IPv4 address
			info.IP = ip.String()
			info.MAC = iface.HardwareAddr.String()
			info.Interface = iface.Name

			// If MAC is empty, use a placeholder
			if info.MAC == "" {
				info.MAC = "00:00:00:00:00:00"
			}

			// Check internet connectivity
			info.HasInternet = CheckInternetConnectivity()

			return info
		}
	}

	return info
}

// CheckInternetConnectivity is in internet.go - use that version

// GetAllNetworkInterfaces returns info about all active network interfaces
func GetAllNetworkInterfaces() []NetworkInfo {
	var results []NetworkInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		return results
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.To4() == nil {
				continue
			}

			mac := iface.HardwareAddr.String()
			if mac == "" {
				mac = "00:00:00:00:00:00"
			}

			results = append(results, NetworkInfo{
				IP:        ip.String(),
				MAC:       mac,
				Interface: iface.Name,
			})
		}
	}

	return results
}
