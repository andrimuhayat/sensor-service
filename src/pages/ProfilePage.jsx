import React, { useEffect, useState, useCallback } from 'react';
import useProfile from '../hooks/useProfile';
import ProfileCard from '../components/ProfileCard';
import './ProfilePage.css';

/**
 * ProfilePage Component
 * Main page component that fetches and displays user profile
 * Integrates with backend API and WebSocket for real-time updates
 * 
 * Handles:
 * - User authentication and email extraction
 * - Profile data fetching and caching
 * - Real-time updates via WebSocket
 * - Error handling and user feedback
 * - Responsive layout for mobile and desktop
 */
export const ProfilePage = () => {
  const [userEmail, setUserEmail] = useState(null);
  const [authError, setAuthError] = useState(null);
  const [toastMessage, setToastMessage] = useState('');
  const [toastType, setToastType] = useState('info'); // 'info', 'success', 'error'

  // Extract user email from token or context
  useEffect(() => {
    try {
      // Try to get email from localStorage (set during login)
      const storedEmail = localStorage.getItem('userEmail');
      if (storedEmail) {
        setUserEmail(storedEmail);
        return;
      }

      // Try to decode JWT token to get email
      const token = localStorage.getItem('token');
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          if (payload.email) {
            setUserEmail(payload.email);
            return;
          }
        } catch (err) {
          console.error('Failed to decode token:', err);
        }
      }

      setAuthError('Unable to determine user email. Please log in again.');
    } catch (err) {
      console.error('Auth error:', err);
      setAuthError('Authentication error occurred');
    }
  }, []);

  // Use custom hook for profile management
  const {
    profile,
    loading,
    error,
    isUpdating,
    fetchProfile,
    updateProfile,
    closeWebSocket,
  } = useProfile(userEmail);

  // Show toast notification
  const showToast = useCallback((message, type = 'info') => {
    setToastMessage(message);
    setToastType(type);
    setTimeout(() => setToastMessage(''), 4000);
  }, []);

  // Handle profile update with feedback
  const handleProfileUpdate = useCallback(
    async (updates) => {
      try {
        const success = await updateProfile(updates);
        if (success) {
          showToast('Profile updated successfully!', 'success');
          return true;
        } else {
          showToast('Failed to update profile', 'error');
          return false;
        }
      } catch (err) {
        console.error('Update error:', err);
        showToast('An error occurred while updating profile', 'error');
        return false;
      }
    },
    [updateProfile, showToast]
  );

  // Handle profile refresh
  const handleRefresh = useCallback(() => {
    fetchProfile();
    showToast('Profile refreshed', 'info');
  }, [fetchProfile, showToast]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      closeWebSocket();
    };
  }, [closeWebSocket]);

  // Authentication error state
  if (authError) {
    return (
      <div className="profile-page profile-page--error">
        <div className="profile-page__container">
          <div className="profile-page__error-content" role="alert">
            <h1 className="profile-page__error-title">Authentication Required</h1>
            <p className="profile-page__error-message">{authError}</p>
            <a
              href="/login"
              className="profile-page__button profile-page__button--primary"
              aria-label="Go to login page"
            >
              Go to Login
            </a>
          </div>
        </div>
      </div>
    );
  }

  // Loading initial state
  if (!userEmail) {
    return (
      <div className="profile-page profile-page--loading">
        <div className="profile-page__container">
          <div className="profile-page__spinner" role="status" aria-live="polite">
            <div className="spinner" aria-hidden="true"></div>
            <p className="sr-only">Loading profile page...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="profile-page">
      {/* Header */}
      <header className="profile-page__header">
        <div className="profile-page__container">
          <h1 className="profile-page__title">My Profile</h1>
          <p className="profile-page__subtitle">Manage your account information</p>
        </div>
      </header>

      {/* Main Content */}
      <main className="profile-page__main">
        <div className="profile-page__container">
          {/* Toast Notification */}
          {toastMessage && (
            <div
              className={`profile-page__toast profile-page__toast--${toastType}`}
              role="status"
              aria-live="polite"
              aria-atomic="true"
            >
              <div className="profile-page__toast-content">
                <span className="profile-page__toast-icon" aria-hidden="true">
                  {toastType === 'success' && '✓'}
                  {toastType === 'error' && '✕'}
                  {toastType === 'info' && 'ℹ'}
                </span>
                <span className="profile-page__toast-message">{toastMessage}</span>
              </div>
            </div>
          )}

          {/* Profile Card */}
          <div className="profile-page__card-wrapper">
            <ProfileCard
              profile={profile}
              isLoading={loading}
              isUpdating={isUpdating}
              error={error}
              onUpdate={handleProfileUpdate}
              onRefresh={handleRefresh}
            />
          </div>

          {/* Real-time Status Indicator */}
          {profile && (
            <div className="profile-page__status" aria-live="polite">
              <span className="profile-page__status-indicator" aria-hidden="true"></span>
              <span className="profile-page__status-text">
                Real-time updates enabled
              </span>
            </div>
          )}
        </div>
      </main>

      {/* Footer */}
      <footer className="profile-page__footer">
        <div className="profile-page__container">
          <p className="profile-page__footer-text">
            Last updated: {profile?.updated_at ? new Date(profile.updated_at).toLocaleString() : 'Never'}
          </p>
        </div>
      </footer>
    </div>
  );
};

export default ProfilePage;
