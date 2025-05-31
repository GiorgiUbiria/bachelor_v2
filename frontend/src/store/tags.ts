import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import apiService from '../services/api'

export interface Tag {
  id: string
  name: string
  description?: string
  color?: string
  created_at: string
  updated_at: string
}

export interface ProductTag {
  id: string
  product_id: string
  tag_id: string
  tag: Tag
  created_at: string
}

interface TagsState {
  tags: Tag[]
  productTags: Record<string, ProductTag[]>
  isLoading: boolean
  error: string | null
  isSubmitting: boolean
}

interface TagsActions {
  fetchTags: () => Promise<void>
  fetchProductTags: (productId: string) => Promise<void>
  createTag: (tagData: { name: string; description?: string; color?: string }) => Promise<void>
  addTagToProduct: (productId: string, tagId: string) => Promise<void>
  removeTagFromProduct: (productId: string, tagId: string) => Promise<void>
  clearError: () => void
  setLoading: (loading: boolean) => void
}

type TagsStore = TagsState & TagsActions

export const useTagsStore = create<TagsStore>()(
  immer((set) => ({
    tags: [],
    productTags: {},
    isLoading: false,
    error: null,
    isSubmitting: false,

    fetchTags: async () => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.tags.getAll()
        set((state) => {
          state.tags = response.data.tags || []
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch tags'
          state.isLoading = false
        })
        throw error
      }
    },

    fetchProductTags: async (productId: string) => {
      try {
        const response = await apiService.tags.getProductTags(productId)
        set((state) => {
          state.productTags[productId] = response.data.product_tags || []
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch product tags'
        })
        throw error
      }
    },

    createTag: async (tagData) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        const response = await apiService.tags.create(tagData)
        const newTag = response.data.tag

        set((state) => {
          state.tags.push(newTag)
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to create tag'
          state.isSubmitting = false
        })
        throw error
      }
    },

    addTagToProduct: async (productId: string, tagId: string) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        const response = await apiService.tags.addToProduct({ product_id: productId, tag_id: tagId })
        const newProductTag = response.data.product_tag

        set((state) => {
          if (!state.productTags[productId]) {
            state.productTags[productId] = []
          }
          state.productTags[productId].push(newProductTag)
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to add tag to product'
          state.isSubmitting = false
        })
        throw error
      }
    },

    removeTagFromProduct: async (productId: string, tagId: string) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        // Note: We need to add a remove endpoint to the API service
        // For now, we'll simulate the removal
        set((state) => {
          if (state.productTags[productId]) {
            state.productTags[productId] = state.productTags[productId].filter(
              pt => pt.tag_id !== tagId
            )
          }
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to remove tag from product'
          state.isSubmitting = false
        })
        throw error
      }
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