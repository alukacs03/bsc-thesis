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
