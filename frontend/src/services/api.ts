import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const ML_API_BASE_URL = import.meta.env.VITE_ML_API_URL || 'http://localhost:8000'

// Create axios instances
const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

const mlApi = axios.create({
  baseURL: ML_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use((config) => {
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
})

mlApi.interceptors.request.use((config) => {
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
})

export const apiService = {
  // Auth endpoints
  auth: {
    login: (credentials: { email: string; password: string }) =>
      api.post('/api/v1/auth/login', credentials),
    register: (userData: { name: string; email: string; password: string }) =>
      api.post('/api/v1/auth/register', userData),
    logout: () => api.post('/api/v1/auth/logout'),
    me: () => api.get('/api/v1/auth/me'),
  },

  // Product endpoints
  products: {
    getAll: (params?: { page?: number; limit?: number; category?: string; search?: string }) =>
      api.get('/api/v1/products', { params }),
    getById: (id: string) => api.get(`/api/v1/products/${id}`),
    create: (productData: any) => api.post('/api/v1/products', productData),
    update: (id: string, productData: any) => api.put(`/api/v1/products/${id}`, productData),
    delete: (id: string) => api.delete(`/api/v1/products/${id}`),
    getCategories: () => api.get('/api/v1/products/categories'),
    getRecommendations: (params?: { limit?: number }) =>
      api.get('/api/v1/products/recommendations', { params }),
  },

  // Cart endpoints
  cart: {
    get: () => api.get('/api/v1/cart'),
    addItem: (productId: string, quantity: number) =>
      api.post('/api/v1/cart/items', { product_id: productId, quantity }),
    updateItem: (itemId: string, quantity: number) =>
      api.put(`/api/v1/cart/items/${itemId}`, { quantity }),
    removeItem: (itemId: string) => api.delete(`/api/v1/cart/items/${itemId}`),
    clear: () => api.delete('/api/v1/cart'),
  },

  // Order endpoints
  orders: {
    getAll: (params?: { page?: number; limit?: number; status?: string }) =>
      api.get('/api/v1/orders', { params }),
    getById: (id: string) => api.get(`/api/v1/orders/${id}`),
    create: (orderData?: { shipping_address?: string; notes?: string }) =>
      api.post('/api/v1/orders', orderData || {}),
    updateStatus: (id: string, status: string) =>
      api.put(`/api/v1/orders/${id}/status`, { status }),
    cancel: (id: string) => api.put(`/api/v1/orders/${id}/cancel`),
  },

  // Analytics endpoints
  analytics: {
    dashboard: () => api.get('/api/v1/analytics/dashboard'),
    users: (params?: { period?: string }) => api.get('/api/v1/analytics/users', { params }),
    products: (params?: { period?: string }) => api.get('/api/v1/analytics/products', { params }),
    export: (params?: { format?: string; period?: string }) =>
      api.get('/api/v1/analytics/export', { params }),
  },

  // ML Service endpoints
  ml: {
    // Search endpoints
    search: {
      search: (query: string, params?: { 
        category?: string; 
        min_price?: number; 
        max_price?: number; 
        sort_by?: string; 
        limit?: number 
      }) =>
        mlApi.post('/search', { query, ...params }),
      suggestions: (query: string) =>
        mlApi.get('/search/suggestions', { params: { query } }),
      analytics: () => mlApi.get('/search/analytics'),
      reindex: () => mlApi.post('/search/reindex'),
      status: () => mlApi.get('/search/status'),
    },

    // Trends endpoints
    trends: {
      products: (params?: { period?: string; limit?: number }) =>
        mlApi.get('/trends/products', { params }),
      categories: (params?: { period?: string }) =>
        mlApi.get('/trends/categories', { params }),
      forecast: (productId: string, params?: { days?: number }) =>
        mlApi.get(`/trends/forecast/${productId}`, { params }),
      dashboard: () => mlApi.get('/trends/dashboard'),
      insights: () => mlApi.get('/trends/insights'),
    },

    // Recommendations endpoints
    recommendations: {
      user: (userId: string, params?: { limit?: number; algorithm?: string }) =>
        mlApi.get(`/recommendations/user/${userId}`, { params }),
      product: (productId: string, params?: { limit?: number }) =>
        mlApi.get(`/recommendations/product/${productId}`, { params }),
      retrain: () => mlApi.post('/recommendations/retrain'),
      status: () => mlApi.get('/recommendations/status'),
    },

    // General ML endpoints
    health: () => mlApi.get('/health'),
    models: () => mlApi.get('/models/status'),
  },
}

export default apiService 