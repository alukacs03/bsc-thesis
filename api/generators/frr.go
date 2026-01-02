package generators

import (
	"fmt"
	"strings"
)

type OSPFInterface struct {
	Name              string
	Cost              int
	IsPointToPoint    bool
	IsDummy           bool
	HelloInterval     int
	DeadInterval      int
	PrefixSuppression bool
}

type FRRConfig struct {
	Hostname       string
	RouterID       string
	IsHub          bool
	LoopbackIP     string
	Interfaces     []OSPFInterface
	OSPFArea       int
}

func GenerateFRRConfig(config FRRConfig) string {
	var sb strings.Builder

	sb.WriteString("frr version 10.5.0\n")
	sb.WriteString("frr defaults traditional\n")
	sb.WriteString(fmt.Sprintf("hostname %s\n", config.Hostname))
	sb.WriteString("log syslog informational\n")

	if !config.IsHub {
		sb.WriteString("no ip forwarding\n")
	}
	sb.WriteString("no ipv6 forwarding\n")
	sb.WriteString("service integrated-vtysh-config\n")
	sb.WriteString("!\n")

	if !config.IsHub && config.LoopbackIP != "" {
		sb.WriteString("route-map RM_SET_SRC permit 10\n")
		sb.WriteString(fmt.Sprintf(" set src %s\n", config.LoopbackIP))
		sb.WriteString("exit\n")
		sb.WriteString("!\n")
	}

	for _, iface := range config.Interfaces {
		sb.WriteString(fmt.Sprintf("interface %s\n", iface.Name))

		if iface.IsDummy {
			sb.WriteString(fmt.Sprintf(" ip ospf area %d\n", config.OSPFArea))
			sb.WriteString(" no ip ospf passive\n")
		} else {
			sb.WriteString(fmt.Sprintf(" ip ospf area %d\n", config.OSPFArea))
			sb.WriteString(fmt.Sprintf(" ip ospf cost %d\n", iface.Cost))

			if iface.HelloInterval > 0 {
				sb.WriteString(fmt.Sprintf(" ip ospf dead-interval %d\n", iface.DeadInterval))
				sb.WriteString(fmt.Sprintf(" ip ospf hello-interval %d\n", iface.HelloInterval))
			}

			if iface.IsPointToPoint {
				sb.WriteString(" ip ospf network point-to-point\n")
			}

			if iface.PrefixSuppression {
				sb.WriteString(" ip ospf prefix-suppression\n")
			}

			sb.WriteString(" no ip ospf passive\n")
		}

		sb.WriteString("exit\n")
		sb.WriteString("!\n")
	}

	sb.WriteString("router ospf\n")
	sb.WriteString(fmt.Sprintf(" ospf router-id %s\n", config.RouterID))

	if !config.IsHub {
		sb.WriteString(" log-adjacency-changes\n")
		sb.WriteString(" max-metric router-lsa administrative\n")
	}

	sb.WriteString(" passive-interface default\n")
	sb.WriteString("exit\n")
	sb.WriteString("!\n")

	if !config.IsHub && config.LoopbackIP != "" {
		sb.WriteString("ip protocol ospf route-map RM_SET_SRC\n")
		sb.WriteString("!\n")
	}

	return sb.String()
}

func GenerateFRRConfigForWorker(hostname string, loopbackIP string, hubInterfaces []string) string {
	interfaces := []OSPFInterface{
		{
			Name:    "dummy",
			IsDummy: true,
		},
	}

	for _, ifaceName := range hubInterfaces {
		interfaces = append(interfaces, OSPFInterface{
			Name:              ifaceName,
			Cost:              10,
			IsPointToPoint:    true,
			HelloInterval:     1,
			DeadInterval:      3,
			PrefixSuppression: true,
		})
	}

	config := FRRConfig{
		Hostname:   hostname,
		RouterID:   loopbackIP,
		IsHub:      false,
		LoopbackIP: loopbackIP,
		Interfaces: interfaces,
		OSPFArea:   10,
	}

	return GenerateFRRConfig(config)
}

func GenerateFRRConfigForHub(hostname string, loopbackIP string, hubToHubInterface string, workerInterfaces []string) string {
	interfaces := []OSPFInterface{
		{
			Name:    "dummy",
			IsDummy: true,
		},
	}

	if hubToHubInterface != "" {
		interfaces = append(interfaces, OSPFInterface{
			Name:              hubToHubInterface,
			Cost:              10,
			IsPointToPoint:    true,
			PrefixSuppression: true,
		})
	}

	for _, ifaceName := range workerInterfaces {
		interfaces = append(interfaces, OSPFInterface{
			Name:              ifaceName,
			Cost:              100,
			IsPointToPoint:    true,
			HelloInterval:     1,
			DeadInterval:      3,
			PrefixSuppression: true,
		})
	}

	config := FRRConfig{
		Hostname:   hostname,
		RouterID:   loopbackIP,
		IsHub:      true,
		LoopbackIP: loopbackIP,
		Interfaces: interfaces,
		OSPFArea:   10,
	}

	return GenerateFRRConfig(config)
}
