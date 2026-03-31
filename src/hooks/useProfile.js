import { useState, useEffect, useCallback, useRef } from 'react';

/**
 * Custom hook for managing user profile data with API integration and WebSocket support
 * Handles fetching, updating, and real-time subscription to profile changes
 * 
 * @param {string} email - User email to fetch profile for
 * @returns {Object} Profile state and methods
 */
export const useProfile = (email) => {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isUpdating, setIsUpdating] = useState(false);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  // Fetch profile data from API
  const fetchProfile = useCallback(async () => {
    if (!email) {
      setError('Email is required');
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(`/api/profile/${encodeURIComponent(email)}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch profile: ${response.statusText}`);
      }

      const data = await response.json();
      setProfile(data.data || data);
      setError(null);
    } catch (err) {
      setError(err.message || 'Failed to fetch profile');
      console.error('Profile fetch error:', err);
    } finally {
      setLoading(false);
    }
  }, [email]);

  // Update profile data via API
  const updateProfile = useCallback(async (updates) => {
    if (!email) {
      setError('Email is required');
      return false;
    }

    try {
      setIsUpdating(true);
      setError(null);

      const response = await fetch(`/api/profile/${encodeURIComponent(email)}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(updates),
      });

      if (!response.ok) {
        throw new Error(`Failed to update profile: ${response.statusText}`);
      }

      const data = await response.json();
      setProfile(data.data || data);
      return true;
    } catch (err) {
      setError(err.message || 'Failed to update profile');
      console.error('Profile update error:', err);
      return false;
    } finally {
      setIsUpdating(false);
    }
  }, [email]);

  // Setup WebSocket connection for real-time updates
  const setupWebSocket = useCallback(() => {
    if (!email || wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    try {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/ws/profile/${encodeURIComponent(email)}`;
      
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log('WebSocket connected for profile updates');
        // Send authentication token
        const token = localStorage.getItem('token');
        if (token) {
          wsRef.current.send(JSON.stringify({ type: 'auth', token }));
        }
      };

      wsRef.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          if (message.type === 'profile_update' && message.data) {
            setProfile((prev) => ({ ...prev, ...message.data }));
          }
        } catch (err) {
          console.error('WebSocket message parse error:', err);
        }
      };

      wsRef.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('Real-time connection error');
      };

      wsRef.current.onclose = () => {
        console.log('WebSocket disconnected');
        // Attempt to reconnect after 3 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          setupWebSocket();
        }, 3000);
      };
    } catch (err) {
      console.error('WebSocket setup error:', err);
      setError('Failed to establish real-time connection');
    }
  }, [email]);

  // Cleanup WebSocket on unmount
  const closeWebSocket = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
  }, []);

  // Initial fetch and WebSocket setup
  useEffect(() => {
    fetchProfile();
    setupWebSocket();

    return () => {
      closeWebSocket();
    };
  }, [email, fetchProfile, setupWebSocket, closeWebSocket]);

  return {
    profile,
    loading,
    error,
    isUpdating,
    fetchProfile,
    updateProfile,
    closeWebSocket,
  };
};

export default useProfile;
