'use client';

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import { user as userApi, auth as authApi } from '@/lib/api';
import type { User } from '@/lib/types';

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  login: (email: string, password: string) => Promise<{ error?: string }>;
  signup: (name: string, email: string, password: string) => Promise<{ error?: string }>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const res = await userApi.me();
      if (res.success && res.data) {
        setCurrentUser(res.data);
      } else {
        setCurrentUser(null);
      }
    } catch {
      setCurrentUser(null);
    }
  }, []);

  useEffect(() => {
    refresh().finally(() => setIsLoading(false));
  }, [refresh]);

  const login = useCallback(async (email: string, password: string) => {
    const res = await authApi.login(email, password);
    if (res.success && res.data) {
      setCurrentUser(res.data.user);
      return {};
    }
    return { error: res.error || 'Login failed' };
  }, []);

  const signup = useCallback(async (name: string, email: string, password: string) => {
    const res = await authApi.signup(name, email, password);
    if (res.success && res.data) {
      setCurrentUser(res.data.user);
      return {};
    }
    return { error: res.error || 'Signup failed' };
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout();
    setCurrentUser(null);
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user: currentUser,
        isLoading,
        isAuthenticated: !!currentUser,
        isAdmin: currentUser?.role === 'admin',
        login,
        signup,
        logout,
        refresh,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
