import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const ML_API_BASE_URL = import.meta.env.VITE_ML_API_URL || 'http://localhost:8000'

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
  health: () => api.get('/health'),
  corsTest: () => api.get('/cors-test'),

  auth: {
    login: (credentials: { email: string; password: string }) =>
      api.post('/api/v1/auth/login', credentials),
    register: (userData: { name: string; email: string; password: string }) =>
      api.post('/api/v1/auth/register', userData),
    logout: () => api.post('/api/v1/auth/logout'),
    profile: () => api.get('/api/v1/auth/profile'),
    updateProfile: (profileData: any) => api.put('/api/v1/auth/profile', profileData),
    me: () => api.get('/api/v1/auth/profile'),
  },

  products: {
    getAll: (params?: { page?: number; limit?: number; category?: string; search?: string }) =>
      api.get('/api/v1/products', { params }),
    getById: (id: string) => api.get(`/api/v1/products/${id}`),
    create: (productData: any) => api.post('/api/v1/products', productData),
    update: (id: string, productData: any) => api.put(`/api/v1/products/${id}`, productData),
    delete: (id: string) => api.delete(`/api/v1/products/${id}`),
    getCategories: () => api.get('/api/v1/products/categories'),
    search: (params?: { q?: string; category?: string; page?: number; limit?: number }) =>
      api.get('/api/v1/products/search', { params }),
    getRecommendations: (params?: { limit?: number }) =>
      api.get('/api/v1/products/recommendations', { params }),
    getByCategory: (category: string, params?: { page?: number; limit?: number }) =>
      api.get(`/api/v1/products/category/${category}`, { params }),
  },

  cart: {
    get: () => api.get('/api/v1/cart'),
    addItem: (productId: string, quantity: number) =>
      api.post('/api/v1/cart/add', { product_id: productId, quantity }),
    updateItem: (itemId: string, quantity: number) =>
      api.put(`/api/v1/cart/item/${itemId}`, { quantity }),
    removeItem: (itemId: string) => api.delete(`/api/v1/cart/item/${itemId}`),
    clear: () => api.delete('/api/v1/cart/clear'),
  },

  orders: {
    getAll: (params?: { page?: number; limit?: number; status?: string }) =>
      api.get('/api/v1/orders', { params }),
    getById: (id: string) => api.get(`/api/v1/orders/${id}`),
    create: (orderData: { payment_method: string; shipping_address: string; notes?: string }) =>
      api.post('/api/v1/orders', orderData),
    updateStatus: (id: string, status: string) =>
      api.put(`/api/v1/orders/${id}/status`, { status }),
    cancel: (id: string) => api.put(`/api/v1/orders/${id}/cancel`),
    getStats: () => api.get('/api/v1/orders/stats'),
  },

  analytics: {
    dashboard: () => api.get('/api/v1/analytics/dashboard'),
    user: (params?: { period?: string }) => api.get('/api/v1/analytics/user', { params }),
    users: (params?: { period?: string }) => api.get('/api/v1/analytics/user', { params }), // Alias
    products: (params?: { period?: string }) => api.get('/api/v1/analytics/products', { params }),
    trends: () => api.get('/api/v1/analytics/trends'),
    search: () => api.get('/api/v1/analytics/search'),
    recommendationMetrics: () => api.get('/api/v1/analytics/recommendations/metrics'),
    export: (params?: { format?: string; period?: string }) =>
      api.get('/api/v1/analytics/export', { params }),
  },

  ml: {
    status: () => api.get('/api/v1/ml/status'),
    train: () => api.post('/api/v1/ml/train'),
  },

  chat: {
    message: (message: string) => api.post('/api/v1/chat/message', { message }),
  },

  mlService: {
    root: () => mlApi.get('/'),
    health: () => mlApi.get('/health'),

    search: {
      search: (params: { 
        q: string;
        user_id?: string;
        category?: string; 
        min_price?: number; 
        max_price?: number; 
        sort_by?: string; 
        limit?: number 
      }) =>
        mlApi.get('/search', { params }),
      suggestions: (query: string, limit?: number) =>
        mlApi.get('/search/suggestions', { params: { q: query, limit } }),
      analytics: (days?: number) => 
        mlApi.get('/search/analytics', { params: { days } }),
      reindex: () => mlApi.post('/search/reindex'),
      status: () => mlApi.get('/search/status'),
      categories: () => mlApi.get('/search/categories'),
      popular: (limit?: number) => mlApi.get('/search/popular', { params: { limit } }),
    },

    recommendations: {
      generate: (data: { user_id: string; algorithm?: string; limit?: number }) =>
        mlApi.post('/generate', data),
      user: (userId: string, params?: { algorithm?: string; limit?: number }) =>
        mlApi.get(`/user/${userId}`, { params }),
      similar: (productId: string, limit?: number) =>
        mlApi.get(`/similar/${productId}`, { params: { limit } }),
      popular: (limit?: number) => mlApi.get('/popular', { params: { limit } }),
      train: () => mlApi.post('/train'),
      retrain: () => mlApi.post('/train'), // Alias
      status: () => mlApi.get('/status'),
    },

    trends: {
      sales: (params?: { period?: string; category?: string }) =>
        mlApi.get('/sales', { params }),
      forecast: (params?: { days?: number; category?: string }) =>
        mlApi.get('/forecast', { params }),
      popular: (params?: { period?: string; limit?: number }) =>
        mlApi.get('/popular', { params }),
      
      products: (limit?: number) =>
        mlApi.get('/trends/products', { params: { limit } }),
      categories: () => mlApi.get('/trends/categories'),
      search: () => mlApi.get('/trends/search'),
      seasonal: () => mlApi.get('/trends/seasonal'),
      dashboard: () => mlApi.get('/trends/dashboard'),
      forecastDemand: (productId: string, daysAhead?: number) =>
        mlApi.get(`/forecast/demand/${productId}`, { params: { days_ahead: daysAhead } }),
      retrain: () => mlApi.post('/trends/retrain'),
      status: () => mlApi.get('/trends/status'),
    },

    chatbot: {
      message: (data: { message: string; session_id?: string; user_id?: string }) =>
        mlApi.post('/message', data),
      intents: () => mlApi.get('/intents'),
      train: () => mlApi.post('/train'),
    },
  },
}

export default apiService 