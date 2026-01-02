package generators

import (
	"fmt"
	"strings"
)

type NetworkInterface struct {
	Name           string
	Address        string
	IsDummy        bool
	WireGuardConf  string
}

func GenerateNetworkInterfacesConfig(loopbackIP string, wgInterfaces []NetworkInterface) string {
	var sb strings.Builder

	sb.WriteString("auto dummy\n")
	sb.WriteString("iface dummy inet static\n")
	sb.WriteString(fmt.Sprintf("\taddress %s\n", loopbackIP))
	sb.WriteString("\tpre-up /sbin/ip link add dummy type dummy\n")
	sb.WriteString("\tpost-down /sbin/ip link del dummy\n")

	for _, iface := range wgInterfaces {
		sb.WriteString(fmt.Sprintf("\nauto %s\n", iface.Name))
		sb.WriteString(fmt.Sprintf("iface %s inet static\n", iface.Name))
		sb.WriteString(fmt.Sprintf("\taddress %s\n", iface.Address))
		sb.WriteString(fmt.Sprintf("\tpre-up /sbin/ip link add %s type wireguard\n", iface.Name))
		sb.WriteString(fmt.Sprintf("\tpre-up /usr/bin/wg setconf %s %s\n", iface.Name, iface.WireGuardConf))
		sb.WriteString(fmt.Sprintf("\tpost-down /sbin/ip link delete %s\n", iface.Name))
	}

	return sb.String()
}

func GenerateNetworkInterfacesConfigForWorker(loopbackIP string, hub1Address string, hub2Address string) string {
	wgInterfaces := []NetworkInterface{
		{
			Name:          "wg-hub1",
			Address:       hub1Address,
			WireGuardConf: "/etc/wireguard/wg-hub1.conf",
		},
		{
			Name:          "wg-hub2",
			Address:       hub2Address,
			WireGuardConf: "/etc/wireguard/wg-hub2.conf",
		},
	}

	return GenerateNetworkInterfacesConfig(loopbackIP, wgInterfaces)
}

func GenerateNetworkInterfacesConfigForHub(loopbackIP string, hubToHubAddress string, otherHubName string, workerInterfaces []NetworkInterface) string {
	var wgInterfaces []NetworkInterface

	if hubToHubAddress != "" && otherHubName != "" {
		wgInterfaces = append(wgInterfaces, NetworkInterface{
			Name:          otherHubName,
			Address:       hubToHubAddress,
			WireGuardConf: fmt.Sprintf("/etc/wireguard/%s.conf", otherHubName),
		})
	}

	wgInterfaces = append(wgInterfaces, workerInterfaces...)

	return GenerateNetworkInterfacesConfig(loopbackIP, wgInterfaces)
}

func NewWorkerInterface(workerHostname string, address string) NetworkInterface {
	interfaceName := fmt.Sprintf("wg-%s", workerHostname)
	return NetworkInterface{
		Name:          interfaceName,
		Address:       address,
		WireGuardConf: fmt.Sprintf("/etc/wireguard/%s.conf", interfaceName),
	}
}
