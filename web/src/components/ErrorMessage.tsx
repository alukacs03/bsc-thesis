import { AlertTriangle } from 'lucide-react';

export function ErrorMessage({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="p-6 bg-red-50 border border-red-200 rounded-lg">
      <div className="flex items-center space-x-3">
        <AlertTriangle className="w-6 h-6 text-red-600" />
        <div className="flex-1">
          <p className="text-red-800">{message}</p>
        </div>
        {onRetry && (
          <button
            onClick={onRetry}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            Retry
          </button>
        )}
      </div>
    </div>
  );
}
