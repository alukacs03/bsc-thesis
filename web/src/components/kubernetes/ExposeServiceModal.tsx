import { useMemo, useState } from "react";
import { X, Share2, CheckCircle, AlertCircle } from "lucide-react";
import { toast } from "sonner";
import { kubernetesAPI } from "@/services/api/kubernetes";
import { handleAPIError } from "@/utils/errorHandler";

export interface ExposeTargetInfo {
  namespace: string;
  name: string;
  kind: string;
}

interface ExposeServiceModalProps {
  target: ExposeTargetInfo;
  onClose: () => void;
  onSuccess: () => void;
}

function sanitizeName(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

export default function ExposeServiceModal({ target, onClose, onSuccess }: ExposeServiceModalProps) {
  const [serviceName, setServiceName] = useState(() => sanitizeName(`${target.name}-nodeport`));
  const [selectorKey, setSelectorKey] = useState("app");
  const [selectorValue, setSelectorValue] = useState(target.name);
  const [servicePort, setServicePort] = useState("80");
  const [targetPort, setTargetPort] = useState("80");
  const [nodePort, setNodePort] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ success: boolean; output: string } | null>(null);

  const yaml = useMemo(() => {
    const lines = [
      "apiVersion: v1",
      "kind: Service",
      "metadata:",
      `  name: ${serviceName || sanitizeName(`${target.name}-nodeport`)}`,
      `  namespace: ${target.namespace}`,
      "spec:",
      "  type: NodePort",
      "  selector:",
      `    ${selectorKey || "app"}: ${selectorValue || target.name}`,
      "  ports:",
      "    - name: http",
      `      port: ${servicePort || "80"}`,
      `      targetPort: ${targetPort || servicePort || "80"}`,
      "      protocol: TCP",
    ];
    if (nodePort.trim()) {
      lines.push(`      nodePort: ${nodePort.trim()}`);
    }
    return `${lines.join("\n")}\n`;
  }, [nodePort, selectorKey, selectorValue, serviceName, servicePort, target.name, target.namespace, targetPort]);

  const handleApply = async () => {
    if (!serviceName.trim()) {
      toast.error("Service name is required");
      return;
    }
    if (!selectorKey.trim() || !selectorValue.trim()) {
      toast.error("Selector key/value are required");
      return;
    }
    if (!servicePort.trim() || !targetPort.trim()) {
      toast.error("Port and targetPort are required");
      return;
    }
    if (nodePort.trim() && Number(nodePort) < 30000) {
      toast.error("NodePort must be >= 30000 or left blank");
      return;
    }

    setLoading(true);
    setResult(null);

    try {
      const response = await kubernetesAPI.applyManifest(yaml);
      const output = response.output || response.error || "";
      setResult({ success: response.success, output });
      if (response.success) {
        toast.success("Service exposed", { description: `${target.namespace}/${serviceName}` });
        onSuccess();
      } else {
        toast.error("Failed to expose service");
      }
    } catch (error) {
      const message = handleAPIError(error, "expose service");
      toast.error(message);
      setResult({ success: false, output: message });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between p-6 border-b border-slate-200">
          <div className="flex items-center space-x-3">
            <Share2 className="w-5 h-5 text-blue-600" />
            <div>
              <h3 className="text-lg font-semibold text-slate-800">Expose via NodePort</h3>
              <p className="text-xs text-slate-500">
                {target.kind} {target.namespace}/{target.name}
              </p>
            </div>
          </div>
          <button
            onClick={handleClose}
            disabled={loading}
            className="text-slate-400 hover:text-slate-600 transition-colors disabled:opacity-50"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6 space-y-4 flex-1 overflow-auto">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Service Name</label>
              <input
                value={serviceName}
                onChange={(e) => setServiceName(sanitizeName(e.target.value))}
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Namespace</label>
              <input
                value={target.namespace}
                disabled
                className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm bg-slate-100 text-slate-600"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Selector Key</label>
              <input
                value={selectorKey}
                onChange={(e) => setSelectorKey(e.target.value)}
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Selector Value</label>
              <input
                value={selectorValue}
                onChange={(e) => setSelectorValue(e.target.value)}
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Service Port</label>
              <input
                value={servicePort}
                onChange={(e) => setServicePort(e.target.value)}
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Target Port</label>
              <input
                value={targetPort}
                onChange={(e) => setTargetPort(e.target.value)}
                disabled={loading}
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">NodePort (optional)</label>
              <input
                value={nodePort}
                onChange={(e) => setNodePort(e.target.value)}
                disabled={loading}
                placeholder="30080"
                className="w-full px-3 py-2 border border-slate-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">Preview YAML</label>
            <textarea
              value={yaml}
              readOnly
              className="w-full h-48 px-3 py-2 border border-slate-200 rounded-lg font-mono text-xs bg-slate-50 text-slate-700 resize-none"
            />
          </div>

          {result && (
            <div className={`rounded-lg p-4 ${result.success ? "bg-green-50 border border-green-200" : "bg-red-50 border border-red-200"}`}>
              <div className="flex items-start space-x-3">
                {result.success ? (
                  <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                ) : (
                  <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                )}
                <div className="flex-1">
                  <p className={`text-sm font-semibold ${result.success ? "text-green-800" : "text-red-800"}`}>
                    {result.success ? "Success" : "Failed"}
                  </p>
                  <pre className={`text-xs mt-2 whitespace-pre-wrap font-mono ${result.success ? "text-green-700" : "text-red-700"}`}>
                    {result.output}
                  </pre>
                </div>
              </div>
            </div>
          )}

          <p className="text-xs text-slate-500">
            This creates a NodePort service. If NodePort is left blank, Kubernetes will assign one automatically.
          </p>
        </div>

        <div className="flex justify-end space-x-3 p-6 border-t border-slate-200">
          <button
            onClick={handleClose}
            disabled={loading}
            className="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-800 rounded-lg transition-colors disabled:opacity-50"
          >
            {result?.success ? "Close" : "Cancel"}
          </button>
          {!result?.success && (
            <button
              onClick={handleApply}
              disabled={loading}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 flex items-center space-x-2"
            >
              {loading ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Applying...</span>
                </>
              ) : (
                <>
                  <Share2 className="w-4 h-4" />
                  <span>Expose</span>
                </>
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
