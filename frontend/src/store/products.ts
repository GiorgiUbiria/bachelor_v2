import { create } from 'zustand'
import { immer } from 'zustand/middleware/immer'
import { persist } from 'zustand/middleware'
import apiService from '../services/api'

export interface Product {
  id: string
  name: string
  description: string
  price: number
  stock: number
  category: string
  image_url?: string
  rating?: number
  reviews_count?: number
  tags?: string[]
  created_at: string
  updated_at: string
}

export interface ProductFilters {
  search?: string
  category?: string
  minPrice?: number
  maxPrice?: number
  tags?: string[]
  inStock?: boolean
  sortBy?: 'name' | 'price' | 'rating' | 'created_at'
  sortOrder?: 'asc' | 'desc'
  page?: number
  limit?: number
}

interface ProductsState {
  products: Product[]
  filteredProducts: Product[]
  currentProduct: Product | null
  categories: string[]
  recommendations: Product[]
  searchSuggestions: string[]
  filters: ProductFilters
  pagination: {
    currentPage: number
    totalPages: number
    totalItems: number
    itemsPerPage: number
  }
  isLoading: boolean
  isLoadingRecommendations: boolean
  error: string | null
  cache: Record<string, { data: Product[]; timestamp: number }>
}

interface ProductsActions {
  // Product fetching
  fetchProducts: (filters?: ProductFilters) => Promise<void>
  fetchProductById: (id: string) => Promise<void>
  fetchCategories: () => Promise<void>
  fetchRecommendations: (productId?: string) => Promise<void>
  fetchSearchSuggestions: (query: string) => Promise<void>
  
  // Filtering and search
  setFilters: (filters: Partial<ProductFilters>) => void
  clearFilters: () => void
  applyFilters: () => void
  searchProducts: (query: string) => Promise<void>
  
  // Cache management
  clearCache: () => void
  getCachedProducts: (key: string) => Product[] | null
  setCachedProducts: (key: string, products: Product[]) => void
  
  // Utility
  clearError: () => void
  clearCurrentProduct: () => void
}

type ProductsStore = ProductsState & ProductsActions

const defaultFilters: ProductFilters = {
  search: '',
  category: '',
  minPrice: undefined,
  maxPrice: undefined,
  tags: [],
  inStock: undefined,
  sortBy: 'created_at',
  sortOrder: 'desc',
  page: 1,
  limit: 12
}

