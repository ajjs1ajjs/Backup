import { create } from 'zustand';
import { persist } from 'zustand/middleware';

const API_URL =
  window.localStorage.getItem('apiUrl') ||
  process.env.REACT_APP_API_URL ||
  'http://localhost:8000';

export const useAuthStore = create(
  persist(
    (set) => ({
      isAuthenticated: false,
      username: '',
      token: null,
      authError: '',
      login: async (username, password) => {
        try {
          const response = await fetch(`${API_URL}/api/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
          });
          
          if (response.ok) {
            const data = await response.json();
            set({ isAuthenticated: true, username, token: data.token, authError: '' });
            return true;
          }

          const errorData = await response.json().catch(() => ({}));
          if (errorData?.code === 'PASSWORD_CHANGE_REQUIRED') {
            set({ authError: 'PASSWORD_CHANGE_REQUIRED' });
          } else {
            set({ authError: 'INVALID_CREDENTIALS' });
          }
          return false;
        } catch (error) {
          console.error('Login failed:', error);
          set({ authError: 'NETWORK_ERROR' });
          return false;
        }
      },
      changePasswordFirstLogin: async (username, currentPassword, newPassword) => {
        const response = await fetch(`${API_URL}/api/auth/change-password-first-login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username, currentPassword, newPassword })
        });

        if (!response.ok) {
          const payload = await response.json().catch(() => ({}));
          throw new Error(payload?.error || 'Failed to change password');
        }

        const data = await response.json();
        set({ isAuthenticated: true, username, token: data.token, authError: '' });
        return true;
      },
      logout: () => set({ isAuthenticated: false, username: '', token: null, authError: '' }),
    }),
    { name: 'auth-storage' }
  )
);
