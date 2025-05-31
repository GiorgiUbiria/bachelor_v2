import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import { persist } from 'zustand/middleware'

export interface UserPreferences {
  // Display preferences
  theme: 'light' | 'dark' | 'system'
  language: string
  currency: string
  timezone: string
  
  // Shopping preferences
  defaultView: 'grid' | 'list'
  itemsPerPage: number
  sortBy: 'name' | 'price' | 'rating' | 'created_at'
  sortOrder: 'asc' | 'desc'
  
  // Notification preferences
  emailNotifications: {
    orderUpdates: boolean
    promotions: boolean
    recommendations: boolean
    newsletter: boolean
  }
  pushNotifications: {
    orderUpdates: boolean
    promotions: boolean
    recommendations: boolean
  }
  
  // Privacy preferences
  allowAnalytics: boolean
  allowPersonalization: boolean
  allowLocationTracking: boolean
  
  // Accessibility preferences
  reducedMotion: boolean
  highContrast: boolean
  fontSize: 'small' | 'medium' | 'large'
  
  // Shopping behavior
  autoSaveCart: boolean
  showRecommendations: boolean
  rememberFilters: boolean
}

interface SettingsState {
  preferences: UserPreferences
  isLoading: boolean
  error: string | null
  hasUnsavedChanges: boolean
}

interface SettingsActions {
  // Preference management
  updatePreferences: (preferences: Partial<UserPreferences>) => void
  resetPreferences: () => void
  savePreferences: () => Promise<void>
  loadPreferences: () => Promise<void>
  
  // Theme management
  setTheme: (theme: UserPreferences['theme']) => void
  toggleTheme: () => void
  
  // Display preferences
  setDefaultView: (view: UserPreferences['defaultView']) => void
  setItemsPerPage: (count: number) => void
  setSorting: (sortBy: UserPreferences['sortBy'], sortOrder: UserPreferences['sortOrder']) => void
  
  // Notification preferences
  updateEmailNotifications: (notifications: Partial<UserPreferences['emailNotifications']>) => void
  updatePushNotifications: (notifications: Partial<UserPreferences['pushNotifications']>) => void
  
  // Privacy preferences
  updatePrivacySettings: (settings: Partial<Pick<UserPreferences, 'allowAnalytics' | 'allowPersonalization' | 'allowLocationTracking'>>) => void
  
  // Accessibility
  updateAccessibilitySettings: (settings: Partial<Pick<UserPreferences, 'reducedMotion' | 'highContrast' | 'fontSize'>>) => void
  
  // Utility
  clearError: () => void
  markSaved: () => void
}

type SettingsStore = SettingsState & SettingsActions

const defaultPreferences: UserPreferences = {
  // Display preferences
  theme: 'system',
  language: 'en',
  currency: 'USD',
  timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  
  // Shopping preferences
  defaultView: 'grid',
  itemsPerPage: 12,
  sortBy: 'created_at',
  sortOrder: 'desc',
  
  // Notification preferences
  emailNotifications: {
    orderUpdates: true,
    promotions: false,
    recommendations: true,
    newsletter: false
  },
  pushNotifications: {
    orderUpdates: true,
    promotions: false,
    recommendations: false
  },
  
  // Privacy preferences
  allowAnalytics: true,
  allowPersonalization: true,
  allowLocationTracking: false,
  
  // Accessibility preferences
  reducedMotion: false,
  highContrast: false,
  fontSize: 'medium',
  
  // Shopping behavior
  autoSaveCart: true,
  showRecommendations: true,
  rememberFilters: true
}

