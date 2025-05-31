import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import apiService from '../services/api'

export interface Discount {
  id: string
  product_id?: string
  category?: string
  discount_type: string
  discount_value: number
  start_date: string
  end_date: string
  min_order_amount?: number
  max_discount_amount?: number
  usage_limit?: number
  usage_count: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface SmartDiscountSuggestion {
  discount_type: string
  discount_value: number
  reason: string
  confidence: number
  product_id?: string
  category?: string
}

interface DiscountsState {
  activeDiscounts: Discount[]
  allDiscounts: Discount[]
  smartSuggestions: SmartDiscountSuggestion[]
  isLoading: boolean
  error: string | null
  isSubmitting: boolean
}

interface DiscountsActions {
  fetchActiveDiscounts: () => Promise<void>
  fetchAllDiscounts: () => Promise<void>
  fetchSmartSuggestions: (productId?: string, category?: string) => Promise<void>
  createDiscount: (discountData: {
    product_id?: string;
    category?: string;
    discount_type: string;
    discount_value: number;
    start_date: string;
    end_date: string;
    min_order_amount?: number;
    max_discount_amount?: number;
    usage_limit?: number;
  }) => Promise<void>
  applyDiscount: (discountCode: string, orderTotal: number) => Promise<{ discountAmount: number; finalTotal: number }>
  calculateDiscount: (discount: Discount, orderTotal: number) => number
  clearError: () => void
  setLoading: (loading: boolean) => void
}

type DiscountsStore = DiscountsState & DiscountsActions

export const useDiscountsStore = create<DiscountsStore>()(
  immer((set, get) => ({
    activeDiscounts: [],
    allDiscounts: [],
    smartSuggestions: [],
    isLoading: false,
    error: null,
    isSubmitting: false,

    fetchActiveDiscounts: async () => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.discounts.getActive()
        set((state) => {
          state.activeDiscounts = response.data.discounts || []
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch active discounts'
          state.isLoading = false
        })
        throw error
      }
    },

    fetchAllDiscounts: async () => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        // Note: We need to add a getAll endpoint to the API service
        // For now, we'll use the active discounts endpoint
        const response = await apiService.discounts.getActive()
        set((state) => {
          state.allDiscounts = response.data.discounts || []
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch discounts'
          state.isLoading = false
        })
        throw error
      }
    },

    fetchSmartSuggestions: async (productId?: string, category?: string) => {
      try {
        let suggestions: SmartDiscountSuggestion[] = []

        if (productId) {
          const response = await apiService.mlService.smartDiscounts.suggestProduct(productId)
          if (response.data.discount) {
            suggestions.push({
              ...response.data.discount,
              product_id: productId
            })
          }
        }

        if (category) {
          const response = await apiService.mlService.smartDiscounts.suggestCategory(category)
          if (response.data.discount) {
            suggestions.push({
              ...response.data.discount,
              category: category
            })
          }
        }

        set((state) => {
          state.smartSuggestions = suggestions
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch smart discount suggestions'
        })
        throw error
      }
    },

    createDiscount: async (discountData) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        const response = await apiService.discounts.create(discountData)
        const newDiscount = response.data.discount

        set((state) => {
          state.allDiscounts.push(newDiscount)
          if (newDiscount.is_active) {
            state.activeDiscounts.push(newDiscount)
          }
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to create discount'
          state.isSubmitting = false
        })
        throw error
      }
    },

    applyDiscount: async (discountCode: string, orderTotal: number) => {
      // This would typically call an API endpoint to apply a discount code
      // For now, we'll simulate the logic
      const { activeDiscounts } = get()
      
      // Find discount by code (assuming discount code is the ID for now)
      const discount = activeDiscounts.find(d => d.id === discountCode)
      
      if (!discount) {
        throw new Error('Invalid discount code')
      }

      if (!discount.is_active) {
        throw new Error('Discount is not active')
      }

      const now = new Date()
      const startDate = new Date(discount.start_date)
      const endDate = new Date(discount.end_date)

      if (now < startDate || now > endDate) {
        throw new Error('Discount is not valid at this time')
      }

      if (discount.min_order_amount && orderTotal < discount.min_order_amount) {
        throw new Error(`Minimum order amount of $${discount.min_order_amount} required`)
      }

      if (discount.usage_limit && discount.usage_count >= discount.usage_limit) {
        throw new Error('Discount usage limit reached')
      }

      const discountAmount = get().calculateDiscount(discount, orderTotal)
      const finalTotal = Math.max(0, orderTotal - discountAmount)

      return { discountAmount, finalTotal }
    },

    calculateDiscount: (discount: Discount, orderTotal: number) => {
      let discountAmount = 0

      switch (discount.discount_type) {
        case 'percentage':
          discountAmount = (orderTotal * discount.discount_value) / 100
          break
        case 'fixed':
          discountAmount = discount.discount_value
          break
        case 'buy_one_get_one':
          // Simplified BOGO logic
          discountAmount = orderTotal * 0.5
          break
        default:
          discountAmount = 0
      }

      // Apply maximum discount limit if specified
      if (discount.max_discount_amount) {
        discountAmount = Math.min(discountAmount, discount.max_discount_amount)
      }

      return Math.min(discountAmount, orderTotal)
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
  }))
) 