import { useState, useEffect } from 'react';
import { X, Rocket, AlertCircle, CheckCircle, Pencil, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { kubernetesAPI } from '../../services/api/kubernetes';
import { handleAPIError } from '../../utils/errorHandler';

export interface EditResourceInfo {
  namespace: string;
  kind: string;
  name: string;
}

interface DeployModalProps {
  onClose: () => void;
  onSuccess: () => void;
  editResource?: EditResourceInfo;
}

export default function DeployModal({ onClose, onSuccess, editResource }: DeployModalProps) {
  const [yaml, setYaml] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingYaml, setLoadingYaml] = useState(false);
  const [result, setResult] = useState<{ success: boolean; output: string } | null>(null);

  const isEditMode = !!editResource;

  useEffect(() => {
    if (editResource) {
      setLoadingYaml(true);
      kubernetesAPI.getResourceYAML(editResource.namespace, editResource.kind, editResource.name)
        .then((response) => {
          if (response.yaml) {
            setYaml(response.yaml);
          } else if (response.error) {
            toast.error(`Failed to load resource: ${response.error}`);
          }
        })
        .catch((error) => {
          const message = handleAPIError(error, 'load resource');
          toast.error(message);
        })
        .finally(() => {
          setLoadingYaml(false);
        });
    }
  }, [editResource]);

  const handleDeploy = async () => {
    if (!yaml.trim()) {
      toast.error('Please enter YAML content');
      return;
    }

    setLoading(true);
    setResult(null);

    try {
      const response = await kubernetesAPI.applyManifest(yaml);

      setResult({
        success: response.success,
        output: response.output || response.error || '',
      });

      if (response.success) {
        toast.success(isEditMode ? 'Resource updated successfully' : 'Manifest applied successfully');
        onSuccess();
      } else {
        toast.error('Failed to apply manifest');
      }
    } catch (error) {
      const message = handleAPIError(error, 'apply manifest');
      toast.error(message);
      setResult({
        success: false,
        output: message,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      onClose();
    }
  };

  const title = isEditMode
    ? `Edit ${editResource.kind}: ${editResource.name}`
    : 'Deploy Workload';

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-3xl w-full mx-4 max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between p-6 border-b border-slate-200">
          <div className="flex items-center space-x-3">
            {isEditMode ? (
              <Pencil className="w-5 h-5 text-blue-600" />
            ) : (
              <Rocket className="w-5 h-5 text-blue-600" />
            )}
            <h3 className="text-lg font-semibold text-slate-800">{title}</h3>
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
          {loadingYaml ? (
            <div className="flex items-center justify-center h-64">
              <Loader2 className="w-8 h-8 text-blue-600 animate-spin" />
              <span className="ml-3 text-slate-600">Loading resource...</span>
            </div>
          ) : (
            <>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  Kubernetes Manifest (YAML)
                </label>
                <textarea
                  value={yaml}
                  onChange={(e) => setYaml(e.target.value)}
                  disabled={loading}
                  className="w-full h-80 px-3 py-2 border border-slate-300 rounded-lg font-mono text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:bg-slate-100 disabled:cursor-not-allowed resize-none"
                  placeholder={isEditMode ? '' : `apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: my-app
        image: nginx:alpine`}
                />
              </div>

              <p className="text-xs text-slate-500">
                {isEditMode
                  ? 'Modify the YAML and click Apply to update the resource. Changes will be applied using kubectl apply.'
                  : 'Paste any valid Kubernetes manifest. Supports Deployments, Services, ConfigMaps, Secrets, and more. Multi-document YAML (separated by ---) is supported.'}
              </p>
            </>
          )}

          {result && (
            <div className={`rounded-lg p-4 ${result.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
              <div className="flex items-start space-x-3">
                {result.success ? (
                  <CheckCircle className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                ) : (
                  <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5" />
                )}
                <div className="flex-1">
                  <p className={`text-sm font-semibold ${result.success ? 'text-green-800' : 'text-red-800'}`}>
                    {result.success ? 'Success' : 'Failed'}
                  </p>
                  <pre className={`text-xs mt-2 whitespace-pre-wrap font-mono ${result.success ? 'text-green-700' : 'text-red-700'}`}>
                    {result.output}
                  </pre>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="flex justify-end space-x-3 p-6 border-t border-slate-200">
          <button
            onClick={handleClose}
            disabled={loading}
            className="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-800 rounded-lg transition-colors disabled:opacity-50"
          >
            {result?.success ? 'Close' : 'Cancel'}
          </button>
          {!result?.success && !loadingYaml && (
            <button
              onClick={handleDeploy}
              disabled={loading || !yaml.trim()}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 flex items-center space-x-2"
            >
              {loading ? (
                <>
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Applying...</span>
                </>
              ) : (
                <>
                  {isEditMode ? <Pencil className="w-4 h-4" /> : <Rocket className="w-4 h-4" />}
                  <span>{isEditMode ? 'Apply' : 'Deploy'}</span>
                </>
              )}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
