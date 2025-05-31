import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import { persist } from 'zustand/middleware'

interface AppSettings {
  theme: 'light' | 'dark' | 'system'
  language: string
  currency: string
  notifications: {
    email: boolean
    push: boolean
    marketing: boolean
  }
  preferences: {
    defaultView: 'grid' | 'list'
    itemsPerPage: number
    autoRefresh: boolean
  }
}

interface AppState {
  isInitialized: boolean
  isOnline: boolean
  lastSync: string | null
  settings: AppSettings
  cache: {
    products: Record<string, any>
    categories: string[]
    lastUpdated: Record<string, string>
  }
}

interface AppActions {
  initialize: () => Promise<void>
  setOnlineStatus: (isOnline: boolean) => void
  updateSettings: (settings: Partial<AppSettings>) => void
  clearCache: () => void
  updateCache: (key: string, data: any) => void
  getFromCache: (key: string) => any
  syncData: () => Promise<void>
}

type AppStore = AppState & AppActions

const defaultSettings: AppSettings = {
  theme: 'system',
  language: 'en',
  currency: 'USD',
  notifications: {
    email: true,
    push: true,
    marketing: false
  },
  preferences: {
    defaultView: 'grid',
    itemsPerPage: 12,
    autoRefresh: true
  }
}

export const useAppStore = create<AppStore>()(
  persist(
    immer((set, get) => ({
      isInitialized: false,
      isOnline: navigator.onLine,
      lastSync: null,
      settings: defaultSettings,
      cache: {
        products: {},
        categories: [],
        lastUpdated: {}
      },

      initialize: async () => {
        set((state) => {
          state.isInitialized = true
          state.isOnline = navigator.onLine
        })

        // Set up online/offline listeners
        const handleOnline = () => get().setOnlineStatus(true)
        const handleOffline = () => get().setOnlineStatus(false)
        
        window.addEventListener('online', handleOnline)
        window.addEventListener('offline', handleOffline)

        // Auto-sync if online
        if (navigator.onLine) {
          await get().syncData()
        }
      },

      setOnlineStatus: (isOnline: boolean) => {
        set((state) => {
          state.isOnline = isOnline
        })

        // Auto-sync when coming back online
        if (isOnline) {
          get().syncData()
        }
      },

      updateSettings: (newSettings: Partial<AppSettings>) => {
        set((state) => {
          state.settings = { ...state.settings, ...newSettings }
        })
      },

      clearCache: () => {
        set((state) => {
          state.cache = {
            products: {},
            categories: [],
            lastUpdated: {}
          }
        })
      },

      updateCache: (key: string, data: any) => {
        set((state) => {
          state.cache.products[key] = data
          state.cache.lastUpdated[key] = new Date().toISOString()
        })
      },

      getFromCache: (key: string) => {
        const cache = get().cache
        const data = cache.products[key]
        const lastUpdated = cache.lastUpdated[key]
        
        // Check if cache is still valid (5 minutes)
        if (data && lastUpdated) {
          const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000)
          if (new Date(lastUpdated) > fiveMinutesAgo) {
            return data
          }
        }
        
        return null
      },

      syncData: async () => {
        try {
          set((state) => {
            state.lastSync = new Date().toISOString()
          })
          
          // This would trigger syncing across all stores
          // For now, we'll just update the timestamp
        } catch (error) {
          console.error('Failed to sync data:', error)
        }
      }
    })),
    {
      name: 'app-store',
      partialize: (state) => ({
        settings: state.settings,
        cache: state.cache,
        lastSync: state.lastSync
      })
    }
  )
) 