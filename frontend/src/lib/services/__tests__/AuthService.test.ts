/**
 * Tests for AuthService
 * Focuses on authentication business logic and validation
 */

import { AuthService } from '../AuthService';
import { ApiClient } from '@/lib/api';
import type { LoginResponse, RegisterResponse, User } from '@/lib/types';

// Mock ApiClient
jest.mock('@/lib/api');
const MockedApiClient = ApiClient as jest.MockedClass<typeof ApiClient>;

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage });

describe('AuthService', () => {
  let authService: AuthService;
  let mockApiClient: jest.Mocked<ApiClient>;

  const mockUser: User = {
    id: '123',
    username: 'testuser',
    full_name: 'Test User',
    email: 'test@example.com',
    role: 'user',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    is_active: true,
  };

  beforeEach(() => {
    mockApiClient = new MockedApiClient() as jest.Mocked<ApiClient>;
    mockApiClient.setToken = jest.fn();
    mockApiClient.getToken = jest.fn();
    
    // Clear localStorage mocks
    mockLocalStorage.getItem.mockClear();
    mockLocalStorage.setItem.mockClear();
    mockLocalStorage.removeItem.mockClear();
    
    authService = new AuthService(mockApiClient);
  });

  describe('login', () => {
    const mockLoginResponse: LoginResponse = {
      user: mockUser,
      token: 'test-token',
      expires_at: '2024-01-02T00:00:00Z',
    };

    it('should login successfully', async () => {
      mockApiClient.post.mockResolvedValue(mockLoginResponse);

      const result = await authService.login('testuser', 'password');

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/login', {
        username: 'testuser',
        password: 'password',
      });
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('auth_token', 'test-token');
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith('auth_user', JSON.stringify(mockUser));
      expect(mockApiClient.setToken).toHaveBeenCalledWith('test-token');
      expect(result).toEqual(mockLoginResponse);
    });

    it('should validate username is required', async () => {
      await expect(
        authService.login('', 'password')
      ).rejects.toThrow('Username is required');
    });

    it('should validate password is required', async () => {
      await expect(
        authService.login('testuser', '')
      ).rejects.toThrow('Password is required');
    });

    it('should trim username', async () => {
      mockApiClient.post.mockResolvedValue(mockLoginResponse);

      await authService.login('  testuser  ', 'password');

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/login', {
        username: 'testuser',
        password: 'password',
      });
    });
  });

  describe('register', () => {
    const mockRegisterResponse: RegisterResponse = {
      user: mockUser,
      token: 'test-token',
      expires_at: '2024-01-02T00:00:00Z',
    };

    it('should register successfully', async () => {
      mockApiClient.post.mockResolvedValue(mockRegisterResponse);

      const result = await authService.register(
        'testuser',
        'password123',
        'Test User',
        'test@example.com'
      );

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/register', {
        username: 'testuser',
        password: 'password123',
        full_name: 'Test User',
        email: 'test@example.com',
      });
      expect(result).toEqual(mockRegisterResponse);
    });

    it('should validate all required fields', async () => {
      await expect(
        authService.register('', 'password', 'Test User', 'test@example.com')
      ).rejects.toThrow('Username is required');

      await expect(
        authService.register('testuser', '', 'Test User', 'test@example.com')
      ).rejects.toThrow('Password is required');

      await expect(
        authService.register('testuser', 'password', '', 'test@example.com')
      ).rejects.toThrow('Full name is required');

      await expect(
        authService.register('testuser', 'password', 'Test User', '')
      ).rejects.toThrow('Email is required');
    });

    it('should validate email format', async () => {
      await expect(
        authService.register('testuser', 'password', 'Test User', 'invalid-email')
      ).rejects.toThrow('Please enter a valid email address');
    });

    it('should normalize email to lowercase', async () => {
      mockApiClient.post.mockResolvedValue(mockRegisterResponse);

      await authService.register('testuser', 'password', 'Test User', 'TEST@EXAMPLE.COM');

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/register', {
        username: 'testuser',
        password: 'password',
        full_name: 'Test User',
        email: 'test@example.com',
      });
    });
  });

  describe('logout', () => {
    it('should logout successfully', async () => {
      mockApiClient.post.mockResolvedValue(undefined);

      await authService.logout();

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/logout');
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('auth_token');
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('auth_user');
      expect(mockApiClient.setToken).toHaveBeenCalledWith(null);
    });

    it('should clear local data even if server call fails', async () => {
      mockApiClient.post.mockRejectedValue(new Error('Server error'));

      await authService.logout();

      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('auth_token');
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('auth_user');
      expect(mockApiClient.setToken).toHaveBeenCalledWith(null);
    });
  });

  describe('getAuthState', () => {
    it('should return authenticated state when token and user exist', () => {
      mockLocalStorage.getItem.mockImplementation((key) => {
        if (key === 'auth_token') return 'test-token';
        if (key === 'auth_user') return JSON.stringify(mockUser);
        return null;
      });

      const state = authService.getAuthState();

      expect(state.isAuthenticated).toBe(true);
      expect(state.user).toEqual(mockUser);
      expect(state.token).toBe('test-token');
    });

    it('should return unauthenticated state when no token', () => {
      mockLocalStorage.getItem.mockReturnValue(null);

      const state = authService.getAuthState();

      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
      expect(state.token).toBeNull();
    });
  });

  describe('validatePassword', () => {
    it('should validate strong password', () => {
      const result = authService.validatePassword('Password123');
      
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should reject weak passwords', () => {
      const result = authService.validatePassword('weak');
      
      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Password must be at least 8 characters long');
      expect(result.errors).toContain('Password must contain at least one uppercase letter');
      expect(result.errors).toContain('Password must contain at least one number');
    });
  });

  describe('validateUsername', () => {
    it('should validate valid username', () => {
      const result = authService.validateUsername('validuser123');
      
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should reject invalid usernames', () => {
      expect(authService.validateUsername('').isValid).toBe(false);
      expect(authService.validateUsername('ab').isValid).toBe(false);
      expect(authService.validateUsername('user@name').isValid).toBe(false);
    });
  });
});