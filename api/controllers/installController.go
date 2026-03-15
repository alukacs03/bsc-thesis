package controllers

import (
	"bytes"
	"gluon-api/config"
	"gluon-api/logger"
	"os"
	"strings"
	"text/template"

	"github.com/gofiber/fiber/v2"
)

func ServeAgentBinary(c *fiber.Ctx) error {
	path := config.Current().AgentBinaryPath
	data, err := os.ReadFile(path)
	if err != nil {
		logger.Error("Agent binary not found", "path", path, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agent binary not found",
		})
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", "attachment; filename=gluon-agent")
	return c.Send(data)
}

func ServeInstallScript(c *fiber.Ctx) error {
	host := c.Hostname()
	scheme := c.Protocol()

	apiURL := scheme + "://" + host
	if h := c.Get("Host"); h != "" {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			port := parts[1]
			isStandard := (scheme == "http" && port == "80") || (scheme == "https" && port == "443")
			if !isStandard {
				apiURL = scheme + "://" + parts[0] + ":" + port
			}
		}
	}

	role := c.Query("role", "worker")
	hostname := c.Query("hostname")
	provider := c.Query("provider", "local")

	data := map[string]string{
		"APIURL":   apiURL,
		"Role":     role,
		"Hostname": hostname,
		"Provider": provider,
	}

	tmpl, err := template.New("install").Parse(installScriptTmpl)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to render install script",
		})
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to render install script",
		})
	}

	c.Set("Content-Type", "text/plain")
	return c.SendString(buf.String())
}

var installScriptTmpl = `#!/bin/bash
set -euo pipefail

API_URL="{{.APIURL}}"
ROLE="{{.Role}}"
HOSTNAME="{{if .Hostname}}{{.Hostname}}{{else}}$(hostname -f 2>/dev/null || hostname){{end}}"
PROVIDER="{{.Provider}}"

while [ $# -gt 0 ]; do
  case "$1" in
    --api-url)  API_URL="$2";  shift 2;;
    --role)     ROLE="$2";     shift 2;;
    --hostname) HOSTNAME="$2"; shift 2;;
    --provider) PROVIDER="$2"; shift 2;;
    *) echo "[gluon] unknown option: $1"; exit 1;;
  esac
done

log() { echo "[gluon] $*"; }

# pre-flight
[ "$(id -u)" -eq 0 ] || { log "must run as root"; exit 1; }
command -v apt >/dev/null || { log "apt not found"; exit 1; }
if systemctl is-active --quiet gluon-agent 2>/dev/null; then
  log "gluon-agent already running"; exit 0
fi

# system prep
log "configuring system..."
cat > /etc/sysctl.d/99-gluon.conf <<SYSCTL
net.ipv4.ip_forward=1
net.ipv4.conf.all.rp_filter=0
net.ipv4.conf.default.rp_filter=0
net.bridge.bridge-nf-call-iptables=1
net.bridge.bridge-nf-call-ip6tables=1
SYSCTL
sysctl --system >/dev/null 2>&1
cat > /etc/modules-load.d/gluon.conf <<MODULES
wireguard
br_netfilter
overlay
MODULES
modprobe wireguard br_netfilter overlay 2>/dev/null || true
swapoff -a
sed -i '/\sswap\s/d' /etc/fstab

# tls bootstrap
log "fetching CA certificate..."
mkdir -p /etc/gluon
curl -sk "$API_URL/api/ca.crt" -o /etc/gluon/ca.crt

# download agent
log "downloading agent binary..."
curl -f --cacert /etc/gluon/ca.crt "$API_URL/install/agent" -o /usr/local/bin/gluon-agent
chmod +x /usr/local/bin/gluon-agent

# agent config
if [ -f /etc/gluon/agent.conf ] && grep -q '"api_key"' /etc/gluon/agent.conf; then
  log "existing agent config found, keeping it"
else
  log "writing agent config..."
  cat > /etc/gluon/agent.conf <<CONF
{
  "api_url": "$API_URL",
  "desired_role": "$ROLE",
  "hostname": "$HOSTNAME",
  "provider": "$PROVIDER",
  "ca_cert_path": "/etc/gluon/ca.crt"
}
CONF
fi

# systemd service
log "installing systemd service..."
cat > /etc/systemd/system/gluon-agent.service <<UNIT
[Unit]
Description=Gluon Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/gluon-agent
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable --now gluon-agent

log "done! agent installed and running"
log "  api:      $API_URL"
log "  role:     $ROLE"
log "  hostname: $HOSTNAME"
log "  provider: $PROVIDER"
`
