export const getStatusColor = (status: string): string => {
  switch (status) {
    case 'Active':
    case 'Running':
    case 'Ready':
      return 'bg-green-100 text-green-800';
    case 'Inactive':
    case 'NotReady':
      return 'bg-yellow-100 text-yellow-800';
    case 'Error':
    case 'Unknown':
      return 'bg-red-100 text-red-800';
    default: {
      return 'bg-gray-100 text-gray-800';
    }
  }
};