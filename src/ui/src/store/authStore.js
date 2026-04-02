import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
  isAuthenticated: boolean;
  username: string;
  token: string | null;
  login: (username: string, token?: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: true,
      username: 'admin',
      token: 'demo-token',
      login: (username, token) => set({ isAuthenticated: true, username, token }),
      logout: () => set({ isAuthenticated: false, username: '', token: null }),
    }),
    { name: 'auth-storage' }
  )
);
