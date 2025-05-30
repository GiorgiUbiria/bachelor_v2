import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

export type Theme = 'light' | 'dark' | 'system'

export interface Toast {
  id: string
  title?: string
  description: string
  type: 'success' | 'error' | 'warning' | 'info'
  duration?: number
}

interface UIState {
  theme: Theme
  toasts: Toast[]
  isLoading: boolean
  loadingMessage?: string
  sidebarOpen: boolean
}

interface UIActions {
  setTheme: (theme: Theme) => void
  addToast: (toast: Omit<Toast, 'id'>) => void
  removeToast: (id: string) => void
  clearToasts: () => void
  setLoading: (loading: boolean, message?: string) => void
  setSidebarOpen: (open: boolean) => void
}

type UIStore = UIState & UIActions

export const useUIStore = create<UIStore>()(
  persist(
    immer((set) => ({
      // Initial state
      theme: 'system',
      toasts: [],
      isLoading: false,
      loadingMessage: undefined,
      sidebarOpen: false,

      // Actions
      setTheme: (theme: Theme) => {
        set((state) => {
          state.theme = theme
        })
        
        // Apply theme to document
        const root = document.documentElement
        if (theme === 'dark') {
          root.classList.add('dark')
        } else if (theme === 'light') {
          root.classList.remove('dark')
        } else {
          // System theme
          const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
          if (prefersDark) {
            root.classList.add('dark')
          } else {
            root.classList.remove('dark')
          }
        }
      },

      addToast: (toast: Omit<Toast, 'id'>) => {
        const id = Math.random().toString(36).substr(2, 9)
        set((state) => {
          state.toasts.push({ ...toast, id })
        })

        // Auto remove toast after duration
        const duration = toast.duration || 5000
        setTimeout(() => {
          set((state) => {
            state.toasts = state.toasts.filter(t => t.id !== id)
          })
        }, duration)
      },

      removeToast: (id: string) => {
        set((state) => {
          state.toasts = state.toasts.filter(t => t.id !== id)
        })
      },

      clearToasts: () => {
        set((state) => {
          state.toasts = []
        })
      },

      setLoading: (loading: boolean, message?: string) => {
        set((state) => {
          state.isLoading = loading
          state.loadingMessage = message
        })
      },

      setSidebarOpen: (open: boolean) => {
        set((state) => {
          state.sidebarOpen = open
        })
      },
    })),
    {
      name: 'ui-storage',
      partialize: (state) => ({
        theme: state.theme,
        sidebarOpen: state.sidebarOpen,
      }),
    }
  )
) 