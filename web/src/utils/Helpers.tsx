export const getStatusColor = (status: string, noBg?: boolean): string => {
  switch (status) {
    case 'Active':
    case 'Running':
    case 'Ready':
    case 'Online':
    case 'online':
      return noBg ? 'text-green-800' : 'bg-green-300 text-green-800';
    case 'Inactive':
    case 'NotReady':
    case 'Offline':
    case 'offline':
      return noBg ? 'text-gray-800' : 'bg-gray-300 text-gray-800';
    case 'Maintenance':
    case 'maintenance':
    case 'Warning':
      return noBg ? 'text-yellow-800' : 'bg-yellow-300 text-yellow-800';
    case 'Error':
    case 'Unknown':
      return noBg ? 'text-red-800' : 'bg-red-300 text-red-800';
    default: {
      return noBg ? 'text-gray-800' : 'bg-gray-100 text-gray-800';
    }
  }
};

export const getMetricColor = (value: number, type: 'cpu' | 'memory' | 'disk') => {
    if (type === 'cpu' || type === 'memory') {
      if (value > 80) return 'text-red-600';
      if (value > 60) return 'text-yellow-600';
      return 'text-green-600';
    }
    // Disk usage
    if (value > 85) return 'text-red-600';
    if (value > 70) return 'text-yellow-600';
    return 'text-green-600';
};