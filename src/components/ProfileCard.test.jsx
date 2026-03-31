import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ProfileCard from './ProfileCard';

describe('ProfileCard Component', () => {
  const mockProfile = {
    email: 'test@example.com',
    name: 'Test User',
    bio: 'Test bio',
    avatar_url: 'https://example.com/avatar.jpg',
    role: 'user',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  };

  const mockOnUpdate = jest.fn();
  const mockOnRefresh = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('View Mode', () => {
    it('should render profile information', () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.getByText('Test User')).toBeInTheDocument();
      expect(screen.getByText('test@example.com')).toBeInTheDocument();
      expect(screen.getByText('Test bio')).toBeInTheDocument();
      expect(screen.getByText('user')).toBeInTheDocument();
    });

    it('should render avatar image', () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      const avatar = screen.getByAltText("Test User's avatar");
      expect(avatar).toBeInTheDocument();
      expect(avatar).toHaveAttribute('src', mockProfile.avatar_url);
    });

    it('should render timestamps', () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.getByText(/Created:/)).toBeInTheDocument();
      expect(screen.getByText(/Updated:/)).toBeInTheDocument();
    });

    it('should render Edit Profile button', () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      const editButton = screen.getByRole('button', { name: /Edit profile/i });
      expect(editButton).toBeInTheDocument();
    });

    it('should not render bio if not provided', () => {
      const profileWithoutBio = { ...mockProfile, bio: '' };

      render(
        <ProfileCard
          profile={profileWithoutBio}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.queryByText('Test bio')).not.toBeInTheDocument();
    });
  });

  describe('Edit Mode', () => {
    it('should switch to edit mode when Edit button is clicked', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      const editButton = screen.getByRole('button', { name: /Edit profile/i });
      fireEvent.click(editButton);

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test User')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Test bio')).toBeInTheDocument();
      });
    });

    it('should populate form fields with profile data', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      await waitFor(() => {
        expect(screen.getByDisplayValue('Test User')).toBeInTheDocument();
        expect(screen.getByDisplayValue('Test bio')).toBeInTheDocument();
        expect(screen.getByDisplayValue(mockProfile.avatar_url)).toBeInTheDocument();
      });
    });

    it('should update form fields when user types', async () => {
      const user = userEvent.setup();

      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const nameInput = await screen.findByDisplayValue('Test User');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Name');

      expect(nameInput).toHaveValue('Updated Name');
    });

    it('should validate required name field', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const nameInput = await screen.findByDisplayValue('Test User');
      fireEvent.change(nameInput, { target: { value: '' } });

      const saveButton = screen.getByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        expect(screen.getByText('Name is required')).toBeInTheDocument();
      });
    });

    it('should validate name length', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const nameInput = await screen.findByDisplayValue('Test User');
      const longName = 'a'.repeat(101);
      fireEvent.change(nameInput, { target: { value: longName } });

      const saveButton = screen.getByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        expect(screen.getByText(/Name must be less than 100 characters/)).toBeInTheDocument();
      });
    });

    it('should validate bio length', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const bioInput = await screen.findByDisplayValue('Test bio');
      const longBio = 'a'.repeat(501);
      fireEvent.change(bioInput, { target: { value: longBio } });

      const saveButton = screen.getByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        expect(screen.getByText(/Bio must be less than 500 characters/)).toBeInTheDocument();
      });
    });

    it('should validate avatar URL format', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const avatarInput = await screen.findByDisplayValue(mockProfile.avatar_url);
      fireEvent.change(avatarInput, { target: { value: 'not-a-url' } });

      const saveButton = screen.getByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        expect(screen.getByText(/Please enter a valid URL/)).toBeInTheDocument();
      });
    });

    it('should call onUpdate with form data on submit', async () => {
      mockOnUpdate.mockResolvedValueOnce(true);

      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const nameInput = await screen.findByDisplayValue('Test User');
      fireEvent.change(nameInput, { target: { value: 'Updated Name' } });

      const saveButton = screen.getByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        expect(mockOnUpdate).toHaveBeenCalledWith({
          name: 'Updated Name',
          bio: 'Test bio',
          avatar_url: mockProfile.avatar_url,
        });
      });
    });

    it('should cancel edit and restore original values', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const nameInput = await screen.findByDisplayValue('Test User');
      fireEvent.change(nameInput, { target: { value: 'Updated Name' } });

      const cancelButton = screen.getByRole('button', { name: /Cancel/i });
      fireEvent.click(cancelButton);

      await waitFor(() => {
        expect(screen.getByText('Test User')).toBeInTheDocument();
        expect(screen.queryByDisplayValue('Updated Name')).not.toBeInTheDocument();
      });
    });
  });

  describe('Loading State', () => {
    it('should render loading skeleton', () => {
      render(
        <ProfileCard
          profile={null}
          isLoading={true}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.getByRole('status')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should render error message', () => {
      render(
        <ProfileCard
          profile={null}
          isLoading={false}
          isUpdating={false}
          error="Failed to load profile"
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.getByText('Unable to Load Profile')).toBeInTheDocument();
      expect(screen.getByText('Failed to load profile')).toBeInTheDocument();
    });

    it('should render retry button on error', () => {
      render(
        <ProfileCard
          profile={null}
          isLoading={false}
          isUpdating={false}
          error="Failed to load profile"
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      const retryButton = screen.getByRole('button', { name: /Retry/i });
      expect(retryButton).toBeInTheDocument();
    });

    it('should call onRefresh when retry button is clicked', () => {
      render(
        <ProfileCard
          profile={null}
          isLoading={false}
          isUpdating={false}
          error="Failed to load profile"
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      const retryButton = screen.getByRole('button', { name: /Retry/i });
      fireEvent.click(retryButton);

      expect(mockOnRefresh).toHaveBeenCalled();
    });

    it('should show error banner when profile exists but error occurs', () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error="Update failed"
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.getByText('Update failed')).toBeInTheDocument();
    });
  });

  describe('Success Message', () => {
    it('should show success message after update', async () => {
      mockOnUpdate.mockResolvedValueOnce(true);

      const { rerender } = render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const saveButton = await screen.findByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        expect(mockOnUpdate).toHaveBeenCalled();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA labels', () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      expect(screen.getByRole('region', { name: /User profile/i })).toBeInTheDocument();
    });

    it('should have proper form labels', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      await waitFor(() => {
        expect(screen.getByLabelText(/Name/)).toBeInTheDocument();
        expect(screen.getByLabelText(/Bio/)).toBeInTheDocument();
        expect(screen.getByLabelText(/Avatar URL/)).toBeInTheDocument();
      });
    });

    it('should have proper error descriptions', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      const nameInput = await screen.findByDisplayValue('Test User');
      fireEvent.change(nameInput, { target: { value: '' } });

      const saveButton = screen.getByRole('button', { name: /Save Changes/i });
      fireEvent.click(saveButton);

      await waitFor(() => {
        const errorElement = screen.getByText('Name is required');
        expect(errorElement).toHaveAttribute('id', 'name-error');
      });
    });
  });

  describe('Character Counter', () => {
    it('should display character count for bio', async () => {
      render(
        <ProfileCard
          profile={mockProfile}
          isLoading={false}
          isUpdating={false}
          error={null}
          onUpdate={mockOnUpdate}
          onRefresh={mockOnRefresh}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /Edit profile/i }));

      await waitFor(() => {
        expect(screen.getByText(/\/500/)).toBeInTheDocument();
      });
    });
  });
});
