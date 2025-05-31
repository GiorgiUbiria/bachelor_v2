import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import { useAuthStore } from './auth'
import { useCartStore } from './cart'
import { useFavoritesStore } from './favorites'
import { useOrdersStore } from './orders'
import { useProductsStore } from './products'
import { useCommentsStore } from './comments'
import { useTagsStore } from './tags'
import { useDiscountsStore } from './discounts'
import { useAppStore } from './app'
import { useUIStore } from './ui'

interface StoreManagerState {
  isInitialized: boolean
  isHydrated: boolean
  syncInProgress: boolean
  lastSyncTime: string | null
  errors: Record<string, string>
}

interface StoreManagerActions {
  initialize: () => Promise<void>
  syncAllStores: () => Promise<void>
  clearAllCaches: () => void
  resetAllStores: () => void
  handleAuthChange: (isAuthenticated: boolean) => Promise<void>
  
  // Error management
  addError: (store: string, error: string) => void
  clearError: (store: string) => void
  clearAllErrors: () => void
  
  // Hydration management
  setHydrated: (hydrated: boolean) => void
  waitForHydration: () => Promise<void>
}

type StoreManager = StoreManagerState & StoreManagerActions

export const useStoreManager = create<StoreManager>()(
  immer((set, get) => ({
    isInitialized: false,
    isHydrated: false,
    syncInProgress: false,
    lastSyncTime: null,
    errors: {},

    initialize: async () => {
      if (get().isInitialized) return

      try {
        // Wait for hydration first
        await get().waitForHydration()

        // Initialize app store first
        await useAppStore.getState().initialize()

        // Check if user is authenticated
        const { isAuthenticated } = useAuthStore.getState()
        
        if (isAuthenticated) {
          // Initialize authenticated user stores
          await get().handleAuthChange(true)
        }

        set({ isInitialized: true })
      } catch (error) {
        console.error('Failed to initialize store manager:', error)
        get().addError('store-manager', 'Failed to initialize application')
      }
    },

    syncAllStores: async () => {
      if (get().syncInProgress) return

      set({ syncInProgress: true })
      
      try {
        const { isAuthenticated } = useAuthStore.getState()
        
        if (isAuthenticated) {
          // Sync all authenticated stores in parallel
          await Promise.allSettled([
            useCartStore.getState().fetchCart(),
            useFavoritesStore.getState().fetchFavorites(),
            useOrdersStore.getState().fetchOrders(),
            useOrdersStore.getState().fetchOrderStats(),
            useCommentsStore.getState().fetchComments(''), // This might need product ID
            useTagsStore.getState().fetchTags(),
            useDiscountsStore.getState().fetchActiveDiscounts()
          ])
        }

        // Always sync products and categories
        await Promise.allSettled([
          useProductsStore.getState().fetchProducts(),
          useProductsStore.getState().fetchCategories()
        ])

        set({ lastSyncTime: new Date().toISOString() })
      } catch (error) {
        console.error('Failed to sync stores:', error)
        get().addError('sync', 'Failed to sync data')
      } finally {
        set({ syncInProgress: false })
      }
    },

    clearAllCaches: () => {
      useProductsStore.getState().clearCache()
      useAppStore.getState().clearCache()
      
      // Clear any other caches
      get().clearAllErrors()
    },

    resetAllStores: () => {
      // Reset all stores to initial state
      useCartStore.setState({ cart: null, isLoading: false, error: null })
      useFavoritesStore.setState({ favorites: [], isLoading: false, error: null })
      useOrdersStore.setState({ orders: [], currentOrder: null, isLoading: false, error: null })
      useCommentsStore.setState({ commentsByProduct: {}, isLoading: false, error: null })
      useTagsStore.setState({ tags: [], productTags: {}, isLoading: false, error: null })
      useDiscountsStore.setState({ 
        activeDiscounts: [], 
        allDiscounts: [], 
        smartSuggestions: [], 
        isLoading: false, 
        error: null 
      })
      
      get().clearAllCaches()
      get().clearAllErrors()
    },

    handleAuthChange: async (isAuthenticated: boolean) => {
      if (isAuthenticated) {
        // User logged in - fetch all user-specific data
        try {
          await Promise.allSettled([
            useCartStore.getState().fetchCart(),
            useFavoritesStore.getState().fetchFavorites(),
            useOrdersStore.getState().fetchOrders(),
            useOrdersStore.getState().fetchOrderStats()
          ])
        } catch (error) {
          console.error('Failed to fetch user data after login:', error)
          get().addError('auth', 'Failed to load user data')
        }
      } else {
        // User logged out - clear all user-specific data
        get().resetAllStores()
      }
    },

    // Error management
    addError: (store: string, error: string) => {
      set((state) => {
        state.errors[store] = error
      })
      
      // Also add to UI store for toast notifications
      useUIStore.getState().addToast({
        type: 'error',
        title: 'Error',
        description: error
      })
    },

    clearError: (store: string) => {
      set((state) => {
        delete state.errors[store]
      })
    },

    clearAllErrors: () => {
      set({ errors: {} })
    },

    // Hydration management
    setHydrated: (hydrated: boolean) => {
      set({ isHydrated: hydrated })
    },

    waitForHydration: () => {
      return new Promise<void>((resolve) => {
        const checkHydration = () => {
          if (get().isHydrated) {
            resolve()
          } else {
            setTimeout(checkHydration, 50)
          }
        }
        checkHydration()
      })
    }
  }))
)

// Auto-initialize when the module loads
if (typeof window !== 'undefined') {
  // Set up hydration detection
  const unsubscribeAuth = useAuthStore.persist.onFinishHydration(() => {
    useStoreManager.getState().setHydrated(true)
    unsubscribeAuth()
  })

  // Initialize after a short delay to ensure all stores are ready
  setTimeout(() => {
    useStoreManager.getState().initialize()
  }, 100)
}

// Export a hook for easy access to store manager
export const useStoreSync = () => {
  const { syncAllStores, syncInProgress, lastSyncTime } = useStoreManager()
  return { syncAllStores, syncInProgress, lastSyncTime }
}

export const useStoreErrors = () => {
  const { errors, clearError, clearAllErrors } = useStoreManager()
  return { errors, clearError, clearAllErrors }
} 