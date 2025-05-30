import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { apiService } from '../services/api'

export interface User {
  id: string
  email: string
  name: string
  created_at: string
  updated_at: string
}

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null
}

interface AuthActions {
  login: (email: string, password: string) => Promise<void>
  register: (email: string, name: string, password: string) => Promise<void>
  logout: () => void
  clearError: () => void
  setLoading: (loading: boolean) => void
  checkAuth: () => Promise<void>
}

type AuthStore = AuthState & AuthActions

export const useAuthStore = create<AuthStore>()(
  persist(
    immer((set, get) => ({
      // Initial state
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (email: string, password: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const response = await apiService.auth.login({ email, password })
          const data = response.data

          set((state) => {
            state.user = data.user
            state.token = data.token
            state.isAuthenticated = true
            state.isLoading = false
            state.error = null
          })
        } catch (error: any) {
          const errorMessage = error.response?.data?.error || error.message || 'Login failed'
          set((state) => {
            state.error = errorMessage
            state.isLoading = false
          })
          throw new Error(errorMessage)
        }
      },

      register: async (email: string, name: string, password: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const response = await apiService.auth.register({ email, name, password })
          const data = response.data

          set((state) => {
            state.user = data.user
            state.token = data.token
            state.isAuthenticated = true
            state.isLoading = false
            state.error = null
          })
        } catch (error: any) {
          const errorMessage = error.response?.data?.error || error.message || 'Registration failed'
          set((state) => {
            state.error = errorMessage
            state.isLoading = false
          })
          throw new Error(errorMessage)
        }
      },

      logout: async () => {
        try {
          // Call logout endpoint if authenticated
          if (get().isAuthenticated) {
            await apiService.auth.logout()
          }
        } catch (error) {
          // Ignore logout errors, still clear local state
          console.warn('Logout request failed:', error)
        } finally {
          set((state) => {
            state.user = null
            state.token = null
            state.isAuthenticated = false
            state.error = null
          })
        }
      },

      checkAuth: async () => {
        const { token } = get()
        if (!token) return

        try {
          set((state) => {
            state.isLoading = true
          })

          const response = await apiService.auth.me()
          const user = response.data

          set((state) => {
            state.user = user
            state.isAuthenticated = true
            state.isLoading = false
          })
        } catch (error) {
          // Token is invalid, clear auth state
          set((state) => {
            state.user = null
            state.token = null
            state.isAuthenticated = false
            state.isLoading = false
          })
        }
      },

      clearError: () => {
        set((state) => {
          state.error = null
        })
      },

      setLoading: (loading: boolean) => {
        set((state) => {
          state.isLoading = loading
        })
      },
    })),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
) 