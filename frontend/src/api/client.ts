import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';

// Define API error response type
export interface ApiErrorResponse {
  status: number;
  message: string;
  details?: string;
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';

// Create a custom axios instance with enhanced configuration
const apiClient: AxiosInstance = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Include cookies for authentication
  timeout: 30000, // 30 second timeout
  // Enhanced retry configuration
  validateStatus: (status) => status < 500, // Don't throw for client errors (4xx)
});

// Request interceptor for API calls
apiClient.interceptors.request.use(
  (config) => {
    // Add auth headers if available
    if (typeof window !== 'undefined') {
      // Try to get auth token from storage
      const authData = localStorage.getItem('auth') || sessionStorage.getItem('auth');
      
      if (authData) {
        try {
          const { accessToken } = JSON.parse(authData);
          if (accessToken && config.headers) {
            config.headers['Authorization'] = `Bearer ${accessToken}`;
          }
        } catch (e) {
          // Invalid auth data in storage, ignore
          console.error('Invalid auth data in storage');
        }
      }
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for API calls with enhanced error handling
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean; _retryCount?: number };
    
    // Initialize retry count
    if (!originalRequest._retryCount) {
      originalRequest._retryCount = 0;
    }
    
    // Handle network errors with exponential backoff
    if (!error.response && originalRequest._retryCount < 3) {
      originalRequest._retryCount++;
      const delay = Math.min(1000 * Math.pow(2, originalRequest._retryCount - 1), 10000);
      
      await new Promise(resolve => setTimeout(resolve, delay));
      return apiClient(originalRequest);
    }
    
    // Handle unauthorized errors (401) with token refresh
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        // Try to refresh the token
        const refreshResponse = await apiClient.post<{ accessToken: string }>('/api/auth/refresh');
        const { accessToken } = refreshResponse.data;
        
        // Update stored auth data
        if (typeof window !== 'undefined') {
          const authData = localStorage.getItem('auth') || sessionStorage.getItem('auth');
          if (authData) {
            try {
              const parsed = JSON.parse(authData);
              parsed.accessToken = accessToken;
              localStorage.setItem('auth', JSON.stringify(parsed));
            } catch (e) {
              // Ignore parsing errors
            }
          }
        }
        
        // Update the authorization header
        if (originalRequest.headers) {
          originalRequest.headers['Authorization'] = `Bearer ${accessToken}`;
        } else {
          originalRequest.headers = { 'Authorization': `Bearer ${accessToken}` };
        }
        
        // Retry the original request
        return apiClient(originalRequest);
      } catch (refreshError) {
        // If refresh token fails, redirect to login
        if (typeof window !== 'undefined') {
          // Clear any auth data from storage
          localStorage.removeItem('auth');
          sessionStorage.removeItem('auth');
          
          // Redirect to login page
          window.location.href = '/login';
        }
        
        return Promise.reject(refreshError);
      }
    }
    
    // Handle server errors (5xx) with retry
    if (error.response?.status && error.response.status >= 500 && originalRequest._retryCount < 2) {
      originalRequest._retryCount++;
      const delay = Math.min(1000 * Math.pow(2, originalRequest._retryCount - 1), 5000);
      
      await new Promise(resolve => setTimeout(resolve, delay));
      return apiClient(originalRequest);
    }
    
    // Extract error message from response data
    const responseData = error.response?.data as { message?: string; details?: string } | undefined;
    
    // Create a standardized error object
    const errorResponse: ApiErrorResponse = {
      status: error.response?.status || (error.code === 'ECONNABORTED' ? 408 : 0),
      message: responseData?.message || getErrorMessage(error),
      details: responseData?.details || error.message,
    };
    
    // Log errors in development
    if (process.env.NODE_ENV !== 'production') {
      console.error('API Error:', errorResponse);
    }
    
    return Promise.reject(errorResponse);
  }
);

// Helper function to get appropriate error message
function getErrorMessage(error: AxiosError): string {
  if (error.code === 'ECONNABORTED') {
    return 'Request timeout. Please try again.';
  }
  if (error.code === 'ERR_NETWORK') {
    return 'Network error. Please check your connection.';
  }
  if (!error.response) {
    return 'Unable to connect to server. Please try again.';
  }
  return 'An unexpected error occurred';
}

// Generic GET request function
export const get = async <T>(url: string, params?: object): Promise<T> => {
  const response: AxiosResponse<T> = await apiClient.get(url, { params });
  return response.data;
};

// Generic POST request function
export const post = async <T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
  const response: AxiosResponse<T> = await apiClient.post(url, data, config);
  return response.data;
};

// Generic PUT request function
export const put = async <T>(url: string, data?: unknown): Promise<T> => {
  const response: AxiosResponse<T> = await apiClient.put(url, data);
  return response.data;
};

// Generic DELETE request function
export const del = async <T>(url: string): Promise<T> => {
  const response: AxiosResponse<T> = await apiClient.delete(url);
  return response.data;
};

export default apiClient;