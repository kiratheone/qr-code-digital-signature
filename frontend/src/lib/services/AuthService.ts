/**
 * Authentication Service - Business Logic Layer
 * Handles all authentication-related operations including login, logout, and session management
 */

import { ApiClient } from '@/lib/api';
import type {
  User,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  AuthState,
} from '@/lib/types';

export class AuthService {
  private static readonly TOKEN_KEY = 'auth_token';
  private static readonly USER_KEY = 'auth_user';

  constructor(private apiClient: ApiClient) {
    // Initialize token from localStorage on service creation
    this.initializeFromStorage();
  }

  /**
   * Login with username and password
   */
  async login(username: string, password: string): Promise<LoginResponse> {
    // Validate input
    if (!username.trim()) {
      throw new Error('Username is required');
    }
    if (!password.trim()) {
      throw new Error('Password is required');
    }

    const loginData: LoginRequest = {
      username: username.trim(),
      password,
    };

    const response = await this.apiClient.post<LoginResponse>('/auth/login', loginData);

    // Store authentication data
    this.storeAuthData(response.token, response.user);

    return response;
  }

  /**
   * Register a new user account
   */
  async register(
    username: string,
    password: string,
    fullName: string,
    email: string
  ): Promise<RegisterResponse> {
    // Validate input
    if (!username.trim()) {
      throw new Error('Username is required');
    }
    if (!password.trim()) {
      throw new Error('Password is required');
    }
    if (!fullName.trim()) {
      throw new Error('Full name is required');
    }
    if (!email.trim()) {
      throw new Error('Email is required');
    }
    if (!this.isValidEmail(email)) {
      throw new Error('Please enter a valid email address');
    }

    const registerData: RegisterRequest = {
      username: username.trim(),
      password,
      full_name: fullName.trim(),
      email: email.trim().toLowerCase(),
    };

    const response = await this.apiClient.post<RegisterResponse>('/auth/register', registerData);

    // Store authentication data
    this.storeAuthData(response.token, response.user);

    return response;
  }

  /**
   * Logout current user
   */
  async logout(): Promise<void> {
    try {
      // Call logout endpoint to invalidate server-side session
      await this.apiClient.post<void>('/auth/logout');
    } catch (error) {
      // Continue with logout even if server call fails
      console.warn('Server logout failed:', error);
    } finally {
      // Always clear local authentication data
      this.clearAuthData();
    }
  }

  /**
   * Get current authentication state
   */
  getAuthState(): AuthState {
    const token = this.getStoredToken();
    const user = this.getStoredUser();

    return {
      user,
      token,
      isAuthenticated: !!(token && user),
      isLoading: false,
    };
  }

  /**
   * Check if user is currently authenticated
   */
  isAuthenticated(): boolean {
    const token = this.getStoredToken();
    const user = this.getStoredUser();
    return !!(token && user);
  }

  /**
   * Get current user
   */
  getCurrentUser(): User | null {
    return this.getStoredUser();
  }

  /**
   * Get current token
   */
  getCurrentToken(): string | null {
    return this.getStoredToken();
  }

  /**
   * Validate session with server
   */
  async validateSession(): Promise<User> {
    const response = await this.apiClient.get<{ user: User }>('/auth/me');
    
    // Update stored user data
    if (response.user) {
      this.storeUser(response.user);
    }

    return response.user;
  }

  /**
   * Initialize authentication from localStorage
   */
  private initializeFromStorage(): void {
    const token = this.getStoredToken();
    if (token) {
      this.apiClient.setToken(token);
    }
  }

  /**
   * Store authentication data in localStorage and API client
   */
  private storeAuthData(token: string, user: User): void {
    localStorage.setItem(AuthService.TOKEN_KEY, token);
    localStorage.setItem(AuthService.USER_KEY, JSON.stringify(user));
    this.apiClient.setToken(token);
  }

  /**
   * Store user data in localStorage
   */
  private storeUser(user: User): void {
    localStorage.setItem(AuthService.USER_KEY, JSON.stringify(user));
  }

  /**
   * Clear all authentication data
   */
  private clearAuthData(): void {
    localStorage.removeItem(AuthService.TOKEN_KEY);
    localStorage.removeItem(AuthService.USER_KEY);
    this.apiClient.setToken(null);
  }

  /**
   * Get stored token from localStorage
   */
  private getStoredToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(AuthService.TOKEN_KEY);
  }

  /**
   * Get stored user from localStorage
   */
  private getStoredUser(): User | null {
    if (typeof window === 'undefined') return null;
    
    const userJson = localStorage.getItem(AuthService.USER_KEY);
    if (!userJson) return null;

    try {
      return JSON.parse(userJson);
    } catch {
      // Clear invalid user data
      localStorage.removeItem(AuthService.USER_KEY);
      return null;
    }
  }

  /**
   * Validate email format
   */
  private isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  /**
   * Validate password strength (basic validation)
   */
  validatePassword(password: string): { isValid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (password.length < 8) {
      errors.push('Password must be at least 8 characters long');
    }

    if (!/[A-Z]/.test(password)) {
      errors.push('Password must contain at least one uppercase letter');
    }

    if (!/[a-z]/.test(password)) {
      errors.push('Password must contain at least one lowercase letter');
    }

    if (!/\d/.test(password)) {
      errors.push('Password must contain at least one number');
    }

    return {
      isValid: errors.length === 0,
      errors,
    };
  }

  /**
   * Validate username format
   */
  validateUsername(username: string): { isValid: boolean; error?: string } {
    if (!username.trim()) {
      return { isValid: false, error: 'Username is required' };
    }

    if (username.length < 3) {
      return { isValid: false, error: 'Username must be at least 3 characters long' };
    }

    if (username.length > 50) {
      return { isValid: false, error: 'Username must be less than 50 characters' };
    }

    if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
      return { isValid: false, error: 'Username can only contain letters, numbers, hyphens, and underscores' };
    }

    return { isValid: true };
  }
}