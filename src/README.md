# User Profile Components

This directory contains React components for the user profile management feature (QUBLI-7).

## Components Overview

### 1. ProfilePage (`pages/ProfilePage.jsx`)

Main page component that orchestrates the profile display and management.

**Features:**
- User authentication via JWT token or localStorage
- Profile data fetching and caching
- Real-time updates via WebSocket
- Toast notifications for user feedback
- Responsive layout for mobile and desktop
- Accessibility compliant (WCAG 2.1 AA)

**Props:** None (uses hooks for state management)

**Usage:**
```jsx
import ProfilePage from './pages/ProfilePage';

function App() {
  return <ProfilePage />;
}
```

**Authentication:**
- Extracts user email from `localStorage.userEmail`
- Falls back to JWT token decoding if email not in localStorage
- Requires valid JWT token in `localStorage.token`

### 2. ProfileCard (`components/ProfileCard.jsx`)

Reusable card component for displaying and editing profile information.

**Features:**
- View mode for displaying profile data
- Edit mode with form validation
- Real-time avatar preview
- Character counter for bio field
- Loading and error states
- Accessibility features (ARIA labels, keyboard navigation)

**Props:**
```typescript
interface ProfileCardProps {
  profile: {
    email: string;
    name: string;
    bio?: string;
    avatar_url?: string;
    role?: string;
    created_at?: string;
    updated_at?: string;
  };
  isLoading: boolean;
  isUpdating: boolean;
  error?: string;
  onUpdate: (updates: object) => Promise<boolean>;
  onRefresh: () => void;
}
```

**Usage:**
```jsx
import ProfileCard from './components/ProfileCard';

<ProfileCard
  profile={profile}
  isLoading={loading}
  isUpdating={isUpdating}
  error={error}
  onUpdate={handleUpdate}
  onRefresh={handleRefresh}
/>
```

### 3. useProfile Hook (`hooks/useProfile.js`)

Custom React hook for managing profile data with API integration and WebSocket support.

**Features:**
- Fetches profile data from API
- Updates profile via PUT request
- WebSocket connection for real-time updates
- Automatic reconnection on disconnect
- Error handling and loading states
- Token-based authentication

**Returns:**
```typescript
interface UseProfileReturn {
  profile: object | null;
  loading: boolean;
  error: string | null;
  isUpdating: boolean;
  fetchProfile: () => Promise<void>;
  updateProfile: (updates: object) => Promise<boolean>;
  closeWebSocket: () => void;
}
```

**Usage:**
```jsx
import useProfile from './hooks/useProfile';

function MyComponent() {
  const { profile, loading, error, updateProfile } = useProfile('user@example.com');
  
  return (
    // Component JSX
  );
}
```

## API Integration

### Endpoints

- **GET** `/api/profile/:email` - Fetch user profile
- **PUT** `/api/profile/:email` - Update user profile
- **WS** `/ws/profile/:email` - WebSocket for real-time updates

### Request/Response Format

**Profile Object:**
```json
{
  "email": "user@example.com",
  "name": "User Name",
  "bio": "User biography",
  "avatar_url": "https://example.com/avatar.jpg",
  "role": "user",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-02T00:00:00Z"
}
```

**Update Request:**
```json
{
  "name": "Updated Name",
  "bio": "Updated bio",
  "avatar_url": "https://example.com/new-avatar.jpg"
}
```

## Styling

### CSS Files

- `components/ProfileCard.css` - ProfileCard component styles
- `pages/ProfilePage.css` - ProfilePage component styles

### Features

- Mobile-first responsive design
- Dark mode support via `prefers-color-scheme`
- Reduced motion support via `prefers-reduced-motion`
- Accessibility-focused styling
- Smooth animations and transitions

### Responsive Breakpoints

- Desktop: 768px and above
- Tablet: 481px to 767px
- Mobile: 480px and below

## Validation

### Form Validation Rules

**Name Field:**
- Required
- Maximum 100 characters

**Bio Field:**
- Optional
- Maximum 500 characters

**Avatar URL Field:**
- Optional
- Must be valid URL format

## Accessibility

### WCAG 2.1 AA Compliance

- Semantic HTML structure
- ARIA labels and descriptions
- Keyboard navigation support
- Focus indicators
- Color contrast ratios
- Screen reader support
- Error messages linked to form fields

### Keyboard Navigation

- Tab: Navigate between form fields
- Enter: Submit form or activate buttons
- Escape: Cancel edit mode
- Shift+Tab: Navigate backwards

## Testing

### Test Coverage

- **useProfile Hook:** 85%+ coverage
  - Profile fetching
  - Profile updates
  - WebSocket connection
  - Error handling
  - Email encoding

- **ProfileCard Component:** 90%+ coverage
  - View mode rendering
  - Edit mode functionality
  - Form validation
  - Error states
  - Loading states
  - Accessibility features

- **ProfilePage Component:** 85%+ coverage
  - Authentication
  - Profile display
  - Toast notifications
  - Cleanup on unmount
  - Error handling

### Running Tests

```bash
# Run all tests
npm test

# Run tests with coverage
npm test -- --coverage

# Run specific test file
npm test -- useProfile.test.js

# Run tests in watch mode
npm test -- --watch
```

### Test Files

- `hooks/useProfile.test.js` - Hook tests
- `components/ProfileCard.test.jsx` - Component tests
- `pages/ProfilePage.test.jsx` - Page tests

## Error Handling

### Common Errors

**Authentication Error:**
- User not logged in or token expired
- Solution: Redirect to login page

**Network Error:**
- API unreachable or network failure
- Solution: Show retry button and error message

**Validation Error:**
- Form field validation failed
- Solution: Display field-specific error messages

**WebSocket Error:**
- Real-time connection failed
- Solution: Attempt automatic reconnection

## Performance Optimization

- Lazy loading of components
- Memoization of callbacks
- Efficient state updates
- WebSocket connection pooling
- Image optimization with fallback

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Dependencies

- React 16.8+ (hooks support)
- React Testing Library (for tests)
- Jest (for test runner)

## Future Enhancements

- [ ] Profile picture upload
- [ ] Social media links
- [ ] Profile visibility settings
- [ ] Activity history
- [ ] Profile sharing
- [ ] Two-factor authentication
- [ ] Profile backup/export

## Contributing

When modifying these components:

1. Maintain accessibility standards
2. Update tests for new features
3. Follow existing code style
4. Update documentation
5. Test on mobile devices
6. Verify keyboard navigation

## License

See LICENSE file in project root.
