import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import apiService from '../services/api'

export interface OrderItem {
  id: string
  order_id: string
  product_id: string
  quantity: number
  price: number
  created_at: string
  product: {
    id: string
    name: string
    price: number
    image_url?: string
    category: string
  }
}

export interface Order {
  id: string
  user_id: string
  total: number
  status: 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled'
  created_at: string
  updated_at: string
  order_items: OrderItem[]
}

export interface OrderStats {
  total_orders: number
  total_spent: number
  pending_orders: number
  average_order_value: number
}

interface OrdersState {
  orders: Order[]
  currentOrder: Order | null
  orderStats: OrderStats | null
  isLoading: boolean
  error: string | null
}

interface OrdersActions {
  fetchOrders: (params?: { page?: number; limit?: number; status?: string }) => Promise<void>
  fetchOrderById: (id: string) => Promise<void>
  createOrder: (orderData: { cart_item_ids?: string[]; payment_method: string; shipping_address: string }) => Promise<Order>
  cancelOrder: (id: string) => Promise<void>
  updateOrderStatus: (id: string, status: string) => Promise<void>
  fetchOrderStats: () => Promise<void>
  clearError: () => void
  clearCurrentOrder: () => void
}

type OrdersStore = OrdersState & OrdersActions

export const useOrdersStore = create<OrdersStore>()(
  immer((set, get) => ({
    orders: [],
    currentOrder: null,
    orderStats: null,
    isLoading: false,
    error: null,

    fetchOrders: async (params) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.orders.getAll(params)
        set((state) => {
          state.orders = response.data.orders || []
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch orders'
          state.isLoading = false
        })
      }
    },

    fetchOrderById: async (id: string) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.orders.getById(id)
        set((state) => {
          state.currentOrder = response.data.order
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch order'
          state.isLoading = false
        })
      }
    },

    createOrder: async (orderData) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.orders.create(orderData)
        const newOrder = response.data.order
        
        set((state) => {
          state.orders.unshift(newOrder)
          state.currentOrder = newOrder
          state.isLoading = false
        })

        return newOrder
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to create order'
          state.isLoading = false
        })
        throw error
      }
    },

    cancelOrder: async (id: string) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        await apiService.orders.cancel(id)
        set((state) => {
          // Update order status in the list
          const orderIndex = state.orders.findIndex(order => order.id === id)
          if (orderIndex !== -1) {
            state.orders[orderIndex].status = 'cancelled'
          }
          
          // Update current order if it's the same
          if (state.currentOrder?.id === id) {
            state.currentOrder.status = 'cancelled'
          }
          
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to cancel order'
          state.isLoading = false
        })
        throw error
      }
    },

    updateOrderStatus: async (id: string, status: string) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        await apiService.orders.updateStatus(id, status)
        set((state) => {
          // Update order status in the list
          const orderIndex = state.orders.findIndex(order => order.id === id)
          if (orderIndex !== -1) {
            state.orders[orderIndex].status = status as any
          }
          
          // Update current order if it's the same
          if (state.currentOrder?.id === id) {
            state.currentOrder.status = status as any
          }
          
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to update order status'
          state.isLoading = false
        })
        throw error
      }
    },

    fetchOrderStats: async () => {
      try {
        const response = await apiService.orders.getStats()
        set((state) => {
          state.orderStats = response.data.stats
        })
      } catch (error: any) {
        console.error('Failed to fetch order stats:', error)
      }
    },

    clearError: () => {
      set((state) => {
        state.error = null
      })
    },

    clearCurrentOrder: () => {
      set((state) => {
        state.currentOrder = null
      })
    },
  }))
) 