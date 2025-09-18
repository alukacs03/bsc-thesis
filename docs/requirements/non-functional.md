# Non functional requirements

- **Stability**:
    - The system must tolerate the loss of a single hub node without any service downtime.
- **Availability**:
    - Demo workloads should exhibit a target uptime of over 99.5%.
- **Performance**:
    - Failover events (e.g. pod rescheduling) must complete within 30 seconds.
    - Network latency must not exceed [to-be-defined] ms.
    - FRR, WireGuard and the edge agents must not bog down VPS performance.
- **Seurity**:
    - All control-plane traffic must be restricted to the WireGuard network.
    - There must be no public open ports without a purpose.
- **Reproducibility**:
    - The full cluster setup must be achievable with documented steps.

---
# Hungarian
# Nem funkcionális követelmények
- **Stabilitás**:
    - A rendszernek el kell viselnie egyetlen hub node kiesését anélkül, hogy szolgáltatáskiesés történne.
- **Rendelkezésre állás**:
    - A demó munkaterheléseknek 99,5% feletti célzott rendelkezésre állást kell mutatniuk.
- **Teljesítmény**:
    - A failover eseményeknek (pl. pod-újraütemezés) 30 másodpercen belül le kell zárulniuk.
    - A hálózati késleltetés nem haladhatja meg a [meghatározandó] ms értéket.
    - Az FRR, a WireGuard és az edge agentek nem helyezhetnek túl nagy terhelést a VPS-re.
- **Biztonság**:
    - Az összes vezérlési síkhoz tartozó forgalmat a WireGuard hálózatra kell korlátozni.
    - Nem lehetnek nyilvánosan elérhető portok, amelyeknek nincs kifejezett céljuk.
- **Reprodukálhatóság**:
    - A teljes klaszter felállítása dokumentált lépések mentén újra végrehajtható kell legyen.