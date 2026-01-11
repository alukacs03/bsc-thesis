// Base configuration and error handling
export class APIError extends Error {
  status: number;
  data?: unknown;

  constructor(
    message: string,
    status: number,
    data?: unknown
  ) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.data = data;
  }
}

export interface APIResponse<T> {
  data?: T;
  error?: string;
}

class APIClient {
  private baseURL = import.meta.env.VITE_API_BASE_URL || '/api';
  private debug = import.meta.env.VITE_API_DEBUG === 'true';

  constructor() {
    if (this.debug) {
      console.debug('API Client initialized with base URL:', this.baseURL);
    }
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit
  ): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    if (this.debug) {
      console.debug('Making request to:', url);
    }

    const config: RequestInit = {
      credentials: 'include', // Always send cookies for JWT
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new APIError(
          errorData.error || `HTTP ${response.status}`,
          response.status,
          errorData
        );
      }

      if (response.status === 204) {
        return {} as T;
      }

      const bodyText = await response.text();
      if (!bodyText) {
        return {} as T;
      }
      try {
        return JSON.parse(bodyText) as T;
      } catch (err) {
        throw new APIError('Invalid JSON response', response.status, bodyText);
      }
    } catch (error) {
      if (error instanceof APIError) {
        throw error;
      }
      throw new APIError(
        error instanceof Error ? error.message : 'Network error',
        0
      );
    }
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  }
}

export const apiClient = new APIClient();
