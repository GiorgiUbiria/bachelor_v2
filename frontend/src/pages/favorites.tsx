import { useState, useEffect } from 'react'
import { Link } from 'react-router'
import { 
  Heart, 
  Grid3X3, 
  List, 
  Search, 
  Filter, 
  ShoppingCart,
  Package,
  Star,
  X,
  SortAsc,
  SortDesc
} from 'lucide-react'
import { useFavoritesStore } from '../store/favorites'
import { useCartStore } from '../store/cart'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'

export default function FavoritesPage() {
  const { favorites, isLoading, error, fetchFavorites, removeFavorite } = useFavoritesStore()
  const { addItem: addToCart, isLoading: cartLoading } = useCartStore()
  const { isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()

  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [categoryFilter, setCategoryFilter] = useState<string>('all')
  const [sortBy, setSortBy] = useState<string>('newest')
  const [priceRange, setPriceRange] = useState<{ min: number; max: number }>({ min: 0, max: 1000 })

  useEffect(() => {
    if (isAuthenticated) {
      fetchFavorites()
    }
  }, [isAuthenticated, fetchFavorites])

  const handleRemoveFavorite = async (productId: string) => {
    try {
      await removeFavorite(productId)
      addToast({
        type: 'success',
        description: 'Removed from favorites'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to remove from favorites'
      })
    }
  }

  const handleAddToCart = async (productId: string) => {
    try {
      await addToCart(productId, 1)
      addToast({
        type: 'success',
        description: 'Added to cart'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to add to cart'
      })
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  // Get unique categories from favorites
  const categories = Array.from(new Set(favorites.map(fav => fav.product.category)))

  // Filter and sort favorites
  const filteredFavorites = favorites.filter(favorite => {
    const matchesSearch = favorite.product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         favorite.product.description.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesCategory = categoryFilter === 'all' || favorite.product.category === categoryFilter
    const matchesPrice = favorite.product.price >= priceRange.min && favorite.product.price <= priceRange.max
    
    return matchesSearch && matchesCategory && matchesPrice
  })

  const sortedFavorites = [...filteredFavorites].sort((a, b) => {
    switch (sortBy) {
      case 'newest':
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      case 'oldest':
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
      case 'name_asc':
        return a.product.name.localeCompare(b.product.name)
      case 'name_desc':
        return b.product.name.localeCompare(a.product.name)
      case 'price_asc':
        return a.product.price - b.product.price
      case 'price_desc':
        return b.product.price - a.product.price
      default:
        return 0
    }
  })

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            Please log in to view your favorites.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (isLoading && favorites.length === 0) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <Skeleton className="h-8 w-48" />
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {[...Array(8)].map((_, i) => (
              <Skeleton key={i} className="h-80" />
            ))}
          </div>
        </div>
      </div>
    )
  }

  const ProductCard = ({ favorite, variant = 'grid' }: { favorite: any; variant?: 'grid' | 'list' }) => (
    <Card className={`group hover:shadow-lg transition-shadow ${variant === 'list' ? 'flex flex-row' : ''}`}>
      <div className={`relative overflow-hidden ${variant === 'list' ? 'w-48 flex-shrink-0' : 'aspect-square'}`}>
        <Link to={`/products/${favorite.product.id}`}>
          {favorite.product.image_url ? (
            <img 
              src={favorite.product.image_url} 
              alt={favorite.product.name}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            />
          ) : (
            <div className="w-full h-full bg-muted flex items-center justify-center">
              <Package className="w-12 h-12 text-muted-foreground" />
            </div>
          )}
        </Link>
        
        {/* Favorite Button */}
        <Button
          size="icon"
          variant="ghost"
          className="absolute top-2 right-2 bg-white/80 hover:bg-white"
          onClick={() => handleRemoveFavorite(favorite.product.id)}
        >
          <Heart className="w-4 h-4 fill-red-500 text-red-500" />
        </Button>

        {/* Category Badge */}
        <Badge 
          variant="secondary" 
          className="absolute top-2 left-2 bg-white/90"
        >
          {favorite.product.category}
        </Badge>
      </div>

      <CardContent className={`p-4 ${variant === 'list' ? 'flex-1 flex flex-col justify-between' : ''}`}>
        <div className="space-y-2">
          <Link to={`/products/${favorite.product.id}`}>
            <h3 className="font-medium line-clamp-2 hover:text-primary transition-colors">
              {favorite.product.name}
            </h3>
          </Link>
          
          {variant === 'list' && (
            <p className="text-sm text-muted-foreground line-clamp-2">
              {favorite.product.description}
            </p>
          )}
          
          <div className="flex items-center justify-between">
            <span className="text-lg font-bold text-primary">
              {formatCurrency(favorite.product.price)}
            </span>
            {variant === 'grid' && (
              <div className="flex items-center gap-1">
                <Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
                <span className="text-sm text-muted-foreground">4.5</span>
              </div>
            )}
          </div>

          <div className="text-xs text-muted-foreground">
            Added {formatDate(favorite.created_at)}
          </div>
        </div>

        <div className={`flex gap-2 ${variant === 'list' ? 'mt-4' : 'mt-3'}`}>
          <Button 
            size="sm" 
            className="flex-1"
            onClick={() => handleAddToCart(favorite.product.id)}
            disabled={cartLoading}
          >
            <ShoppingCart className="w-4 h-4 mr-2" />
            Add to Cart
          </Button>
          <Button 
            size="sm" 
            variant="outline"
            onClick={() => handleRemoveFavorite(favorite.product.id)}
          >
            <X className="w-4 h-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  )

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold flex items-center gap-2">
              <Heart className="w-8 h-8 text-red-500" />
              My Favorites
            </h1>
            <p className="text-muted-foreground">
              {favorites.length} {favorites.length === 1 ? 'item' : 'items'} in your favorites
            </p>
          </div>

          {/* View Mode Toggle */}
          <div className="flex items-center gap-2">
            <Button
              variant={viewMode === 'grid' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setViewMode('grid')}
            >
              <Grid3X3 className="w-4 h-4" />
            </Button>
            <Button
              variant={viewMode === 'list' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setViewMode('list')}
            >
              <List className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Filters */}
        <Card>
          <CardContent className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {/* Search */}
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search favorites..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>

              {/* Category Filter */}
              <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                <SelectTrigger>
                  <Filter className="w-4 h-4 mr-2" />
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

              {/* Sort */}
              <Select value={sortBy} onValueChange={setSortBy}>
                <SelectTrigger>
                  <SelectValue placeholder="Sort by" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="newest">Newest First</SelectItem>
                  <SelectItem value="oldest">Oldest First</SelectItem>
                  <SelectItem value="name_asc">Name: A to Z</SelectItem>
                  <SelectItem value="name_desc">Name: Z to A</SelectItem>
                  <SelectItem value="price_asc">Price: Low to High</SelectItem>
                  <SelectItem value="price_desc">Price: High to Low</SelectItem>
                </SelectContent>
              </Select>

              {/* Price Range */}
              <div className="flex gap-2">
                <Input
                  type="number"
                  placeholder="Min $"
                  value={priceRange.min || ''}
                  onChange={(e) => setPriceRange(prev => ({ ...prev, min: Number(e.target.value) || 0 }))}
                  className="w-20"
                />
                <Input
                  type="number"
                  placeholder="Max $"
                  value={priceRange.max || ''}
                  onChange={(e) => setPriceRange(prev => ({ ...prev, max: Number(e.target.value) || 1000 }))}
                  className="w-20"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Error State */}
        {error && (
          <Alert>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Empty State */}
        {sortedFavorites.length === 0 ? (
          <Card>
            <CardContent className="p-12 text-center">
              <Heart className="w-16 h-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-medium mb-2">
                {searchQuery || categoryFilter !== 'all' 
                  ? 'No favorites match your filters'
                  : 'No favorites yet'
                }
              </h3>
              <p className="text-muted-foreground mb-4">
                {searchQuery || categoryFilter !== 'all'
                  ? 'Try adjusting your search or filters'
                  : 'Start adding products to your favorites to see them here'
                }
              </p>
              {!searchQuery && categoryFilter === 'all' && (
                <Button asChild>
                  <Link to="/products">Browse Products</Link>
                </Button>
              )}
            </CardContent>
          </Card>
        ) : (
          <>
            {/* Results Count */}
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">
                Showing {sortedFavorites.length} of {favorites.length} favorites
              </p>
              
              {(searchQuery || categoryFilter !== 'all') && (
                <Button 
                  variant="outline" 
                  size="sm"
                  onClick={() => {
                    setSearchQuery('')
                    setCategoryFilter('all')
                    setPriceRange({ min: 0, max: 1000 })
                  }}
                >
                  Clear Filters
                </Button>
              )}
            </div>

            {/* Products Grid/List */}
            <div className={
              viewMode === 'grid' 
                ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6'
                : 'space-y-4'
            }>
              {sortedFavorites.map((favorite) => (
                <ProductCard 
                  key={favorite.id} 
                  favorite={favorite} 
                  variant={viewMode}
                />
              ))}
            </div>
          </>
        )}

        {/* Quick Actions */}
        {sortedFavorites.length > 0 && (
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium mb-1">Quick Actions</h3>
                  <p className="text-sm text-muted-foreground">
                    Manage all your favorites at once
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button 
                    variant="outline"
                    onClick={async () => {
                      try {
                        await Promise.all(
                          sortedFavorites.map(fav => addToCart(fav.product.id, 1))
                        )
                        addToast({
                          type: 'success',
                          description: `Added ${sortedFavorites.length} items to cart`
                        })
                      } catch (error) {
                        addToast({
                          type: 'error',
                          description: 'Failed to add some items to cart'
                        })
                      }
                    }}
                    disabled={cartLoading}
                  >
                    <ShoppingCart className="w-4 h-4 mr-2" />
                    Add All to Cart
                  </Button>
                  <Button asChild>
                    <Link to="/products">
                      <Heart className="w-4 h-4 mr-2" />
                      Find More
                    </Link>
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
} 