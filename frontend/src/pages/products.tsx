import { useEffect, useState } from 'react'
import { useSearchParams, Link } from 'react-router'
import { 
  Search, 
  Grid, 
  List, 
  Star, 
  ShoppingCart, 
  TrendingUp, 
  Sparkles
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Upvotes } from '../components/upvotes'

import apiService from '../services/api'
import { useAuthStore } from '../store/auth'
import { useCartStore } from '../store/cart'
import { useUIStore } from '../store/ui'

interface Product {
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
}

interface SearchSuggestion {
  query: string
  type: string
}

export default function ProductsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const { isAuthenticated } = useAuthStore()
  const { addItem } = useCartStore()
  const { addToast } = useUIStore()
  
  const [products, setProducts] = useState<Product[]>([])
  const [recommendations, setRecommendations] = useState<Product[]>([])
  const [searchSuggestions, setSuggestions] = useState<SearchSuggestion[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [searchLoading, setSearchLoading] = useState(false)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  
  const [searchQuery, setSearchQuery] = useState(searchParams.get('search') || '')
  const [selectedCategory, setSelectedCategory] = useState(searchParams.get('category') || '')
  const [sortBy, setSortBy] = useState(searchParams.get('sort') || 'name')
  const [showSuggestions, setShowSuggestions] = useState(false)
  
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const itemsPerPage = 12

  // Fetch products
  const fetchProducts = async (params: any = {}) => {
    try {
      setSearchLoading(true)
      const response = await apiService.products.getAll({
        page: currentPage,
        limit: itemsPerPage,
        category: selectedCategory || undefined,
        search: searchQuery || undefined,
        ...params
      })
      
      setProducts(response.data.products || [])
      setTotalPages(Math.ceil((response.data.total || 0) / itemsPerPage))
    } catch (error) {
      console.error('Error fetching products:', error)
      addToast({
        type: 'error',
        description: 'Failed to load products'
      })
    } finally {
      setSearchLoading(false)
    }
  }

  // Fetch search suggestions
  const fetchSuggestions = async (query: string) => {
    if (query.length < 2) {
      setSuggestions([])
      return
    }
    
    try {
      const response = await apiService.mlService.search.suggestions(query, 5)
      setSuggestions(response.data.suggestions || [])
    } catch (error) {
      console.error('Error fetching suggestions:', error)
    }
  }

  // Fetch categories
  const fetchCategories = async () => {
    try {
      const response = await apiService.mlService.search.categories()
      setCategories(response.data.categories || [])
    } catch (error) {
      console.error('Error fetching categories:', error)
    }
  }

  // Fetch recommendations
  const fetchRecommendations = async () => {
    if (!isAuthenticated) return
    
    try {
      const response = await apiService.products.getRecommendations({ limit: 4 })
      setRecommendations(response.data.products || [])
    } catch (error) {
      console.error('Error fetching recommendations:', error)
    }
  }

  useEffect(() => {
    const loadData = async () => {
      setLoading(true)
      await Promise.all([
        fetchProducts(),
        fetchCategories(),
        fetchRecommendations()
      ])
      setLoading(false)
    }
    
    loadData()
  }, [currentPage, selectedCategory, sortBy, isAuthenticated])

  useEffect(() => {
    const delayedSearch = setTimeout(() => {
      if (searchQuery !== searchParams.get('search')) {
        fetchProducts()
      }
    }, 500)

    return () => clearTimeout(delayedSearch)
  }, [searchQuery])

  useEffect(() => {
    const delayedSuggestions = setTimeout(() => {
      fetchSuggestions(searchQuery)
    }, 300)

    return () => clearTimeout(delayedSuggestions)
  }, [searchQuery])

  // Update URL params
  useEffect(() => {
    const params = new URLSearchParams()
    if (searchQuery) params.set('search', searchQuery)
    if (selectedCategory) params.set('category', selectedCategory)
    if (sortBy !== 'name') params.set('sort', sortBy)
    setSearchParams(params)
  }, [searchQuery, selectedCategory, sortBy, setSearchParams])

  const handleAddToCart = async (productId: string) => {
    if (!isAuthenticated) {
      addToast({
        type: 'error',
        description: 'Please login to add items to cart'
      })
      return
    }

    try {
      await addItem(productId, 1)
      addToast({
        type: 'success',
        description: 'Product added to cart'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to add product to cart'
      })
    }
  }

  const handleSearch = (query: string) => {
    setSearchQuery(query)
    setShowSuggestions(false)
    setCurrentPage(1)
  }

  const ProductCard = ({ product, variant = 'grid' }: { product: Product; variant?: 'grid' | 'list' }) => (
    <Card className={`hover:shadow-lg transition-shadow ${variant === 'list' ? 'flex' : ''}`}>
      <div className={`${variant === 'grid' ? 'aspect-square' : 'w-48 h-48'} bg-muted ${variant === 'grid' ? 'rounded-t-lg' : 'rounded-l-lg'} flex items-center justify-center overflow-hidden`}>
        {product.image_url ? (
          <img 
            src={product.image_url} 
            alt={product.name}
            className={`w-full h-full object-cover ${variant === 'grid' ? 'rounded-t-lg' : 'rounded-l-lg'}`}
          />
        ) : (
          <div className="text-muted-foreground text-center">
            <div className="text-4xl mb-2">📦</div>
            <div className="text-sm">No Image</div>
          </div>
        )}
      </div>
      
      <CardContent className={`p-4 ${variant === 'list' ? 'flex-1' : ''}`}>
        <div className="flex items-start justify-between mb-2">
          <h3 className={`font-semibold ${variant === 'grid' ? 'line-clamp-1' : 'line-clamp-2'}`}>
            {product.name}
          </h3>
          <Badge variant="secondary">{product.category}</Badge>
        </div>
        
        <p className={`text-muted-foreground text-sm mb-3 ${variant === 'grid' ? 'line-clamp-2' : 'line-clamp-3'}`}>
          {product.description}
        </p>
        
        {product.tags && product.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mb-3">
            {product.tags.slice(0, 3).map((tag, index) => (
              <Badge key={index} variant="outline" className="text-xs">
                {tag}
              </Badge>
            ))}
          </div>
        )}
        
        <div className="flex items-center justify-between mb-3">
          <span className="text-lg font-bold">${product.price}</span>
          <div className="flex items-center gap-2">
            {product.rating && (
              <div className="flex items-center gap-1">
                <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                <span className="text-sm">{product.rating}</span>
                {product.reviews_count && (
                  <span className="text-xs text-muted-foreground">({product.reviews_count})</span>
                )}
              </div>
            )}
            <Upvotes productId={product.id} variant="minimal" />
          </div>
        </div>
        
        <div className="flex gap-2">
          <Button size="sm" className="flex-1" asChild>
            <Link to={`/products/${product.id}`}>View Details</Link>
          </Button>
          <Button 
            size="sm" 
            variant="outline"
            onClick={() => handleAddToCart(product.id)}
            disabled={product.stock === 0}
          >
            <ShoppingCart className="h-4 w-4" />
          </Button>
        </div>
        
        {product.stock === 0 && (
          <Badge variant="destructive" className="mt-2 w-full justify-center">
            Out of Stock
          </Badge>
        )}
      </CardContent>
    </Card>
  )

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-4">Products</h1>
        <p className="text-muted-foreground">
          Discover our AI-curated collection of products
        </p>
      </div>

      {/* Search and Filters */}
      <div className="mb-8 space-y-4">
        <div className="flex flex-col lg:flex-row gap-4">
          {/* Search Bar */}
          <div className="relative flex-1">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search products with AI-powered suggestions..."
                value={searchQuery}
                onChange={(e) => {
                  setSearchQuery(e.target.value)
                  setShowSuggestions(true)
                }}
                onFocus={() => setShowSuggestions(true)}
                onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
                className="pl-10"
              />
            </div>
            
            {/* Search Suggestions */}
            {showSuggestions && searchSuggestions.length > 0 && (
              <Card className="absolute top-full left-0 right-0 z-10 mt-1">
                <CardContent className="p-2">
                  {searchSuggestions.map((suggestion, index) => (
                    <button
                      key={index}
                      className="w-full text-left px-3 py-2 hover:bg-muted rounded-sm text-sm"
                      onClick={() => handleSearch(suggestion.query)}
                    >
                      <Search className="inline h-3 w-3 mr-2" />
                      {suggestion.query}
                      <Badge variant="outline" className="ml-2 text-xs">
                        {suggestion.type}
                      </Badge>
                    </button>
                  ))}
                </CardContent>
              </Card>
            )}
          </div>

          {/* Filters */}
          <div className="flex gap-2">
            <Select value={selectedCategory || "all"} onValueChange={(value) => setSelectedCategory(value === "all" ? "" : value)}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="All Categories" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Categories</SelectItem>
                {categories.map((category) => (
                  <SelectItem key={category} value={category}>
                    {category}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select value={sortBy || "name"} onValueChange={setSortBy}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="Sort by" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="name">Name</SelectItem>
                <SelectItem value="price_asc">Price: Low to High</SelectItem>
                <SelectItem value="price_desc">Price: High to Low</SelectItem>
                <SelectItem value="rating">Rating</SelectItem>
                <SelectItem value="newest">Newest</SelectItem>
              </SelectContent>
            </Select>

            <div className="flex border rounded-md">
              <Button
                variant={viewMode === 'grid' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('grid')}
              >
                <Grid className="h-4 w-4" />
              </Button>
              <Button
                variant={viewMode === 'list' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('list')}
              >
                <List className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
        {/* Sidebar */}
        <div className="lg:col-span-1">
          {/* ML Recommendations */}
          {isAuthenticated && recommendations.length > 0 && (
            <Card className="mb-6">
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Sparkles className="h-5 w-5 text-purple-500" />
                  Recommended for You
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {recommendations.map((product) => (
                  <div key={product.id} className="flex gap-3">
                    <div className="w-16 h-16 bg-muted rounded-lg flex items-center justify-center overflow-hidden">
                      {product.image_url ? (
                        <img 
                          src={product.image_url} 
                          alt={product.name}
                          className="w-full h-full object-cover"
                        />
                      ) : (
                        <div className="text-xs">📦</div>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="font-medium text-sm line-clamp-2">{product.name}</h4>
                      <p className="text-sm font-bold text-primary">${product.price}</p>
                      <Button size="sm" variant="outline" className="mt-1" asChild>
                        <Link to={`/products/${product.id}`}>View</Link>
                      </Button>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {/* Quick Stats */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-green-500" />
                Quick Stats
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Total Products</span>
                <span className="font-medium">{products.length}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">Categories</span>
                <span className="font-medium">{categories.length}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm text-muted-foreground">In Stock</span>
                <span className="font-medium">
                  {products.filter(p => p.stock > 0).length}
                </span>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Main Content */}
        <div className="lg:col-span-3">
          {/* Results Header */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold">
                {searchQuery ? `Search results for "${searchQuery}"` : 'All Products'}
              </h2>
              <p className="text-sm text-muted-foreground">
                {searchLoading ? 'Searching...' : `${products.length} products found`}
              </p>
            </div>
          </div>

          {/* Products Grid/List */}
          {loading ? (
            <div className={`grid gap-6 ${viewMode === 'grid' ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3' : 'grid-cols-1'}`}>
              {[...Array(6)].map((_, i) => (
                <Card key={i} className={viewMode === 'list' ? 'flex' : ''}>
                  <Skeleton className={`${viewMode === 'grid' ? 'aspect-square w-full' : 'w-48 h-48'}`} />
                  <div className={`p-4 ${viewMode === 'list' ? 'flex-1' : ''}`}>
                    <Skeleton className="h-4 w-3/4 mb-2" />
                    <Skeleton className="h-4 w-1/2 mb-2" />
                    <Skeleton className="h-4 w-full" />
                  </div>
                </Card>
              ))}
            </div>
          ) : products.length === 0 ? (
            <Card className="p-12 text-center">
              <div className="text-muted-foreground">
                <Search className="h-12 w-12 mx-auto mb-4" />
                <h3 className="text-lg font-medium mb-2">No products found</h3>
                <p>Try adjusting your search or filters</p>
              </div>
            </Card>
          ) : (
            <>
              <div className={`grid gap-6 ${viewMode === 'grid' ? 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3' : 'grid-cols-1'}`}>
                {products.map((product) => (
                  <ProductCard key={product.id} product={product} variant={viewMode} />
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex justify-center mt-8 gap-2">
                  <Button
                    variant="outline"
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                  >
                    Previous
                  </Button>
                  
                  {[...Array(Math.min(5, totalPages))].map((_, i) => {
                    const page = i + 1
                    return (
                      <Button
                        key={page}
                        variant={currentPage === page ? 'default' : 'outline'}
                        onClick={() => setCurrentPage(page)}
                      >
                        {page}
                      </Button>
                    )
                  })}
                  
                  <Button
                    variant="outline"
                    onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                    disabled={currentPage === totalPages}
                  >
                    Next
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}