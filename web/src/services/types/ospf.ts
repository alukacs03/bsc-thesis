export interface OSPFNeighbor {
  node_id: number;
  node_hostname: string;
  router_id: string;
  area: string;
  state: string;
  interface: string;
  hello_interval_seconds?: number;
  dead_interval_seconds?: number;
  cost?: number;
  priority?: number;
}

