# Project scope

## In scope:
- Debian 12 VPS nodes (Proxmox VMs for testing purposes)
- Kubernetes cluster (1 control-plane, 1+ workers)
- WireGuard + OSPF networking (with FRR)
- Go master server + agents
- Demo workload(s)
- Observability stack (Prometheus, Grafana, Loki)

## Out of scope:
- Multi-tenancy integration
- Multi-region scalability
- Enterprise-grade security hardening
- Complex Kubernetes ingress and CNI

---
# Hungarian
# Projekt hatóköre
A hatókörbe tartozik:
- Debian 12 VPS node-ok (teszteléshez Proxmox VM-ek)
- Kubernetes klaszter (1 vezérlősík, 1+ worker)
- WireGuard + OSPF hálózat (FRR-rel)
- Go alapú master szerver + agentek
- Demó munkaterhelés(ek)
- Megfigyelési stack (Prometheus, Grafana, Loki)

A hatókörön kívül esik:
- Multi-tenant integráció
- Több régiós skálázhatóság
- Vállalati szintű kiberbiztonsági megerősítés
- Komplex Kubernetes ingress és CNI