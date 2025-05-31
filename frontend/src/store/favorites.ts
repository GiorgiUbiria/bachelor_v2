import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import { persist } from 'zustand/middleware'
import apiService from '../services/api'

export interface Favorite {
  id: string
  user_id: string
  product_id: string
  created_at: string
  product: {
    id: string
    name: string
    price: number
    image_url?: string
    category: string
    description: string
  }
}

interface FavoritesState {
  favorites: Favorite[]
  isLoading: boolean
  isUpdating: boolean
  error: string | null
  optimisticUpdates: Record<string, { action: 'add' | 'remove'; timestamp: number }>
}

interface FavoritesActions {
  fetchFavorites: () => Promise<void>
  addFavorite: (productId: string) => Promise<void>
  removeFavorite: (productId: string) => Promise<void>
  toggleFavorite: (productId: string) => Promise<void>
  
  // Optimistic updates
  optimisticAddFavorite: (productId: string) => void
  optimisticRemoveFavorite: (productId: string) => void
  clearOptimisticUpdates: () => void
  
  // Utility
  isFavorite: (productId: string) => boolean
  getFavoriteCount: () => number
  getFavoritesByCategory: (category: string) => Favorite[]
  clearError: () => void
}

type FavoritesStore = FavoritesState & FavoritesActions

export const useFavoritesStore = create<FavoritesStore>()(
  persist(
    immer((set, get) => ({
      favorites: [],
      isLoading: false,
      isUpdating: false,
      error: null,
      optimisticUpdates: {},

      fetchFavorites: async () => {
        set({ isLoading: true, error: null })
        try {
          const response = await apiService.favorites.getAll()
          set({ favorites: response.data.favorites || [] })
        } catch (error: any) {
          set({ error: error.response?.data?.error || 'Failed to fetch favorites' })
        } finally {
          set({ isLoading: false })
        }
      },

      addFavorite: async (productId: string) => {
        // Optimistic update first
        get().optimisticAddFavorite(productId)
        
        set({ isUpdating: true, error: null })
        try {
          const response = await apiService.favorites.add(productId)
          set((state) => {
            // Remove from optimistic updates and add real data
            delete state.optimisticUpdates[productId]
            
            // Check if already exists to avoid duplicates
            const exists = state.favorites.some(fav => fav.product_id === productId)
            if (!exists) {
              state.favorites.push(response.data.favorite)
            }
          })
        } catch (error: any) {
          // Revert optimistic update on error
          set((state) => {
            delete state.optimisticUpdates[productId]
            state.favorites = state.favorites.filter(fav => fav.product_id !== productId)
          })
          set({ error: error.response?.data?.error || 'Failed to add favorite' })
          throw error
        } finally {
          set({ isUpdating: false })
        }
      },

      removeFavorite: async (productId: string) => {
        // Optimistic update first
        get().optimisticRemoveFavorite(productId)
        
        set({ isUpdating: true, error: null })
        try {
          await apiService.favorites.remove(productId)
          set((state) => {
            delete state.optimisticUpdates[productId]
            state.favorites = state.favorites.filter(fav => fav.product_id !== productId)
          })
        } catch (error: any) {
          // Revert optimistic update on error - refetch to restore state
          await get().fetchFavorites()
          set({ error: error.response?.data?.error || 'Failed to remove favorite' })
          throw error
        } finally {
          set({ isUpdating: false })
        }
      },

      toggleFavorite: async (productId: string) => {
        const isFav = get().isFavorite(productId)
        if (isFav) {
          await get().removeFavorite(productId)
        } else {
          await get().addFavorite(productId)
        }
      },

      // Optimistic updates
      optimisticAddFavorite: (productId: string) => {
        set((state) => {
          state.optimisticUpdates[productId] = {
            action: 'add',
            timestamp: Date.now()
          }
          
          // Don't add to favorites array yet since we don't have product data
          // The UI should check both favorites array and optimistic updates
        })
      },

      optimisticRemoveFavorite: (productId: string) => {
        set((state) => {
          state.optimisticUpdates[productId] = {
            action: 'remove',
            timestamp: Date.now()
          }
        })
      },

      clearOptimisticUpdates: () => {
        set({ optimisticUpdates: {} })
      },

      // Utility functions
      isFavorite: (productId: string) => {
        const state = get()
        
        // Check optimistic updates first
        const optimistic = state.optimisticUpdates[productId]
        if (optimistic) {
          return optimistic.action === 'add'
        }
        
        // Check actual favorites
        return state.favorites.some(fav => fav.product_id === productId)
      },

      getFavoriteCount: () => {
        const state = get()
        let count = state.favorites.length
        
        // Adjust for optimistic updates
        Object.entries(state.optimisticUpdates).forEach(([productId, update]) => {
          const existsInFavorites = state.favorites.some(fav => fav.product_id === productId)
          
          if (update.action === 'add' && !existsInFavorites) {
            count++
          } else if (update.action === 'remove' && existsInFavorites) {
            count--
          }
        })
        
        return Math.max(0, count)
      },

      getFavoritesByCategory: (category: string) => {
        return get().favorites.filter(fav => fav.product.category === category)
      },

      clearError: () => {
        set({ error: null })
      }
    })),
    {
      name: 'favorites-store',
      partialize: (state) => ({
        favorites: state.favorites
      })
    }
  )
) 