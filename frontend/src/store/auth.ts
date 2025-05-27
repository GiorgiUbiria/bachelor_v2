import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

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
}

type AuthStore = AuthState & AuthActions

export const useAuthStore = create<AuthStore>()(
  persist(
    immer((set) => ({
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
          const response = await fetch('http://localhost:8080/api/v1/auth/login', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
          })

          if (!response.ok) {
            const errorData = await response.json()
            throw new Error(errorData.error || 'Login failed')
          }

          const data = await response.json()

          set((state) => {
            state.user = data.user
            state.token = data.token
            state.isAuthenticated = true
            state.isLoading = false
            state.error = null
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Login failed'
            state.isLoading = false
          })
          throw error
        }
      },

      register: async (email: string, name: string, password: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          const response = await fetch('http://localhost:8080/api/v1/auth/register', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, name, password }),
          })

          if (!response.ok) {
            const errorData = await response.json()
            throw new Error(errorData.error || 'Registration failed')
          }

          const data = await response.json()

          set((state) => {
            state.user = data.user
            state.token = data.token
            state.isAuthenticated = true
            state.isLoading = false
            state.error = null
          })
        } catch (error) {
          set((state) => {
            state.error = error instanceof Error ? error.message : 'Registration failed'
            state.isLoading = false
          })
          throw error
        }
      },

      logout: () => {
        set((state) => {
          state.user = null
          state.token = null
          state.isAuthenticated = false
          state.error = null
        })
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