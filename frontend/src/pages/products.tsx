import { useState, useEffect } from 'react'
import { Link } from 'react-router'
import { useAuthStore } from '../store/auth'
import { useCartStore } from '../store/cart'
import { apiService } from '../services/api'
import LoadingSpinner from '../components/LoadingSpinner'

interface Product {
  id: string
  name: string
  description: string
  price: number
  category: string
  stock: number
  image_url: string
  created_at: string
  updated_at: string
}

interface ProductResponse {
  products: Product[]
  pagination: {
    page: number
    limit: number
    total: number
    total_pages: number
  }
}

// interface RecommendationResponse {
//   recommendations: Array<{
//     product_id: string
//     score: number
//     algorithm: string
//   }>
//   user_id: string
//   algorithm: string
//   total: number
// }

export default function Products() {
  const { isAuthenticated } = useAuthStore()
  const { addItem } = useCartStore()
  const [products, setProducts] = useState<Product[]>([])
  const [recommendations, setRecommendations] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [categories, setCategories] = useState<string[]>([])
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [addingToCart, setAddingToCart] = useState<string | null>(null)
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null)

  useEffect(() => {
    fetchProducts()
    fetchCategories()
    if (isAuthenticated) {
      fetchRecommendations()
    }
  }, [currentPage, selectedCategory, isAuthenticated])

  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => {
        setMessage(null)
      }, 3000)
      return () => clearTimeout(timer)
    }
  }, [message])

  const fetchProducts = async () => {
    try {
      setLoading(true)
      const params: any = {
        page: currentPage,
        limit: 12
      }
      
      if (selectedCategory) {
        params.category = selectedCategory
      }
      
      if (searchQuery) {
        params.search = searchQuery
      }

      const response = await apiService.products.getAll(params)
      const data: ProductResponse = response.data
      
      setProducts(data.products)
      setTotalPages(data.pagination.total_pages)
    } catch (error) {
      console.error('Error fetching products:', error)
    } finally {
      setLoading(false)
    }
  }

  const fetchCategories = async () => {
    try {
      const response = await apiService.products.getCategories()
      setCategories(response.data.categories)
    } catch (error) {
      console.error('Error fetching categories:', error)
    }
  }

  const fetchRecommendations = async () => {
    try {
      const response = await apiService.products.getRecommendations({ limit: 6 })
      const data = response.data
      if (data.recommendations) {
        setRecommendations(data.recommendations.map((r: any) => r.product_id || r.id))
      }
    } catch (error) {
      console.error('Error fetching recommendations:', error)
    }
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setCurrentPage(1)
    fetchProducts()
  }

  const handleCategoryChange = (category: string) => {
    setSelectedCategory(category)
    setCurrentPage(1)
  }

  const handleAddToCart = async (productId: string) => {
    if (!isAuthenticated) {
      setMessage({ type: 'error', text: 'Please log in to add items to cart' })
      return
    }

    try {
      setAddingToCart(productId)
      await addItem(productId, 1)
      setMessage({ type: 'success', text: 'Item added to cart successfully!' })
    } catch (error) {
      setMessage({ type: 'error', text: 'Failed to add item to cart' })
    } finally {
      setAddingToCart(null)
    }
  }

  const isRecommended = (productId: string) => {
    return recommendations.includes(productId)
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">Products</h1>
        
        {/* Message */}
        {message && (
          <div className={`mb-4 p-3 rounded-md ${
            message.type === 'success' 
              ? 'bg-green-50 border border-green-200 text-green-700' 
              : 'bg-red-50 border border-red-200 text-red-700'
          }`}>
            {message.text}
          </div>
        )}
        
        {/* Search and Filters */}
        <div className="flex flex-col md:flex-row gap-4 mb-6">
          <form onSubmit={handleSearch} className="flex-1">
            <div className="flex">
              <input
                type="text"
                placeholder="Search products..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="flex-1 px-4 py-2 border border-gray-300 rounded-l-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              />
              <button
                type="submit"
                className="px-6 py-2 bg-blue-600 text-white rounded-r-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                Search
              </button>
            </div>
          </form>
          
          <select
            value={selectedCategory}
            onChange={(e) => handleCategoryChange(e.target.value)}
            className="px-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">All Categories</option>
            {categories.map((category) => (
              <option key={category} value={category}>
                {category}
              </option>
            ))}
          </select>
        </div>

        {/* ML Recommendations Banner */}
        {isAuthenticated && recommendations.length > 0 && (
          <div className="bg-gradient-to-r from-purple-50 to-blue-50 border border-purple-200 rounded-lg p-4 mb-6">
            <div className="flex items-center">
              <span className="text-purple-600 text-xl mr-2">🤖</span>
              <div>
                <h3 className="text-lg font-semibold text-purple-900">AI Recommendations</h3>
                <p className="text-purple-700">Products recommended just for you using machine learning</p>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Loading State */}
      {loading ? (
        <div className="flex justify-center items-center py-12">
          <LoadingSpinner size="lg" />
        </div>
      ) : (
        <>
          {/* Products Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 mb-8">
            {products.map((product) => (
              <div
                key={product.id}
                className={`bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow ${
                  isRecommended(product.id) ? 'ring-2 ring-purple-500' : ''
                }`}
              >
                {isRecommended(product.id) && (
                  <div className="bg-gradient-to-r from-purple-500 to-blue-500 text-white text-xs font-semibold px-3 py-1">
                    🤖 AI Recommended
                  </div>
                )}
                
                <div className="aspect-w-1 aspect-h-1 w-full">
                  <img
                    src={product.image_url || '/api/placeholder/300/300'}
                    alt={product.name}
                    className="w-full h-48 object-cover"
                  />
                </div>
                
                <div className="p-4">
                  <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
                    {product.name}
                  </h3>
                  <p className="text-gray-600 text-sm mb-3 line-clamp-2">
                    {product.description}
                  </p>
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-2xl font-bold text-blue-600">
                      ${product.price.toFixed(2)}
                    </span>
                    <span className="text-sm text-gray-500">
                      {product.category}
                    </span>
                  </div>
                  
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm text-gray-600">
                      Stock: {product.stock}
                    </span>
                    {product.stock === 0 && (
                      <span className="text-red-500 text-sm font-medium">Out of Stock</span>
                    )}
                  </div>
                  
                  <div className="space-y-2">
                    <button
                      onClick={() => handleAddToCart(product.id)}
                      disabled={addingToCart === product.id || product.stock === 0}
                      className="w-full bg-green-600 text-white py-2 px-4 rounded-md hover:bg-green-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                    >
                      {addingToCart === product.id ? (
                        <>
                          <LoadingSpinner size="sm" className="mr-2" />
                          Adding...
                        </>
                      ) : product.stock === 0 ? (
                        'Out of Stock'
                      ) : (
                        '🛒 Add to Cart'
                      )}
                    </button>
                    
                    <Link
                      to={`/products/${product.id}`}
                      className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors text-center block"
                    >
                      View Details
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex justify-center items-center space-x-2">
              <button
                onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                disabled={currentPage === 1}
                className="px-3 py-2 border border-gray-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Previous
              </button>
              
              <span className="px-4 py-2 text-gray-700">
                Page {currentPage} of {totalPages}
              </span>
              
              <button
                onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                disabled={currentPage === totalPages}
                className="px-3 py-2 border border-gray-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                Next
              </button>
            </div>
          )}

          {/* Empty State */}
          {products.length === 0 && !loading && (
            <div className="text-center py-12">
              <div className="text-gray-400 text-6xl mb-4">📦</div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">No products found</h3>
              <p className="text-gray-600">Try adjusting your search or filter criteria</p>
            </div>
          )}
        </>
      )}
    </div>
  )
} 