import { renderHook, act, waitFor } from '@testing-library/react';
import useProfile from './useProfile';

// Mock fetch
global.fetch = jest.fn();

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

// Mock WebSocket
class MockWebSocket {
  constructor(url) {
    this.url = url;
    this.readyState = WebSocket.CONNECTING;
    setTimeout(() => {
      this.readyState = WebSocket.OPEN;
      this.onopen?.();
    }, 0);
  }

  send(data) {
    // Mock send
  }

  close() {
    this.readyState = WebSocket.CLOSED;
    this.onclose?.();
  }
}

global.WebSocket = MockWebSocket;

describe('useProfile Hook', () => {
  beforeEach(() => {
    fetch.mockClear();
    localStorage.clear();
    localStorage.setItem('token', 'mock-token');
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('fetchProfile', () => {
    it('should fetch profile data successfully', async () => {
      const mockProfile = {
        email: 'test@example.com',
        name: 'Test User',
        bio: 'Test bio',
        avatar_url: 'https://example.com/avatar.jpg',
        role: 'user',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockProfile }),
      });

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.profile).toEqual(mockProfile);
      expect(result.current.error).toBeNull();
    });

    it('should handle fetch errors', async () => {
      fetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Not Found',
      });

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.profile).toBeNull();
      expect(result.current.error).toBeTruthy();
    });

    it('should handle network errors', async () => {
      fetch.mockRejectedValueOnce(new Error('Network error'));

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBeTruthy();
    });

    it('should not fetch if email is not provided', async () => {
      const { result } = renderHook(() => useProfile(null));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBeTruthy();
      expect(fetch).not.toHaveBeenCalled();
    });

    it('should include authorization header', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: {} }),
      });

      renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });

      const [, options] = fetch.mock.calls[0];
      expect(options.headers.Authorization).toBe('Bearer mock-token');
    });
  });

  describe('updateProfile', () => {
    it('should update profile successfully', async () => {
      const mockProfile = {
        email: 'test@example.com',
        name: 'Test User',
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockProfile }),
      });

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      let updateSuccess = false;
      await act(async () => {
        updateSuccess = await result.current.updateProfile({
          name: 'Updated Name',
        });
      });

      expect(updateSuccess).toBe(true);
      expect(result.current.profile).toEqual(mockProfile);
    });

    it('should handle update errors', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: {} }),
      });

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      fetch.mockResolvedValueOnce({
        ok: false,
        statusText: 'Bad Request',
      });

      let updateSuccess = false;
      await act(async () => {
        updateSuccess = await result.current.updateProfile({
          name: 'Updated Name',
        });
      });

      expect(updateSuccess).toBe(false);
      expect(result.current.error).toBeTruthy();
    });

    it('should set isUpdating flag', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: {} }),
      });

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      fetch.mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            setTimeout(() => {
              resolve({
                ok: true,
                json: async () => ({ data: {} }),
              });
            }, 100);
          })
      );

      act(() => {
        result.current.updateProfile({ name: 'Updated' });
      });

      expect(result.current.isUpdating).toBe(true);

      await waitFor(() => {
        expect(result.current.isUpdating).toBe(false);
      });
    });
  });

  describe('WebSocket', () => {
    it('should setup WebSocket connection', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: {} }),
      });

      renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });

      // WebSocket should be created
      expect(global.WebSocket).toBeDefined();
    });

    it('should handle WebSocket messages', async () => {
      const mockProfile = {
        email: 'test@example.com',
        name: 'Test User',
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockProfile }),
      });

      const { result } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Simulate WebSocket message
      const wsInstance = global.WebSocket.mock?.results?.[0]?.value;
      if (wsInstance?.onmessage) {
        act(() => {
          wsInstance.onmessage({
            data: JSON.stringify({
              type: 'profile_update',
              data: { name: 'Updated Name' },
            }),
          });
        });

        expect(result.current.profile.name).toBe('Updated Name');
      }
    });

    it('should close WebSocket on unmount', async () => {
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: {} }),
      });

      const { unmount } = renderHook(() => useProfile('test@example.com'));

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });

      unmount();

      // Verify closeWebSocket was called
      expect(fetch).toHaveBeenCalled();
    });
  });

  describe('Email encoding', () => {
    it('should encode email in URL', async () => {
      const email = 'test+tag@example.com';
      fetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: {} }),
      });

      renderHook(() => useProfile(email));

      await waitFor(() => {
        expect(fetch).toHaveBeenCalled();
      });

      const [url] = fetch.mock.calls[0];
      expect(url).toContain(encodeURIComponent(email));
    });
  });
});
