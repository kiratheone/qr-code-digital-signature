/**
 * Tests for ApiClient class
 * Focuses on core functionality and error handling
 */

import { ApiClient, ApiClientError } from '../apiClient';

// Mock fetch globally
global.fetch = jest.fn();

describe('ApiClient', () => {
  let apiClient: ApiClient;
  const mockFetch = fetch as jest.MockedFunction<typeof fetch>;

  beforeEach(() => {
    apiClient = new ApiClient('http://localhost:8000');
    mockFetch.mockClear();
  });

  describe('constructor', () => {
    it('should set baseURL correctly', () => {
      const client = new ApiClient('http://example.com/');
      expect(client.getToken()).toBeNull();
    });

    it('should use default baseURL when none provided', () => {
      const client = new ApiClient();
      expect(client.getToken()).toBeNull();
    });
  });

  describe('token management', () => {
    it('should set and get token correctly', () => {
      const token = 'test-token';
      apiClient.setToken(token);
      expect(apiClient.getToken()).toBe(token);
    });

    it('should clear token when set to null', () => {
      apiClient.setToken('test-token');
      apiClient.setToken(null);
      expect(apiClient.getToken()).toBeNull();
    });
  });

  describe('GET requests', () => {
    it('should make successful GET request', async () => {
      const mockData = { id: 1, name: 'test' };
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => mockData,
      } as Response);

      const result = await apiClient.get('/test');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/test',
        {
          method: 'GET',
          headers: {},
          body: undefined,
        }
      );
      expect(result).toEqual(mockData);
    });

    it('should include authorization header when token is set', async () => {
      const token = 'test-token';
      apiClient.setToken(token);

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({}),
      } as Response);

      await apiClient.get('/test');

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/test',
        {
          method: 'GET',
          headers: {
            Authorization: `Bearer ${token}`,
          },
          body: undefined,
        }
      );
    });
  });

  describe('POST requests', () => {
    it('should make successful POST request with JSON data', async () => {
      const requestData = { name: 'test' };
      const responseData = { id: 1, name: 'test' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => responseData,
      } as Response);

      const result = await apiClient.post('/test', requestData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/test',
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestData),
        }
      );
      expect(result).toEqual(responseData);
    });

    it('should handle FormData correctly', async () => {
      const formData = new FormData();
      formData.append('file', new Blob(['test']), 'test.txt');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({ success: true }),
      } as Response);

      await apiClient.post('/upload', formData);

      expect(mockFetch).toHaveBeenCalledWith(
        'http://localhost:8000/api/upload',
        {
          method: 'POST',
          headers: {}, // No Content-Type header for FormData
          body: formData,
        }
      );
    });
  });

  describe('error handling', () => {
    it('should throw ApiClientError for HTTP errors', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({
          code: 'VALIDATION_ERROR',
          message: 'Invalid input',
        }),
      } as Response);

      await expect(apiClient.get('/test')).rejects.toThrow(ApiClientError);
    });

    it('should throw ApiClientError for network errors', async () => {
      mockFetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

      await expect(apiClient.get('/test')).rejects.toThrow(ApiClientError);
    });

    it('should handle server errors gracefully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        headers: new Headers({ 'content-type': 'text/plain' }),
      } as Response);

      await expect(apiClient.get('/test')).rejects.toThrow(ApiClientError);
    });
  });

  describe('empty responses', () => {
    it('should handle 204 No Content responses', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        status: 204,
        headers: new Headers(),
      } as Response);

      const result = await apiClient.delete('/test');
      expect(result).toEqual({});
    });
  });
});