/**
 * Tests for Navigation component
 * Focuses on critical user interactions and authentication states
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import { Navigation } from '../Navigation';
import { useAuthOperations } from '@/hooks';

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

// Mock auth operations hook
jest.mock('@/hooks', () => ({
  useAuthOperations: jest.fn(),
}));

const mockPush = jest.fn();
const mockUseRouter = useRouter as jest.MockedFunction<typeof useRouter>;
const mockUseAuthOperations = useAuthOperations as jest.MockedFunction<typeof useAuthOperations>;

describe('Navigation', () => {
  beforeEach(() => {
    mockUseRouter.mockReturnValue({
      push: mockPush,
    } as any);
    
    mockPush.mockClear();
  });

  it('should show unauthenticated navigation when not logged in', () => {
    mockUseAuthOperations.mockReturnValue({
      isAuthenticated: false,
      user: null,
      logout: jest.fn(),
      isLoggingOut: false,
    } as any);

    render(<Navigation />);

    expect(screen.getByText('Digital Signature System')).toBeInTheDocument();
    expect(screen.getByText('Sign In')).toBeInTheDocument();
    expect(screen.getByText('Register')).toBeInTheDocument();
    expect(screen.queryByText('Documents')).not.toBeInTheDocument();
  });

  it('should show authenticated navigation when logged in', () => {
    const mockUser = {
      id: '1',
      username: 'testuser',
      full_name: 'Test User',
      email: 'test@example.com',
    };

    mockUseAuthOperations.mockReturnValue({
      isAuthenticated: true,
      user: mockUser,
      logout: jest.fn(),
      isLoggingOut: false,
    } as any);

    render(<Navigation />);

    expect(screen.getByText('Digital Signature System')).toBeInTheDocument();
    expect(screen.getByText('Documents')).toBeInTheDocument();
    expect(screen.getByText('Welcome, Test User')).toBeInTheDocument();
    expect(screen.getByText('Sign Out')).toBeInTheDocument();
    expect(screen.queryByText('Sign In')).not.toBeInTheDocument();
    expect(screen.queryByText('Register')).not.toBeInTheDocument();
  });

  it('should call logout when sign out button is clicked', () => {
    const mockLogout = jest.fn();
    const mockUser = {
      id: '1',
      username: 'testuser',
      full_name: 'Test User',
      email: 'test@example.com',
    };

    mockUseAuthOperations.mockReturnValue({
      isAuthenticated: true,
      user: mockUser,
      logout: mockLogout,
      isLoggingOut: false,
    } as any);

    render(<Navigation />);

    const signOutButton = screen.getByText('Sign Out');
    fireEvent.click(signOutButton);

    expect(mockLogout).toHaveBeenCalledTimes(1);
  });

  it('should show loading state when logging out', () => {
    const mockUser = {
      id: '1',
      username: 'testuser',
      full_name: 'Test User',
      email: 'test@example.com',
    };

    mockUseAuthOperations.mockReturnValue({
      isAuthenticated: true,
      user: mockUser,
      logout: jest.fn(),
      isLoggingOut: true,
    } as any);

    render(<Navigation />);

    expect(screen.getByText('Signing out...')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /signing out/i })).toBeDisabled();
  });

  it('should fallback to username when full_name is not available', () => {
    const mockUser = {
      id: '1',
      username: 'testuser',
      email: 'test@example.com',
    };

    mockUseAuthOperations.mockReturnValue({
      isAuthenticated: true,
      user: mockUser,
      logout: jest.fn(),
      isLoggingOut: false,
    } as any);

    render(<Navigation />);

    expect(screen.getByText('Welcome, testuser')).toBeInTheDocument();
  });
});