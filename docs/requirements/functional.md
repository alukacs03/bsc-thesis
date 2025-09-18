# Overall functional requirements

- Nodes must be able to auto-enroll themselves into the cluster through the master server.
    - Nodes must be accepted by an administrator on the master node.
    - The master nodes must host a web server running the (simple) control panel.
- The system must establish a stable, secure overlay network between the nodes (using WireGuard).
- Nodes must dynamically advertise routes (OSPF).
    - Nodes must be fault-tolerant in terms of network failures.
    - Nodes must form a hub-and-spoke topology with two central nodes.
- Kubernetes must be able to orchestrate workloads across nodes through the redundant overlay network.
- System health must be monitored, alerts set up for failures.
- At least one demo application must be deployed over the stack.
- The overall system health must be monitored through Grafana dashboards.

---
# Hungarian
# Funkcionális követelmények
- A node-oknak képesnek kell lenniük az automatikus csatlakozásra a klaszterhez a master szerveren keresztül.
    - A node-okat egy adminisztrátornak kell jóváhagynia a master node-on.
    - A master node-oknak egy (egyszerű) vezérlőpultot futtató webkiszolgálót kell üzemeltetniük.
- A rendszernek stabil, biztonságos overlay hálózatot kell kialakítania a node-ok között (WireGuard használatával).
- A node-oknak dinamikusan kell hirdetniük az útvonalakat (OSPF).
    - A node-oknak hibatűrőnek kell lenniük hálózati meghibásodások esetén.
    - A node-oknak hub-and-spoke topológiát kell kialakítaniuk két központi node-dal.
- A Kubernetesnek képesnek kell lennie a munkaterhelések orchesztálására a node-ok között a redundáns overlay hálózaton keresztül.
- A rendszer állapotát monitorozni kell, és riasztásokat kell beállítani a hibákra.
- Legalább egy demó alkalmazást kell telepíteni a stackre.
- A rendszer általános állapotát Grafana dashboardokon keresztül kell figyelemmel kísérni.