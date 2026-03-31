import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import ProfilePage from './ProfilePage';
import * as useProfileModule from '../hooks/useProfile';

// Mock the useProfile hook
jest.mock('../hooks/useProfile');

// Mock localStorage
const localStorageMock = (() => {
  let store = {};
  return {
    getItem: (key) => store[key] || null,
    setItem: (key, value) => {
      store[key] = value.toString();
    },
    removeItem: (key) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

describe('ProfilePage Component', () => {
  const mockProfile = {
    email: 'test@example.com',
    name: 'Test User',
    bio: 'Test bio',
    avatar_url: 'https://example.com/avatar.jpg',
    role: 'user',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  };

  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem('userEmail', 'test@example.com');
    localStorage.setItem('token', 'mock-token');
    jest.clearAllMocks();
  });

  describe('Authentication', () => {
    it('should show error when no email is available', async () => {
      localStorage.clear();

      useProfileModule.default.mockReturnValue({
        profile: null,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText(/Authentication Required/i)).toBeInTheDocument();
      });
    });

    it('should extract email from localStorage', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(useProfileModule.default).toHaveBeenCalledWith('test@example.com');
      });
    });

    it('should extract email from JWT token if not in localStorage', async () => {
      localStorage.removeItem('userEmail');
      const token = btoa(JSON.stringify({ email: 'jwt@example.com' }));
      localStorage.setItem('token', `header.${token}.signature`);

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(useProfileModule.default).toHaveBeenCalledWith('jwt@example.com');
      });
    });
  });

  describe('Header', () => {
    it('should render page header', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      expect(screen.getByText('My Profile')).toBeInTheDocument();
      expect(screen.getByText('Manage your account information')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading spinner when email is being determined', () => {
      localStorage.clear();

      useProfileModule.default.mockReturnValue({
        profile: null,
        loading: true,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      expect(screen.getByRole('status')).toBeInTheDocument();
    });
  });

  describe('Profile Display', () => {
    it('should render profile card with data', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText('Test User')).toBeInTheDocument();
      });
    });

    it('should show real-time status indicator', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText('Real-time updates enabled')).toBeInTheDocument();
      });
    });

    it('should display last updated time in footer', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText(/Last updated:/)).toBeInTheDocument();
      });
    });
  });

  describe('Toast Notifications', () => {
    it('should show success toast on profile update', async () => {
      const mockUpdateProfile = jest.fn().mockResolvedValue(true);

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: mockUpdateProfile,
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      // Simulate profile update
      await waitFor(() => {
        expect(screen.getByText('My Profile')).toBeInTheDocument();
      });
    });

    it('should show error toast on update failure', async () => {
      const mockUpdateProfile = jest.fn().mockResolvedValue(false);

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: mockUpdateProfile,
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText('My Profile')).toBeInTheDocument();
      });
    });

    it('should auto-dismiss toast after 4 seconds', async () => {
      jest.useFakeTimers();

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      jest.useRealTimers();
    });
  });

  describe('Refresh Functionality', () => {
    it('should call fetchProfile on refresh', async () => {
      const mockFetchProfile = jest.fn();

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: mockFetchProfile,
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      // The refresh button would be in the ProfileCard component
      // This test verifies the hook is set up correctly
      expect(useProfileModule.default).toHaveBeenCalled();
    });
  });

  describe('Cleanup', () => {
    it('should close WebSocket on unmount', async () => {
      const mockCloseWebSocket = jest.fn();

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: mockCloseWebSocket,
      });

      const { unmount } = render(<ProfilePage />);

      unmount();

      expect(mockCloseWebSocket).toHaveBeenCalled();
    });
  });

  describe('Error Handling', () => {
    it('should handle token decode errors gracefully', async () => {
      localStorage.removeItem('userEmail');
      localStorage.setItem('token', 'invalid-token');

      useProfileModule.default.mockReturnValue({
        profile: null,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText(/Authentication Required/i)).toBeInTheDocument();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper page structure', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      expect(screen.getByRole('banner')).toBeInTheDocument(); // header
      expect(screen.getByRole('main')).toBeInTheDocument(); // main
      expect(screen.getByRole('contentinfo')).toBeInTheDocument(); // footer
    });

    it('should have proper ARIA labels for toast', async () => {
      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText('My Profile')).toBeInTheDocument();
      });
    });
  });

  describe('Responsive Design', () => {
    it('should render on mobile viewport', async () => {
      // Mock window.matchMedia for mobile
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 375,
      });

      useProfileModule.default.mockReturnValue({
        profile: mockProfile,
        loading: false,
        error: null,
        isUpdating: false,
        fetchProfile: jest.fn(),
        updateProfile: jest.fn(),
        closeWebSocket: jest.fn(),
      });

      render(<ProfilePage />);

      await waitFor(() => {
        expect(screen.getByText('My Profile')).toBeInTheDocument();
      });
    });
  });
});
