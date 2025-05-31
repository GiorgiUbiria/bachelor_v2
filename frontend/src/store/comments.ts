import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import apiService from '../services/api'

export interface Comment {
  id: string
  user_id: string
  product_id: string
  content: string
  rating: number
  user_name: string
  created_at: string
  updated_at: string
}

interface CommentsState {
  commentsByProduct: Record<string, Comment[]>
  isLoading: boolean
  error: string | null
  isSubmitting: boolean
}

interface CommentsActions {
  fetchComments: (productId: string) => Promise<void>
  addComment: (commentData: { product_id: string; content: string; rating: number }) => Promise<void>
  updateComment: (commentId: string, data: { content?: string; rating?: number }) => Promise<void>
  deleteComment: (commentId: string, productId: string) => Promise<void>
  clearError: () => void
  setLoading: (loading: boolean) => void
}

type CommentsStore = CommentsState & CommentsActions

export const useCommentsStore = create<CommentsStore>()(
  immer((set) => ({
    commentsByProduct: {},
    isLoading: false,
    error: null,
    isSubmitting: false,

    fetchComments: async (productId: string) => {
      set((state) => {
        state.isLoading = true
        state.error = null
      })

      try {
        const response = await apiService.comments.getByProduct(productId)
        set((state) => {
          state.commentsByProduct[productId] = response.data.comments || []
          state.isLoading = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to fetch comments'
          state.isLoading = false
        })
        throw error
      }
    },

    addComment: async (commentData) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        const response = await apiService.comments.add(commentData)
        const newComment = response.data.comment

        set((state) => {
          if (!state.commentsByProduct[commentData.product_id]) {
            state.commentsByProduct[commentData.product_id] = []
          }
          state.commentsByProduct[commentData.product_id].unshift(newComment)
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to add comment'
          state.isSubmitting = false
        })
        throw error
      }
    },

    updateComment: async (commentId: string, data) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        const response = await apiService.comments.update(commentId, data)
        const updatedComment = response.data.comment

        set((state) => {
          // Find and update the comment in the appropriate product's comments
          Object.keys(state.commentsByProduct).forEach(productId => {
            const commentIndex = state.commentsByProduct[productId].findIndex(
              comment => comment.id === commentId
            )
            if (commentIndex !== -1) {
              state.commentsByProduct[productId][commentIndex] = updatedComment
            }
          })
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to update comment'
          state.isSubmitting = false
        })
        throw error
      }
    },

    deleteComment: async (commentId: string, productId: string) => {
      set((state) => {
        state.isSubmitting = true
        state.error = null
      })

      try {
        await apiService.comments.delete(commentId)

        set((state) => {
          if (state.commentsByProduct[productId]) {
            state.commentsByProduct[productId] = state.commentsByProduct[productId].filter(
              comment => comment.id !== commentId
            )
          }
          state.isSubmitting = false
        })
      } catch (error: any) {
        set((state) => {
          state.error = error.response?.data?.error || 'Failed to delete comment'
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