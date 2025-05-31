import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'
import { apiService } from '../services/api'

export interface CartItem {
  id: string
  cart_id: string
  product_id: string
  quantity: number
  created_at: string
  updated_at: string
  product: {
    id: string
    name: string
    price: number
    image_url?: string
    category: string
    stock: number
  }
}

export interface Cart {
  id: string
  user_id: string
  items: CartItem[]
  total_items: number
  total_amount: number
  created_at: string
  updated_at: string
}

interface CartState {
  cart: Cart | null
  isLoading: boolean
  isUpdating: boolean
  error: string | null
  optimisticUpdates: Record<string, { quantity: number; timestamp: number }>
}

interface CartActions {
  fetchCart: () => Promise<void>
  addItem: (productId: string, quantity?: number) => Promise<void>
  updateItem: (itemId: string, quantity: number) => Promise<void>
  removeItem: (itemId: string) => Promise<void>
  clearCart: () => Promise<void>
  
  // Optimistic updates
  optimisticAddItem: (productId: string, quantity: number) => void
  optimisticUpdateItem: (itemId: string, quantity: number) => void
  optimisticRemoveItem: (itemId: string) => void
  clearOptimisticUpdates: () => void
  
  // Utility
  getItemQuantity: (productId: string) => number
  getTotalItems: () => number
  getTotalAmount: () => number
  isInCart: (productId: string) => boolean
  clearError: () => void
}

type CartStore = CartState & CartActions

export const useCartStore = create<CartStore>()(
  persist(
    immer((set, get) => ({
      cart: null,
      isLoading: false,
      isUpdating: false,
      error: null,
      optimisticUpdates: {},

      fetchCart: async () => {
        set({ isLoading: true, error: null })
        try {
          const response = await apiService.cart.get()
          const backendCart = response.data.cart
          const transformedCart: Cart = {
            id: backendCart.id,
            user_id: backendCart.user_id,
            items: backendCart.cart_items || [],
            total_items: response.data.item_count || 0,
            total_amount: response.data.total || 0,
            created_at: backendCart.created_at,
            updated_at: backendCart.updated_at
          }
          
          set({ cart: transformedCart })
        } catch (error: any) {
          set({ error: error.response?.data?.error || 'Failed to fetch cart' })
        } finally {
          set({ isLoading: false })
        }
      },

      addItem: async (productId: string, quantity: number = 1) => {
        // Optimistic update first
        get().optimisticAddItem(productId, quantity)
        
        set({ isUpdating: true, error: null })
        try {
          await apiService.cart.add(productId, quantity)
          // Refetch cart after adding item
          const response = await apiService.cart.get()
          const backendCart = response.data.cart
          const transformedCart: Cart = {
            id: backendCart.id,
            user_id: backendCart.user_id,
            items: backendCart.cart_items || [],
            total_items: response.data.item_count || 0,
            total_amount: response.data.total || 0,
            created_at: backendCart.created_at,
            updated_at: backendCart.updated_at
          }
          
          set({ cart: transformedCart })
          get().clearOptimisticUpdates()
        } catch (error: any) {
          // Revert optimistic update on error
          await get().fetchCart()
          set({ error: error.response?.data?.error || 'Failed to add item to cart' })
          throw error
        } finally {
          set({ isUpdating: false })
        }
      },

      updateItem: async (itemId: string, quantity: number) => {
        // Optimistic update first
        get().optimisticUpdateItem(itemId, quantity)
        
        set({ isUpdating: true, error: null })
        try {
          await apiService.cart.updateItem(itemId, quantity)
          await get().fetchCart()
          get().clearOptimisticUpdates()
        } catch (error: any) {
          // Revert optimistic update on error
          await get().fetchCart()
          set({ error: error.response?.data?.error || 'Failed to update item' })
          throw error
        } finally {
          set({ isUpdating: false })
        }
      },

      removeItem: async (itemId: string) => {
        // Optimistic update first
        get().optimisticRemoveItem(itemId)
        
        set({ isUpdating: true, error: null })
        try {
          await apiService.cart.removeItem(itemId)
          await get().fetchCart()
          get().clearOptimisticUpdates()
        } catch (error: any) {
          // Revert optimistic update on error
          await get().fetchCart()
          set({ error: error.response?.data?.error || 'Failed to remove item' })
          throw error
        } finally {
          set({ isUpdating: false })
        }
      },

      clearCart: async () => {
        set({ isUpdating: true, error: null })
        try {
          await apiService.cart.clear()
          set({ cart: null })
        } catch (error: any) {
          set({ error: error.response?.data?.error || 'Failed to clear cart' })
          throw error
        } finally {
          set({ isUpdating: false })
        }
      },

      // Optimistic updates
      optimisticAddItem: (productId: string, quantity: number) => {
        set((state) => {
          if (!state.cart) return
          
          const existingItem = state.cart.items.find(item => item.product_id === productId)
          if (existingItem) {
            existingItem.quantity += quantity
          } else {
            // We can't add a new item optimistically without product data
            // So we'll just mark it for optimistic tracking
            state.optimisticUpdates[productId] = {
              quantity,
              timestamp: Date.now()
            }
          }
          
          // Recalculate totals
          state.cart.total_items = state.cart.items.reduce((sum, item) => sum + item.quantity, 0)
          state.cart.total_amount = state.cart.items.reduce((sum, item) => sum + (item.product.price * item.quantity), 0)
        })
      },

      optimisticUpdateItem: (itemId: string, quantity: number) => {
        set((state) => {
          if (!state.cart) return
          
          const item = state.cart.items.find(item => item.id === itemId)
          if (item) {
            item.quantity = quantity
            
            // Recalculate totals
            state.cart.total_items = state.cart.items.reduce((sum, item) => sum + item.quantity, 0)
            state.cart.total_amount = state.cart.items.reduce((sum, item) => sum + (item.product.price * item.quantity), 0)
          }
        })
      },

      optimisticRemoveItem: (itemId: string) => {
        set((state) => {
          if (!state.cart) return
          
          state.cart.items = state.cart.items.filter(item => item.id !== itemId)
          
          // Recalculate totals
          state.cart.total_items = state.cart.items.reduce((sum, item) => sum + item.quantity, 0)
          state.cart.total_amount = state.cart.items.reduce((sum, item) => sum + (item.product.price * item.quantity), 0)
        })
      },

      clearOptimisticUpdates: () => {
        set({ optimisticUpdates: {} })
      },

      // Utility functions
      getItemQuantity: (productId: string) => {
        const cart = get().cart
        if (!cart) return 0
        
        const item = cart.items.find(item => item.product_id === productId)
        return item ? item.quantity : 0
      },

      getTotalItems: () => {
        const cart = get().cart
        return cart ? cart.total_items : 0
      },

      getTotalAmount: () => {
        const cart = get().cart
        return cart ? cart.total_amount : 0
      },

      isInCart: (productId: string) => {
        const cart = get().cart
        if (!cart) return false
        
        return cart.items.some(item => item.product_id === productId)
      },

      clearError: () => {
        set({ error: null })
      }
    })),
    {
      name: 'cart-store',
      partialize: (state) => ({
        cart: state.cart
      })
    }
  )
) 