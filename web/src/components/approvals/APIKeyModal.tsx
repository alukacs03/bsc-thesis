import { useState } from 'react';
import { Copy, Check, X, AlertTriangle } from 'lucide-react';

interface APIKeyModalProps {
  apiKey: string;
  nodeName: string;
  onClose: () => void;
}

export default function APIKeyModal({ apiKey, nodeName, onClose }: APIKeyModalProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(apiKey);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4">
        <div className="flex items-center justify-between p-6 border-b border-slate-200">
          <h3 className="text-lg font-semibold text-slate-800">Node Enrollment Approved</h3>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6 space-y-4">
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 flex items-start space-x-3">
            <AlertTriangle className="w-5 h-5 text-yellow-600 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-semibold text-yellow-800">Important: Copy this API key now</p>
              <p className="text-sm text-yellow-700 mt-1">
                This key will only be shown once. Make sure to copy it before closing this dialog.
              </p>
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              API Key for {nodeName}
            </label>
            <div className="flex items-center space-x-2">
              <div className="flex-1 bg-slate-50 border border-slate-200 rounded-lg p-3 font-mono text-sm text-slate-800 break-all">
                {apiKey}
              </div>
              <button
                onClick={handleCopy}
                className="px-4 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors flex items-center space-x-2"
              >
                {copied ? (
                  <>
                    <Check className="w-4 h-4" />
                    <span>Copied!</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-4 h-4" />
                    <span>Copy</span>
                  </>
                )}
              </button>
            </div>
          </div>

          <div className="bg-slate-50 border border-slate-200 rounded-lg p-4">
            <p className="text-sm font-semibold text-slate-800 mb-2">Next Steps:</p>
            <ol className="text-sm text-slate-700 space-y-2 list-decimal list-inside">
              <li>Copy the API key above and save it securely</li>
              <li>Install the Gluon agent on the node: <code className="bg-slate-800 text-slate-100 px-2 py-0.5 rounded text-xs">curl -sSL https://install.gluon.io | sh</code></li>
              <li>Configure the agent with this API key</li>
              <li>The agent will automatically connect and receive network configuration</li>
            </ol>
          </div>
        </div>

        <div className="flex justify-end p-6 border-t border-slate-200">
          <button
            onClick={onClose}
            className="px-6 py-2 bg-slate-600 hover:bg-slate-700 text-white rounded-lg transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
