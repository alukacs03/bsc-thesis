# Gluon — szolgáltatófüggetlen HA felhőplatform

Költséghatékony, szolgáltatófüggetlen, nagy rendelkezésre állású felhőplatform Debian alapú VPS csomópontokból. WireGuard overlay hálózat és OSPF dinamikus routing biztosítja a hibatűrő összeköttetést, fölötte Kubernetes klaszter fut konténerizált workload-ok futtatására.

Szakdolgozat — Lukács Ákos, SZTE Informatikai Intézet, 2026.

## Repó struktúra

```
agent/        — Go nyelvű csomópont-ágens (WireGuard / FRR / k8s konfiguráció)
api/          — Go nyelvű vezérlőszerver (Fiber + GORM + SQLite)
web/          — React + TypeScript admin felület (Vite)
ansible/      — build és deploy playbookok
monitoring/   — Prometheus, Grafana, Loki, Promtail, Alertmanager konfig
chaosmonkey/  — failover mérő eszköz (chaosctl)
```

## Függőségek

- Go 1.22+
- Node.js 20+ és npm
- Ansible 2.14+ (deploymenthez)
- Linux célgépeken: WireGuard, FRR, ifupdown2, containerd, kubeadm/kubelet/kubectl

## Build

### Vezérlőszerver (api)

```bash
cd api
go build -o gluon-api .
```

### Csomópont-ágens (agent)

Linuxra cross-compile:

```bash
cd agent
GOOS=linux GOARCH=amd64 go build -o gluon-agent .
```

### Web frontend

```bash
cd web
npm install
npm run build      # dist/ kerül kiadásra
npm run dev        # fejlesztői szerver localhost:5173
```

### Chaosctl

```bash
cd chaosmonkey
go build -o chaosctl ./cmd/chaosctl
```

Vagy egyben Ansible-lel, a célplatformra optimalizálva:

```bash
ansible-playbook ansible/build-agent.yml
ansible-playbook ansible/build-api.yml
```

## Futtatás (lokálisan, fejlesztéshez)

A vezérlőszerver környezeti változókon keresztül konfigurálható (lásd `api/config/config.go`). Minimális indítás:

```bash
cd api
export GLUON_SECRET_KEY="valami-titkos-kulcs"
export GLUON_TLS_ENABLED=false
./gluon-api
```

Frontend a `web/.env.example` mintájára beállítható a backend URL-re.

## Telepítés VPS csomópontokra

Az inventoryt szerkeszteni kell (`ansible/inventory.yaml`), majd:

```bash
ansible-playbook ansible/deploy-api.yml
ansible-playbook ansible/deploy-agent-full.yml
ansible-playbook ansible/deploy-prometheus.yml
ansible-playbook ansible/deploy-grafana.yml
ansible-playbook ansible/deploy-loki.yml
ansible-playbook ansible/deploy-promtail.yml
ansible-playbook ansible/deploy-alertmanager.yml
```

A csomópontok az API-n keresztül enrollolnak; a webes felületen lehet jóváhagyni a feliratkozást.

## Failover mérés

```bash
cd chaosmonkey
cp chaosmonkey.yaml.example chaosmonkey.yaml   # ha még nincs
./chaosctl run                                 # forgatókönyvek futtatása
./measure-failover.sh                          # ping-alapú mérés
```

## Tesztek

```bash
cd api && go test ./...
cd agent && go test ./...
```
