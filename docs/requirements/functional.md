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