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
    image_url: string
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
  error: string | null
}

interface CartActions {
  fetchCart: () => Promise<void>
  addItem: (productId: string, quantity: number) => Promise<void>
  updateItem: (itemId: string, quantity: number) => Promise<void>
  removeItem: (itemId: string) => Promise<void>
  clearCart: () => Promise<void>
  clearError: () => void
}

export const useCartStore = create<CartState & CartActions>()(
  persist(
    immer((set) => ({
      cart: null,
      isLoading: false,
      error: null,

      fetchCart: async () => {
        try {
          set((state) => {
            state.isLoading = true
            state.error = null
          })

          const response = await apiService.cart.get()
          
          // Transform backend response to match frontend interface
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
          
          set((state) => {
            state.cart = transformedCart
            state.isLoading = false
          })
        } catch (error) {
          set((state) => {
            state.error = 'Failed to fetch cart'
            state.isLoading = false
          })
        }
      },

      addItem: async (productId: string, quantity: number) => {
        try {
          set((state) => {
            state.error = null
          })

          await apiService.cart.addItem(productId, quantity)
          
          // Refetch cart to get updated data
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
          
          set((state) => {
            state.cart = transformedCart
          })
        } catch (error) {
          set((state) => {
            state.error = 'Failed to add item to cart'
          })
          throw error
        }
      },

      updateItem: async (itemId: string, quantity: number) => {
        try {
          set((state) => {
            state.error = null
          })

          await apiService.cart.updateItem(itemId, quantity)
          
          // Refetch cart to get updated data
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
          
          set((state) => {
            state.cart = transformedCart
          })
        } catch (error) {
          set((state) => {
            state.error = 'Failed to update cart item'
          })
          throw error
        }
      },

      removeItem: async (itemId: string) => {
        try {
          set((state) => {
            state.error = null
          })

          await apiService.cart.removeItem(itemId)
          
          // Refetch cart to get updated data
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
          
          set((state) => {
            state.cart = transformedCart
          })
        } catch (error) {
          set((state) => {
            state.error = 'Failed to remove item from cart'
          })
          throw error
        }
      },

      clearCart: async () => {
        try {
          set((state) => {
            state.error = null
          })

          await apiService.cart.clear()
          
          set((state) => {
            state.cart = null
          })
        } catch (error) {
          set((state) => {
            state.error = 'Failed to clear cart'
          })
          throw error
        }
      },

      clearError: () => {
        set((state) => {
          state.error = null
        })
      },
    })),
    {
      name: 'cart-storage',
      partialize: (state) => ({ cart: state.cart }),
    }
  )
) 