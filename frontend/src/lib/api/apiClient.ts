/**
 * Base API client for handling HTTP requests with simple error handling
 * Provides centralized configuration for API calls with authentication support
 */

export interface ApiResponse<T = any> {
  data: T;
  message?: string;
  success: boolean;
}

export interface ApiError {
  code: string;
  message: string;
  details?: string;
}

export class ApiClientError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
    public details?: string
  ) {
    super(message);
    this.name = 'ApiClientError';
  }
}

export class ApiClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000') {
    this.baseURL = baseURL.replace(/\/$/, ''); // Remove trailing slash
  }

  /**
   * Set authentication token for subsequent requests
   */
  setToken(token: string | null): void {
    this.token = token;
  }

  /**
   * Get current authentication token
   */
  getToken(): string | null {
    return this.token;
  }

  /**
   * Perform GET request
   */
  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>('GET', endpoint);
  }

  /**
   * Perform POST request
   */
  async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>('POST', endpoint, data);
  }

  /**
   * Perform PUT request
   */
  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>('PUT', endpoint, data);
  }

  /**
   * Perform DELETE request
   */
  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>('DELETE', endpoint);
  }

  /**
   * Core request method with error handling
   */
  private async request<T>(method: string, endpoint: string, data?: any): Promise<T> {
    const url = `${this.baseURL}/api${endpoint}`;
    
    const headers: Record<string, string> = {};

    // Set authorization header if token exists
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    // Handle different data types
    let body: string | FormData | undefined;
    if (data) {
      if (data instanceof FormData) {
        body = data;
        // Don't set Content-Type for FormData, let browser set it with boundary
      } else {
        headers['Content-Type'] = 'application/json';
        body = JSON.stringify(data);
      }
    }

    try {
      const response = await fetch(url, {
        method,
        headers,
        body,
      });

      // Handle non-JSON responses (like file downloads)
      const contentType = response.headers.get('content-type');
      
      if (!response.ok) {
        // Try to parse error response
        let errorData: ApiError;
        try {
          if (contentType?.includes('application/json')) {
            errorData = await response.json();
          } else {
            errorData = {
              code: 'HTTP_ERROR',
              message: response.statusText || 'Request failed',
            };
          }
        } catch {
          errorData = {
            code: 'HTTP_ERROR',
            message: response.statusText || 'Request failed',
          };
        }

        throw new ApiClientError(
          response.status,
          errorData.code,
          errorData.message,
          errorData.details
        );
      }

      // Handle empty responses
      if (response.status === 204 || !contentType) {
        return {} as T;
      }

      // Parse JSON response
      if (contentType?.includes('application/json')) {
        const result = await response.json();
        return result.data || result; // Handle both wrapped and unwrapped responses
      }

      // Handle other response types (like file downloads)
      return response as unknown as T;

    } catch (error) {
      if (error instanceof ApiClientError) {
        throw error;
      }

      // Handle network errors
      if (error instanceof TypeError && error.message.includes('fetch')) {
        throw new ApiClientError(
          0,
          'NETWORK_ERROR',
          'Network connection failed. Please check your internet connection.',
        );
      }

      // Handle other errors
      throw new ApiClientError(
        500,
        'UNKNOWN_ERROR',
        error instanceof Error ? error.message : 'An unknown error occurred',
      );
    }
  }
}

// Create default instance
export const apiClient = new ApiClient();