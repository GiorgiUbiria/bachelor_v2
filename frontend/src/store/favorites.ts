import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
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
  error: string | null
}

interface FavoritesActions {
  fetchFavorites: () => Promise<void>
  addFavorite: (productId: string) => Promise<void>
  removeFavorite: (productId: string) => Promise<void>
  isFavorite: (productId: string) => boolean
  clearError: () => void
}

type FavoritesStore = FavoritesState & FavoritesActions

export const useFavoritesStore = create<FavoritesStore>()(
  immer((set, get) => ({
    favorites: [],
    isLoading: false,
    error: null,

    fetchFavorites: async () => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.favorites.getAll()
        set((state) => {
          state.favorites = response.data.favorites || []
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch favorites'
          state.isLoading = false
        })
      }
    },

    addFavorite: async (productId: string) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.favorites.add(productId)
        set((state) => {
          state.favorites.push(response.data.favorite)
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to add favorite'
          state.isLoading = false
        })
        throw error
      }
    },

    removeFavorite: async (productId: string) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        await apiService.favorites.remove(productId)
        set((state) => {
          state.favorites = state.favorites.filter(fav => fav.product_id !== productId)
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to remove favorite'
          state.isLoading = false
        })
        throw error
      }
    },

    isFavorite: (productId: string) => {
      return get().favorites.some(fav => fav.product_id === productId)
    },

    clearError: () => {
      set((state) => {
        state.error = null
      })
    },
  }))
) 