export const useSettingsStore = create<SettingsStore>()(
  persist(
    immer((set, get) => ({
      preferences: defaultPreferences,
      isLoading: false,
      error: null,
      hasUnsavedChanges: false,

      // Preference management
      updatePreferences: (newPreferences: Partial<UserPreferences>) => {
        set((state) => {
          state.preferences = { ...state.preferences, ...newPreferences }
          state.hasUnsavedChanges = true
        })
      },

      resetPreferences: () => {
        set((state) => {
          state.preferences = { ...defaultPreferences }
          state.hasUnsavedChanges = true
        })
      },

      savePreferences: async () => {
        set({ isLoading: true, error: null })
        
        try {
          // Here you would typically save to a backend API
          // For now, we'll just simulate the save
          await new Promise(resolve => setTimeout(resolve, 500))
          
          set({ hasUnsavedChanges: false })
        } catch (error: any) {
          set({ error: 'Failed to save preferences' })
          throw error
        } finally {
          set({ isLoading: false })
        }
      },

      loadPreferences: async () => {
        set({ isLoading: true, error: null })
        
        try {
          // Here you would typically load from a backend API
          // For now, we'll use the persisted data
          await new Promise(resolve => setTimeout(resolve, 200))
        } catch (error: any) {
          set({ error: 'Failed to load preferences' })
        } finally {
          set({ isLoading: false })
        }
      },

      // Theme management
      setTheme: (theme: UserPreferences['theme']) => {
        get().updatePreferences({ theme })
        
        // Apply theme immediately
        if (theme === 'dark') {
          document.documentElement.classList.add('dark')
        } else if (theme === 'light') {
          document.documentElement.classList.remove('dark')
        } else {
          // System theme
          const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
          if (prefersDark) {
            document.documentElement.classList.add('dark')
          } else {
            document.documentElement.classList.remove('dark')
          }
        }
      },

      toggleTheme: () => {
        const currentTheme = get().preferences.theme
        const newTheme = currentTheme === 'light' ? 'dark' : 'light'
        get().setTheme(newTheme)
      },

      // Display preferences
      setDefaultView: (view: UserPreferences['defaultView']) => {
        get().updatePreferences({ defaultView: view })
      },

      setItemsPerPage: (count: number) => {
        get().updatePreferences({ itemsPerPage: count })
      },

      setSorting: (sortBy: UserPreferences['sortBy'], sortOrder: UserPreferences['sortOrder']) => {
        get().updatePreferences({ sortBy, sortOrder })
      },

      // Notification preferences
      updateEmailNotifications: (notifications: Partial<UserPreferences['emailNotifications']>) => {
        const currentNotifications = get().preferences.emailNotifications
        get().updatePreferences({
          emailNotifications: { ...currentNotifications, ...notifications }
        })
      },

      updatePushNotifications: (notifications: Partial<UserPreferences['pushNotifications']>) => {
        const currentNotifications = get().preferences.pushNotifications
        get().updatePreferences({
          pushNotifications: { ...currentNotifications, ...notifications }
        })
      },

      // Privacy preferences
      updatePrivacySettings: (settings: Partial<Pick<UserPreferences, 'allowAnalytics' | 'allowPersonalization' | 'allowLocationTracking'>>) => {
        get().updatePreferences(settings)
      },

      // Accessibility
      updateAccessibilitySettings: (settings: Partial<Pick<UserPreferences, 'reducedMotion' | 'highContrast' | 'fontSize'>>) => {
        get().updatePreferences(settings)
        
        // Apply accessibility settings immediately
        if (settings.reducedMotion !== undefined) {
          if (settings.reducedMotion) {
            document.documentElement.style.setProperty('--animation-duration', '0s')
          } else {
            document.documentElement.style.removeProperty('--animation-duration')
          }
        }
        
        if (settings.highContrast !== undefined) {
          if (settings.highContrast) {
            document.documentElement.classList.add('high-contrast')
          } else {
            document.documentElement.classList.remove('high-contrast')
          }
        }
        
        if (settings.fontSize !== undefined) {
          document.documentElement.setAttribute('data-font-size', settings.fontSize)
        }
      },

      // Utility
      clearError: () => {
        set({ error: null })
      },

      markSaved: () => {
        set({ hasUnsavedChanges: false })
      }
    })),
    {
      name: 'settings-store',
      partialize: (state) => ({
        preferences: state.preferences
      })
    }
  )
)

// Initialize theme on load
if (typeof window !== 'undefined') {
  const { preferences } = useSettingsStore.getState()
  useSettingsStore.getState().setTheme(preferences.theme)
} 