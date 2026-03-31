import React, { useState, useCallback } from 'react';
import './ProfileCard.css';

/**
 * ProfileCard Component
 * Displays user profile information with edit mode capability
 * Supports real-time updates and form validation
 * 
 * @param {Object} profile - Profile data object
 * @param {boolean} isLoading - Loading state
 * @param {boolean} isUpdating - Update in progress state
 * @param {string} error - Error message
 * @param {Function} onUpdate - Callback to update profile
 * @param {Function} onRefresh - Callback to refresh profile
 */
export const ProfileCard = ({
  profile,
  isLoading,
  isUpdating,
  error,
  onUpdate,
  onRefresh,
}) => {
  const [isEditMode, setIsEditMode] = useState(false);
  const [formData, setFormData] = useState({
    name: profile?.name || '',
    bio: profile?.bio || '',
    avatar_url: profile?.avatar_url || '',
  });
  const [validationErrors, setValidationErrors] = useState({});
  const [successMessage, setSuccessMessage] = useState('');

  // Validate form fields
  const validateForm = useCallback(() => {
    const errors = {};

    if (!formData.name || formData.name.trim().length === 0) {
      errors.name = 'Name is required';
    } else if (formData.name.length > 100) {
      errors.name = 'Name must be less than 100 characters';
    }

    if (formData.bio && formData.bio.length > 500) {
      errors.bio = 'Bio must be less than 500 characters';
    }

    if (formData.avatar_url && !isValidUrl(formData.avatar_url)) {
      errors.avatar_url = 'Please enter a valid URL';
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  }, [formData]);

  // Validate URL format
  const isValidUrl = (string) => {
    try {
      new URL(string);
      return true;
    } catch (_) {
      return false;
    }
  };

  // Handle form input changes
  const handleInputChange = useCallback((e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    // Clear validation error for this field
    if (validationErrors[name]) {
      setValidationErrors((prev) => ({
        ...prev,
        [name]: '',
      }));
    }
  }, [validationErrors]);

  // Handle form submission
  const handleSubmit = useCallback(async (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    const success = await onUpdate(formData);
    if (success) {
      setIsEditMode(false);
      setSuccessMessage('Profile updated successfully!');
      setTimeout(() => setSuccessMessage(''), 3000);
    }
  }, [formData, validateForm, onUpdate]);

  // Handle cancel edit
  const handleCancel = useCallback(() => {
    setFormData({
      name: profile?.name || '',
      bio: profile?.bio || '',
      avatar_url: profile?.avatar_url || '',
    });
    setValidationErrors({});
    setIsEditMode(false);
  }, [profile]);

  // Loading state
  if (isLoading) {
    return (
      <div className="profile-card profile-card--loading" role="status" aria-live="polite">
        <div className="profile-card__skeleton">
          <div className="skeleton skeleton--avatar" aria-hidden="true"></div>
          <div className="skeleton skeleton--text" aria-hidden="true"></div>
          <div className="skeleton skeleton--text skeleton--short" aria-hidden="true"></div>
        </div>
        <p className="sr-only">Loading profile information...</p>
      </div>
    );
  }

  // Error state
  if (error && !profile) {
    return (
      <div className="profile-card profile-card--error" role="alert">
        <div className="profile-card__error-content">
          <h2 className="profile-card__error-title">Unable to Load Profile</h2>
          <p className="profile-card__error-message">{error}</p>
          <button
            className="profile-card__button profile-card__button--primary"
            onClick={onRefresh}
            aria-label="Retry loading profile"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  // No profile data
  if (!profile) {
    return (
      <div className="profile-card profile-card--empty">
        <p className="profile-card__empty-message">No profile data available</p>
      </div>
    );
  }

  return (
    <div className="profile-card" role="region" aria-label="User profile">
      {/* Success Message */}
      {successMessage && (
        <div className="profile-card__success" role="status" aria-live="polite">
          {successMessage}
        </div>
      )}

      {/* Error Message */}
      {error && profile && (
        <div className="profile-card__error-banner" role="alert">
          {error}
        </div>
      )}

      {/* View Mode */}
      {!isEditMode ? (
        <div className="profile-card__view-mode">
          {/* Avatar */}
          {profile.avatar_url && (
            <div className="profile-card__avatar-container">
              <img
                src={profile.avatar_url}
                alt={`${profile.name}'s avatar`}
                className="profile-card__avatar"
                onError={(e) => {
                  e.target.src = 'https://via.placeholder.com/150?text=No+Image';
                }}
              />
            </div>
          )}

          {/* Profile Info */}
          <div className="profile-card__info">
            <h1 className="profile-card__name">{profile.name}</h1>
            <p className="profile-card__email">{profile.email}</p>

            {profile.bio && (
              <p className="profile-card__bio">{profile.bio}</p>
            )}

            {profile.role && (
              <div className="profile-card__role">
                <span className="profile-card__role-label">Role:</span>
                <span className="profile-card__role-value">{profile.role}</span>
              </div>
            )}

            {/* Timestamps */}
            <div className="profile-card__timestamps">
              {profile.created_at && (
                <p className="profile-card__timestamp">
                  <span className="profile-card__timestamp-label">Created:</span>
                  <time dateTime={profile.created_at}>
                    {new Date(profile.created_at).toLocaleDateString()}
                  </time>
                </p>
              )}
              {profile.updated_at && (
                <p className="profile-card__timestamp">
                  <span className="profile-card__timestamp-label">Updated:</span>
                  <time dateTime={profile.updated_at}>
                    {new Date(profile.updated_at).toLocaleDateString()}
                  </time>
                </p>
              )}
            </div>
          </div>

          {/* Edit Button */}
          <button
            className="profile-card__button profile-card__button--primary"
            onClick={() => setIsEditMode(true)}
            aria-label="Edit profile information"
            disabled={isUpdating}
          >
            Edit Profile
          </button>
        </div>
      ) : (
        /* Edit Mode */
        <form className="profile-card__edit-mode" onSubmit={handleSubmit}>
          <h2 className="profile-card__edit-title">Edit Profile</h2>

          {/* Name Field */}
          <div className="profile-card__form-group">
            <label htmlFor="name" className="profile-card__label">
              Name <span className="profile-card__required" aria-label="required">*</span>
            </label>
            <input
              id="name"
              type="text"
              name="name"
              value={formData.name}
              onChange={handleInputChange}
              className={`profile-card__input ${validationErrors.name ? 'profile-card__input--error' : ''}`}
              placeholder="Enter your name"
              maxLength="100"
              disabled={isUpdating}
              aria-invalid={!!validationErrors.name}
              aria-describedby={validationErrors.name ? 'name-error' : undefined}
            />
            {validationErrors.name && (
              <span id="name-error" className="profile-card__error-text">
                {validationErrors.name}
              </span>
            )}
          </div>

          {/* Bio Field */}
          <div className="profile-card__form-group">
            <label htmlFor="bio" className="profile-card__label">
              Bio
            </label>
            <textarea
              id="bio"
              name="bio"
              value={formData.bio}
              onChange={handleInputChange}
              className={`profile-card__textarea ${validationErrors.bio ? 'profile-card__textarea--error' : ''}`}
              placeholder="Tell us about yourself"
              maxLength="500"
              rows="4"
              disabled={isUpdating}
              aria-invalid={!!validationErrors.bio}
              aria-describedby={validationErrors.bio ? 'bio-error' : undefined}
            />
            <span className="profile-card__char-count">
              {formData.bio.length}/500
            </span>
            {validationErrors.bio && (
              <span id="bio-error" className="profile-card__error-text">
                {validationErrors.bio}
              </span>
            )}
          </div>

          {/* Avatar URL Field */}
          <div className="profile-card__form-group">
            <label htmlFor="avatar_url" className="profile-card__label">
              Avatar URL
            </label>
            <input
              id="avatar_url"
              type="url"
              name="avatar_url"
              value={formData.avatar_url}
              onChange={handleInputChange}
              className={`profile-card__input ${validationErrors.avatar_url ? 'profile-card__input--error' : ''}`}
              placeholder="https://example.com/avatar.jpg"
              disabled={isUpdating}
              aria-invalid={!!validationErrors.avatar_url}
              aria-describedby={validationErrors.avatar_url ? 'avatar_url-error' : undefined}
            />
            {validationErrors.avatar_url && (
              <span id="avatar_url-error" className="profile-card__error-text">
                {validationErrors.avatar_url}
              </span>
            )}
          </div>

          {/* Form Actions */}
          <div className="profile-card__form-actions">
            <button
              type="submit"
              className="profile-card__button profile-card__button--primary"
              disabled={isUpdating}
              aria-busy={isUpdating}
            >
              {isUpdating ? 'Saving...' : 'Save Changes'}
            </button>
            <button
              type="button"
              className="profile-card__button profile-card__button--secondary"
              onClick={handleCancel}
              disabled={isUpdating}
            >
              Cancel
            </button>
          </div>
        </form>
      )}
    </div>
  );
};

export default ProfileCard;
