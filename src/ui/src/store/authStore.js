import { create } from 'zustand';
import { persist } from 'zustand/middleware';

const API_URL = 'http://localhost:8050';

interface AuthState {
  isAuthenticated: boolean;
  username: string;
  token: string | null;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      username: '',
      token: null,
      login: async (username, password) => {
        try {
          const response = await fetch(`${API_URL}/api/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
          });
          
          if (response.ok) {
            const data = await response.json();
            set({ isAuthenticated: true, username, token: data.token });
            return true;
          }
          return false;
        } catch (error) {
          console.error('Login failed:', error);
          return false;
        }
      },
      logout: () => set({ isAuthenticated: false, username: '', token: null }),
    }),
    { name: 'auth-storage' }
  )
);
