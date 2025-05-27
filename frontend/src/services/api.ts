import axios from 'axios'
import type { AxiosInstance, AxiosResponse } from 'axios'

// Create axios instance with base configuration
const api: AxiosInstance = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    // Get token from localStorage (Zustand persist storage)
    const authStorage = localStorage.getItem('auth-storage')
    if (authStorage) {
      try {
        const { state } = JSON.parse(authStorage)
        if (state?.token) {
          config.headers.Authorization = `Bearer ${state.token}`
        }
      } catch (error) {
        console.error('Error parsing auth storage:', error)
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response: AxiosResponse) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid - clear auth state
      localStorage.removeItem('auth-storage')
      window.location.href = '/auth/login'
    }
    return Promise.reject(error)
  }
)

// API service functions
export const apiService = {
  // Auth endpoints
  auth: {
    login: (email: string, password: string) =>
      api.post('/auth/login', { email, password }),
    register: (email: string, name: string, password: string) =>
      api.post('/auth/register', { email, name, password }),
    getProfile: () => api.get('/auth/profile'),
    updateProfile: (data: { name: string }) =>
      api.put('/auth/profile', data),
  },

  // Product endpoints
  products: {
    getAll: (params?: {
      page?: number
      limit?: number
      category?: string
      search?: string
      sort?: string
      order?: string
    }) => api.get('/products', { params }),
    getById: (id: string) => api.get(`/products/${id}`),
    getByCategory: (category: string, params?: { page?: number; limit?: number }) =>
      api.get(`/products/category/${category}`, { params }),
    search: (query: string, params?: { page?: number; limit?: number }) =>
      api.get('/products/search', { params: { q: query, ...params } }),
    getCategories: () => api.get('/products/categories'),
    getRecommendations: (params?: { limit?: number }) =>
      api.get('/products/recommendations', { params }),
  },

  // ML endpoints
  ml: {
    recommendations: {
      generate: (userId: string, algorithm?: string, limit?: number) =>
        api.post('/ml/recommendations/generate', { user_id: userId, algorithm, limit }),
      getUser: (userId: string, algorithm?: string, limit?: number) =>
        api.get(`/ml/recommendations/user/${userId}`, { params: { algorithm, limit } }),
      feedback: (userId: string, productId: string, feedbackType: string) =>
        api.post('/ml/recommendations/feedback', { user_id: userId, product_id: productId, feedback_type: feedbackType }),
      popular: (limit?: number) =>
        api.get('/ml/recommendations/popular', { params: { limit } }),
      similar: (productId: string, limit?: number) =>
        api.get(`/ml/recommendations/similar/${productId}`, { params: { limit } }),
    },
    search: {
      enhanced: (query: string, limit?: number) =>
        api.post('/ml/search/enhanced', { query, limit }),
      suggestions: (query: string) =>
        api.get('/ml/search/suggestions', { params: { query } }),
    },
    trends: {
      sales: (period?: string, category?: string) =>
        api.get('/ml/trends/sales', { params: { period, category } }),
      forecast: (days?: number, category?: string) =>
        api.get('/ml/trends/forecast', { params: { days, category } }),
      popular: (period?: string, limit?: number) =>
        api.get('/ml/trends/popular', { params: { period, limit } }),
    },
    chatbot: {
      message: (message: string, sessionId?: string, userId?: string) =>
        api.post('/ml/chatbot/message', { message, session_id: sessionId, user_id: userId }),
      intents: () => api.get('/ml/chatbot/intents'),
    },
  },

  // Cart endpoints (when implemented)
  cart: {
    get: () => api.get('/cart'),
    add: (productId: string, quantity: number) =>
      api.post('/cart/items', { product_id: productId, quantity }),
    update: (itemId: string, quantity: number) =>
      api.put(`/cart/items/${itemId}`, { quantity }),
    remove: (itemId: string) => api.delete(`/cart/items/${itemId}`),
    clear: () => api.delete('/cart'),
  },

  // Order endpoints (when implemented)
  orders: {
    getAll: () => api.get('/orders'),
    getById: (id: string) => api.get(`/orders/${id}`),
    create: (cartId: string) => api.post('/orders', { cart_id: cartId }),
    updateStatus: (id: string, status: string) =>
      api.put(`/orders/${id}/status`, { status }),
  },
}

export default api 