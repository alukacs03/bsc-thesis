import CardContainer from "@/components/CardContainer";
import CardWithIcon from "@/components/CardWithIcon";
import type { Node } from "@/services/types/node";
import { formatBytes, formatPercent } from "@/utils/format";
import { HardDrive, Database } from "lucide-react";

export default function ClusterStorageTab({ nodes }: { nodes: Node[] }) {
  let total = 0;
  let used = 0;
  for (const n of nodes) {
    if (typeof n.disk_total_bytes === "number" && n.disk_total_bytes > 0) total += n.disk_total_bytes;
    if (typeof n.disk_used_bytes === "number" && n.disk_used_bytes >= 0) used += n.disk_used_bytes;
  }
  const pct = total > 0 ? (used / total) * 100 : null;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <CardWithIcon
          title="Total Disk"
          value={formatBytes(total)}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<HardDrive className="w-6 h-6 text-slate-800" />}
        />
        <CardWithIcon
          title="Used Disk"
          value={formatBytes(used)}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<HardDrive className="w-6 h-6 text-slate-800" />}
        />
        <CardWithIcon
          title="Used %"
          value={formatPercent(pct)}
          textColorClass="text-slate-600"
          valueColorClass="text-slate-800"
          iconBGColorClass="bg-blue-100"
          icon={<Database className="w-6 h-6 text-slate-800" />}
        />
      </div>

      <CardContainer title="Kubernetes Storage" icon={<Database className="w-5 h-5" />}>
        <p className="text-sm text-slate-600">
          This will later show StorageClasses, PVs/PVCs, and per-node volume usage. For now the cards above aggregate disk usage
          from Gluon node telemetry.
        </p>
      </CardContainer>
    </div>
  );
}

