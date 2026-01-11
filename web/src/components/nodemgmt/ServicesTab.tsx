import React from "react";
import { toast } from "sonner";
import { handleAPIError } from "@/utils/errorHandler";
import { servicesAPI } from "@/services/api/services";
import type { SystemService } from "@/services/types/service";

interface ServicesTabProps {
  nodeId: number;
  services?: SystemService[];
}

function statusColor(activeState: string) {
  switch (activeState) {
    case "active":
      return "bg-green-500";
    case "inactive":
      return "bg-slate-400";
    case "failed":
      return "bg-red-500";
    default:
      return "bg-yellow-500";
  }
}

function badgeColor(activeState: string) {
  switch (activeState) {
    case "active":
      return "bg-green-100 text-green-800";
    case "inactive":
      return "bg-slate-100 text-slate-700";
    case "failed":
      return "bg-red-100 text-red-800";
    default:
      return "bg-yellow-100 text-yellow-800";
  }
}

const ServicesTab = ({ nodeId, services }: ServicesTabProps) => {
  const [submitting, setSubmitting] = React.useState(false);
  const list = services ?? [];

  const handleRestart = async (name: string) => {
    if (!confirm(`Queue restart for "${name}"?`)) return;
    setSubmitting(true);
    try {
      await servicesAPI.restart(nodeId, name);
      toast.success("Restart queued. Agent will execute it on the next heartbeat.");
    } catch (e) {
      toast.error(handleAPIError(e, "restart service"));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h3 className="text-lg text-slate-800">System Services</h3>
          <p className="text-xs text-slate-500">Updates on heartbeat (~30s).</p>
        </div>
        <div className="space-y-3">
        {list.length === 0 ? (
            <p className="text-sm text-slate-600">No service telemetry yet (waiting for heartbeat).</p>
        ) : (
        list.map((service) => (
            <div key={service.name} className="flex items-center justify-between p-4 border border-slate-200 rounded-lg">
            <div className="flex items-center space-x-4">
                <div className={`w-3 h-3 rounded-full ${statusColor(service.active_state)}`}></div>
                <div>
                <h4 className="text-slate-800">{service.name}</h4>
                <p className="text-sm text-slate-600">{service.description || "—"}</p>
                </div>
            </div>
            <div className="flex items-center space-x-3">
                <span className={`px-2 py-1 rounded text-xs ${
                badgeColor(service.active_state)
                }`}>
                {(service.active_state || "unknown").toUpperCase()}
                </span>
                <span className={`px-2 py-1 rounded text-xs ${
                service.unit_file_state === 'enabled' || service.unit_file_state === 'enabled-runtime'
                    ? 'bg-blue-100 text-blue-800'
                    : 'bg-slate-100 text-slate-700'
                }`}>
                {(service.unit_file_state || "unknown").toLowerCase()}
                </span>
                <div className="flex space-x-1">
                <button
                    disabled={submitting}
                    onClick={() => handleRestart(service.name)}
                    className="px-2 py-1 text-blue-600 hover:bg-blue-50 border border-blue-600 rounded text-xs transition-colors"
                >
                    Restart
                </button>
                </div>
            </div>
            </div>
        )))
    }
        </div>
    </div>
  )
}

export default ServicesTab
