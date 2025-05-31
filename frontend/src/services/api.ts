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
      api.post('/auth/login', credentials),
    register: (userData: { name: string; email: string; password: string }) =>
      api.post('/auth/register', userData),
    profile: () => api.get('/auth/profile'),
    updateProfile: (profileData: { name?: string; phone?: string }) => 
      api.put('/auth/profile', profileData),
  },

  products: {
    getAll: (params?: { page?: number; limit?: number; category?: string; search?: string }) =>
      api.get('/products', { params }),
    getById: (id: string) => api.get(`/products/${id}`),
    create: (productData: { 
      name: string; 
      description: string; 
      price: number; 
      stock: number; 
      category: string; 
      image_url?: string 
    }) => api.post('/products', productData),
    update: (id: string, productData: { 
      name?: string; 
      description?: string; 
      price?: number; 
      stock?: number; 
      category?: string; 
      image_url?: string 
    }) => api.put(`/products/${id}`, productData),
    delete: (id: string) => api.delete(`/products/${id}`),
    search: (params?: { q?: string; category?: string; page?: number; limit?: number }) =>
      api.get('/products/search', { params }),
    getRecommendations: (params?: { limit?: number }) =>
      api.get('/products/recommendations', { params }),
  },

  cart: {
    get: () => api.get('/cart'),
    addItem: (productId: string, quantity: number) =>
      api.post('/cart/add', { product_id: productId, quantity }),
    updateItem: (itemId: string, quantity: number) =>
      api.put(`/cart/item/${itemId}`, { quantity }),
    removeItem: (itemId: string) => api.delete(`/cart/item/${itemId}`),
    clear: () => api.delete('/cart/clear'),
  },

  orders: {
    getAll: (params?: { page?: number; limit?: number; status?: string }) =>
      api.get('/orders', { params }),
    getById: (id: string) => api.get(`/orders/${id}`),
    create: (orderData: { 
      cart_item_ids?: string[]; 
      payment_method: string; 
      shipping_address: string 
    }) => api.post('/orders', orderData),
    updateStatus: (id: string, status: 'pending' | 'processing' | 'shipped' | 'delivered' | 'cancelled') =>
      api.put(`/orders/${id}/status`, { status }),
    cancel: (id: string) => api.put(`/orders/${id}/cancel`),
    getStats: () => api.get('/orders/stats'),
  },

  comments: {
    getByProduct: (productId: string) => api.get(`/comments/${productId}`),
    add: (commentData: { product_id: string; content: string; rating: number }) =>
      api.post('/comments', commentData),
    update: (commentId: string, commentData: { content?: string; rating?: number }) =>
      api.put(`/comments/${commentId}`, commentData),
    delete: (commentId: string) => api.delete(`/comments/${commentId}`),
  },

  discounts: {
    getActive: () => api.get('/discounts/active'),
    create: (discountData: {
      discount_type: string;
      discount_value: number;
      start_date: string;
      end_date: string;
      product_id?: string;
      category?: string;
      min_order_amount?: number;
      max_discount_amount?: number;
      usage_limit?: number;
    }) => api.post('/discounts', discountData),
  },

  favorites: {
    get: () => api.get('/favorites'),
    add: (productId: string) => api.post('/favorites', { product_id: productId }),
    remove: (productId: string) => api.delete(`/favorites/${productId}`),
  },

  tags: {
    getAll: () => api.get('/tags'),
    create: (tagData: { name: string; description?: string; color?: string }) =>
      api.post('/tags', tagData),
    getByProduct: (productId: string) => api.get(`/tags/products/${productId}`),
    addToProduct: (productId: string, tagId: string) =>
      api.post('/tags/products', { product_id: productId, tag_id: tagId }),
  },

  upvotes: {
    add: (productId: string) => api.post('/upvotes', { product_id: productId }),
    getByProduct: (productId: string) => api.get(`/upvotes/${productId}`),
    remove: (productId: string) => api.delete(`/upvotes/${productId}`),
  },

  analytics: {
    dashboard: () => api.get('/analytics/dashboard'),
    user: (params?: { period?: string }) => api.get('/analytics/user', { params }),
    products: (params?: { period?: string }) => api.get('/analytics/products', { params }),
    trends: () => api.get('/analytics/trends'),
    search: () => api.get('/analytics/search'),
    recommendationMetrics: () => api.get('/analytics/recommendations/metrics'),
    export: (params?: { format?: string; period?: string }) =>
      api.get('/analytics/export', { params }),
  },

  ml: {
    initializeServices: () => api.post('/ml/initialize-services'),
    
    autoTagging: {
      autoTag: (productData: any) => api.post('/ml/auto-tagging/auto-tag', productData),
      getInsights: () => api.get('/ml/auto-tagging/insights'),
      suggest: (productId: string) => api.get(`/ml/auto-tagging/suggest/${productId}`),
    },

    sentiment: {
      analyzeProduct: (productId: string) => api.get(`/ml/sentiment/product/${productId}`),
      analyzeCategory: (category: string) => api.get(`/ml/sentiment/category/${category}`),
      getInsights: () => api.get('/ml/sentiment/insights'),
    },

    smartDiscounts: {
      getInsights: () => api.get('/ml/smart-discounts/insights'),
      suggestForProduct: (productId: string) => api.get(`/ml/smart-discounts/suggest/product/${productId}`),
      suggestForCategory: (category: string) => api.get(`/ml/smart-discounts/suggest/category/${category}`),
    },
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
  },
}

export default apiService 