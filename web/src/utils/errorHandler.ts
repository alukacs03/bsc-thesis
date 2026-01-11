import { APIError } from '@/services/api/client';

export function handleAPIError(error: unknown, _context?: string): string {
  if (error instanceof APIError) {
    // Handle specific HTTP status codes
    switch (error.status) {
      case 401:
        // Session expired - redirect handled by caller if needed
        return 'Session expired. Please log in again.';
      case 403:
        return 'You do not have permission to perform this action.';
      case 404:
        return 'The requested resource was not found.';
      case 409:
        return 'This action conflicts with existing data.';
      case 500:
        return 'Server error. Please try again later.';
      default:
        return error.message || 'An unexpected error occurred.';
    }
  }

  if (error instanceof Error) {
    return error.message;
  }

  return 'An unknown error occurred.';
}