export const useProductsStore = create<ProductsStore>()(
  persist(
    immer((set, get) => ({
      products: [],
      filteredProducts: [],
      currentProduct: null,
      categories: [],
      recommendations: [],
      searchSuggestions: [],
      filters: defaultFilters,
      pagination: {
        currentPage: 1,
        totalPages: 1,
        totalItems: 0,
        itemsPerPage: 12
      },
      isLoading: false,
      isLoadingRecommendations: false,
      error: null,
      cache: {},

      fetchProducts: async (filters) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          // Check cache first
          const cacheKey = JSON.stringify(filters || {})
          const cached = get().getCachedProducts(cacheKey)
          
          if (cached) {
            set((state) => {
              state.products = cached
              state.filteredProducts = cached
              state.isLoading = false
            })
            return
          }

          const response = await apiService.products.getAll(filters)
          const products = response.data.products || []
          
          set((state) => {
            state.products = products
            state.filteredProducts = products
            state.pagination = {
              currentPage: response.data.page || 1,
              totalPages: response.data.total_pages || 1,
              totalItems: response.data.total || 0,
              itemsPerPage: response.data.limit || 12
            }
            state.isLoading = false
          })

          // Cache the results
          get().setCachedProducts(cacheKey, products)
        } catch (error: any) {
          set((state) => {
            state.error = error.response?.data?.error || 'Failed to fetch products'
            state.isLoading = false
          })
        }
      },

      fetchProductById: async (id: string) => {
        set((state) => {
          state.isLoading = true
          state.error = null
        })

        try {
          // Check if product is already in the products array
          const existingProduct = get().products.find(p => p.id === id)
          if (existingProduct) {
            set((state) => {
              state.currentProduct = existingProduct
              state.isLoading = false
            })
            return
          }

          const response = await apiService.products.getById(id)
          set((state) => {
            state.currentProduct = response.data.product
            state.isLoading = false
          })
        } catch (error: any) {
          set((state) => {
            state.error = error.response?.data?.error || 'Failed to fetch product'
            state.isLoading = false
          })
        }
      },

      fetchCategories: async () => {
        try {
          const response = await apiService.products.getCategories()
          set((state) => {
            state.categories = response.data.categories || []
          })
        } catch (error: any) {
          console.error('Failed to fetch categories:', error)
        }
      },

      fetchRecommendations: async (productId) => {
        set((state) => {
          state.isLoadingRecommendations = true
        })

        try {
          const response = productId 
            ? await apiService.products.getSimilar(productId)
            : await apiService.products.getRecommendations()
          
          set((state) => {
            state.recommendations = response.data.products || []
            state.isLoadingRecommendations = false
          })
        } catch (error: any) {
          set((state) => {
            state.isLoadingRecommendations = false
          })
          console.error('Failed to fetch recommendations:', error)
        }
      },

      fetchSearchSuggestions: async (query: string) => {
        if (!query.trim()) {
          set((state) => {
            state.searchSuggestions = []
          })
          return
        }

        try {
          const response = await apiService.products.getSearchSuggestions(query)
          set((state) => {
            state.searchSuggestions = response.data.suggestions || []
          })
        } catch (error: any) {
          console.error('Failed to fetch search suggestions:', error)
        }
      },

      setFilters: (newFilters: Partial<ProductFilters>) => {
        set((state) => {
          state.filters = { ...state.filters, ...newFilters }
        })
      },

      clearFilters: () => {
        set((state) => {
          state.filters = { ...defaultFilters }
        })
      },

      applyFilters: () => {
        const { products, filters } = get()
        
        let filtered = [...products]

        // Apply search filter
        if (filters.search) {
          const searchLower = filters.search.toLowerCase()
          filtered = filtered.filter(product => 
            product.name.toLowerCase().includes(searchLower) ||
            product.description.toLowerCase().includes(searchLower)
          )
        }

        // Apply category filter
        if (filters.category) {
          filtered = filtered.filter(product => product.category === filters.category)
        }

        // Apply price filters
        if (filters.minPrice !== undefined) {
          filtered = filtered.filter(product => product.price >= filters.minPrice!)
        }
        if (filters.maxPrice !== undefined) {
          filtered = filtered.filter(product => product.price <= filters.maxPrice!)
        }

        // Apply stock filter
        if (filters.inStock !== undefined) {
          filtered = filtered.filter(product => filters.inStock ? product.stock > 0 : product.stock === 0)
        }

        // Apply sorting
        if (filters.sortBy) {
          filtered.sort((a, b) => {
            let aValue: any = a[filters.sortBy!]
            let bValue: any = b[filters.sortBy!]

            if (filters.sortBy === 'created_at') {
              aValue = new Date(aValue).getTime()
              bValue = new Date(bValue).getTime()
            }

            if (filters.sortOrder === 'desc') {
              return bValue > aValue ? 1 : -1
            } else {
              return aValue > bValue ? 1 : -1
            }
          })
        }

        set((state) => {
          state.filteredProducts = filtered
        })
      },

      searchProducts: async (query: string) => {
        set((state) => {
          state.filters.search = query
        })
        
        await get().fetchProducts(get().filters)
      },

      getCachedProducts: (key: string) => {
        const cache = get().cache[key]
        if (cache) {
          // Check if cache is still valid (5 minutes)
          const fiveMinutesAgo = Date.now() - 5 * 60 * 1000
          if (cache.timestamp > fiveMinutesAgo) {
            return cache.data
          }
        }
        return null
      },

      setCachedProducts: (key: string, products: Product[]) => {
        set((state) => {
          state.cache[key] = {
            data: products,
            timestamp: Date.now()
          }
        })
      },

      clearCache: () => {
        set((state) => {
          state.cache = {}
        })
      },

      clearError: () => {
        set((state) => {
          state.error = null
        })
      },

      clearCurrentProduct: () => {
        set((state) => {
          state.currentProduct = null
        })
      }
    })),
    {
      name: 'products-store',
      partialize: (state) => ({
        categories: state.categories,
        cache: state.cache
      })
    }
  )
) 