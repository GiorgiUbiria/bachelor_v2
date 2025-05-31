import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1'
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
  health: () => api.get('/api/v1/health'),
  corsTest: () => api.get('/api/v1/cors-test'),

  auth: {
    login: (credentials: { email: string; password: string }) =>
      api.post('/api/v1/auth/login', credentials),
    register: (userData: { name: string; email: string; password: string }) =>
      api.post('/api/v1/auth/register', userData),
    profile: () => api.get('/api/v1/auth/profile'),
    updateProfile: (profileData: { name?: string; phone?: string }) => 
      api.put('/api/v1/auth/profile', profileData),
  },

  products: {
    getAll: (params?: { page?: number; limit?: number; category?: string; search?: string }) =>
      api.get('/api/v1/products', { params }),
    getById: (id: string) => api.get(`/api/v1/products/${id}`),
    create: (productData: any) => api.post('/api/v1/products', productData),
    update: (id: string, productData: any) => api.put(`/api/v1/products/${id}`, productData),
    delete: (id: string) => api.delete(`/api/v1/products/${id}`),
    search: (params: any) => api.get('/api/v1/products/search', { params }),
    getRecommendations: (params?: any) => api.get('/api/v1/products/recommendations', { params }),
    getSimilar: (productId: string) => api.get(`/api/v1/products/${productId}/similar`),
    getSearchSuggestions: (query: string) => 
      api.get('/api/v1/ml/search/suggestions', { params: { q: query, limit: 10 } }),
    getCategories: () => api.get('/api/v1/products/categories'),
  },

  cart: {
    get: () => api.get('/api/v1/cart'),
    add: (productId: string, quantity: number) =>
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
    create: (orderData: { cart_item_ids?: string[]; payment_method: string; shipping_address: string }) => 
      api.post('/api/v1/orders', orderData),
    cancel: (id: string) => api.put(`/api/v1/orders/${id}/cancel`),
    updateStatus: (id: string, status: string) => 
      api.put(`/api/v1/orders/${id}/status`, { status }),
    getStats: () => api.get('/api/v1/orders/stats'),
  },

  comments: {
    getByProduct: (productId: string) => api.get(`/api/v1/comments/${productId}`),
    add: (data: { product_id: string; content: string; rating: number }) => 
      api.post('/api/v1/comments', data),
    update: (commentId: string, data: { content?: string; rating?: number }) => 
      api.put(`/api/v1/comments/${commentId}`, data),
    delete: (commentId: string) => api.delete(`/api/v1/comments/${commentId}`),
  },

  favorites: {
    getAll: () => api.get('/api/v1/favorites'),
    add: (productId: string) => api.post('/api/v1/favorites', { product_id: productId }),
    remove: (productId: string) => api.delete(`/api/v1/favorites/${productId}`),
  },

  upvotes: {
    getByProduct: (productId: string) => api.get(`/api/v1/upvotes/${productId}`),
    add: (productId: string) => api.post('/api/v1/upvotes', { product_id: productId }),
    remove: (productId: string) => api.delete(`/api/v1/upvotes/${productId}`),
  },

  tags: {
    getAll: () => api.get('/api/v1/tags'),
    create: (data: { name: string; description?: string; color?: string }) => 
      api.post('/api/v1/tags', data),
    getProductTags: (productId: string) => api.get(`/api/v1/tags/products/${productId}`),
    addToProduct: (data: { product_id: string; tag_id: string }) => 
      api.post('/api/v1/tags/products', data),
  },

  discounts: {
    getActive: () => api.get('/api/v1/discounts/active'),
    create: (data: {
      product_id?: string;
      category?: string;
      discount_type: string;
      discount_value: number;
      start_date: string;
      end_date: string;
      min_order_amount?: number;
      max_discount_amount?: number;
      usage_limit?: number;
    }) => api.post('/api/v1/discounts', data),
  },

  analytics: {
    dashboard: () => api.get('/api/v1/analytics/dashboard'),
    user: () => api.get('/api/v1/analytics/user'),
    products: () => api.get('/api/v1/analytics/products'),
    search: () => api.get('/api/v1/analytics/search'),
    trends: () => api.get('/api/v1/analytics/trends'),
    export: (params?: any) => api.get('/api/v1/analytics/export', { params }),
    recommendations: () => api.get('/api/v1/analytics/recommendations/metrics'),
  },

  mlService: {
    initialize: () => api.post('/api/v1/ml/initialize-services'),
    
    autoTagging: {
      suggest: (productId: string) => api.get(`/api/v1/ml/auto-tagging/suggest/${productId}`),
      autoTag: (limit?: number) => api.post('/api/v1/ml/auto-tagging/auto-tag', { limit }),
      insights: () => api.get('/api/v1/ml/auto-tagging/insights'),
    },

    sentiment: {
      product: (productId: string) => api.get(`/api/v1/ml/sentiment/product/${productId}`),
      category: (category: string) => api.get(`/api/v1/ml/sentiment/category/${category}`),
      insights: () => api.get('/api/v1/ml/sentiment/insights'),
    },

    smartDiscounts: {
      suggestProduct: (productId: string) => api.get(`/api/v1/ml/smart-discounts/suggest/product/${productId}`),
      suggestCategory: (category: string) => api.get(`/api/v1/ml/smart-discounts/suggest/category/${category}`),
      insights: () => api.get('/api/v1/ml/smart-discounts/insights'),
    },

    search: {
      suggestions: (query: string, limit?: number) => 
        api.get('/api/v1/ml/search/suggestions', { params: { q: query, limit } }),
      popular: (limit?: number) => 
        api.get('/api/v1/ml/search/popular', { params: { limit } }),
      categories: () => api.get('/api/v1/ml/search/categories'),
      analytics: (days?: number) => 
        api.get('/api/v1/ml/search/analytics', { params: { days } }),
    },
  },
}

export default apiService 