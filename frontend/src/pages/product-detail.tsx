import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router'
import { 
  ShoppingCart, Heart, Star, Share2, ArrowLeft, Package, 
  Truck, Shield, RotateCcw, MessageCircle, 
  Tag, Sparkles, TrendingUp, Zap, BarChart3 
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Progress } from '@/components/ui/progress'
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
  created_at: string
  updated_at: string
}

interface Comment {
  id: string
  content: string
  rating: number
  user_name: string
  created_at: string
}

interface SentimentData {
  overall_sentiment: string
  sentiment_score: number
  positive_percentage: number
  negative_percentage: number
  neutral_percentage: number
  insights: string[]
}

interface SmartDiscount {
  discount_type: string
  discount_value: number
  reason: string
  confidence: number
}

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { isAuthenticated } = useAuthStore()
  const { addItem } = useCartStore()
  const { addToast } = useUIStore()
  
  const [product, setProduct] = useState<Product | null>(null)
  const [comments, setComments] = useState<Comment[]>([])
  const [similarProducts, setSimilarProducts] = useState<Product[]>([])
  const [sentimentData, setSentimentData] = useState<SentimentData | null>(null)
  const [smartDiscount, setSmartDiscount] = useState<SmartDiscount | null>(null)
  const [suggestedTags, setSuggestedTags] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [quantity, setQuantity] = useState(1)
  const [isFavorite, setIsFavorite] = useState(false)

  useEffect(() => {
    if (!id) return
    
    const fetchProductData = async () => {
      try {
        setLoading(true)
        
        // Fetch product details
        const [productResponse, commentsResponse] = await Promise.all([
          apiService.products.getById(id),
          apiService.comments.getByProduct(id)
        ])
        
        setProduct(productResponse.data)
        setComments(commentsResponse.data.comments || [])
        
        // Fetch ML-powered features
        const mlPromises = []
        
        // Similar products
        mlPromises.push(
          apiService.mlService.recommendations.similar(id, 4)
            .then(response => setSimilarProducts(response.data.products || []))
            .catch(console.error)
        )
        
        // Sentiment analysis
        mlPromises.push(
          apiService.ml.sentiment.analyzeProduct(id)
            .then(response => setSentimentData(response.data))
            .catch(console.error)
        )
        
        // Smart discount suggestions
        mlPromises.push(
          apiService.ml.smartDiscounts.suggestForProduct(id)
            .then(response => setSmartDiscount(response.data.discount))
            .catch(console.error)
        )
        
        // Auto-tagging suggestions
        mlPromises.push(
          apiService.ml.autoTagging.suggest(id)
            .then(response => setSuggestedTags(response.data.suggested_tags || []))
            .catch(console.error)
        )
        
        await Promise.allSettled(mlPromises)
        
      } catch (error) {
        console.error('Error fetching product:', error)
        addToast({
          type: 'error',
          description: 'Failed to load product details'
        })
      } finally {
        setLoading(false)
      }
    }
    
    fetchProductData()
  }, [id, addToast])

  const handleAddToCart = async () => {
    if (!isAuthenticated) {
      addToast({
        type: 'error',
        description: 'Please login to add items to cart'
      })
      return
    }

    if (!product) return

    try {
      await addItem(product.id, quantity)
      addToast({
        type: 'success',
        description: `Added ${quantity} ${product.name} to cart`
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to add product to cart'
      })
    }
  }

  const handleToggleFavorite = async () => {
    if (!isAuthenticated || !product) return

    try {
      if (isFavorite) {
        await apiService.favorites.remove(product.id)
        setIsFavorite(false)
        addToast({
          type: 'success',
          description: 'Removed from favorites'
        })
      } else {
        await apiService.favorites.add(product.id)
        setIsFavorite(true)
        addToast({
          type: 'success',
          description: 'Added to favorites'
        })
      }
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to update favorites'
      })
    }
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <Skeleton className="aspect-square w-full" />
          <div className="space-y-4">
            <Skeleton className="h-8 w-3/4" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-10 w-32" />
          </div>
        </div>
      </div>
    )
  }

  if (!product) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card className="p-12 text-center">
          <Package className="h-12 w-12 mx-auto mb-4 text-muted-foreground" />
          <h3 className="text-lg font-medium mb-2">Product not found</h3>
          <p className="text-muted-foreground mb-4">The product you're looking for doesn't exist.</p>
          <Button asChild>
            <Link to="/products">Browse Products</Link>
          </Button>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 mb-6 text-sm text-muted-foreground">
        <Link to="/" className="hover:text-foreground">Home</Link>
        <span>/</span>
        <Link to="/products" className="hover:text-foreground">Products</Link>
        <span>/</span>
        <span className="text-foreground">{product.name}</span>
      </div>

      {/* Back Button */}
      <Button variant="ghost" className="mb-6" asChild>
        <Link to="/products">
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to Products
        </Link>
      </Button>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-12">
        {/* Product Image */}
        <div className="space-y-4">
          <div className="aspect-square bg-muted rounded-lg flex items-center justify-center overflow-hidden">
            {product.image_url ? (
              <img 
                src={product.image_url} 
                alt={product.name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="text-center text-muted-foreground">
                <Package className="h-24 w-24 mx-auto mb-4" />
                <p>No image available</p>
              </div>
            )}
          </div>
          
          {/* Smart Discount Alert */}
          {smartDiscount && (
            <Alert className="border-green-200 bg-green-50">
              <Zap className="h-4 w-4 text-green-600" />
              <AlertDescription className="text-green-800">
                <strong>AI-Suggested Discount:</strong> {smartDiscount.discount_value}% off
                <br />
                <span className="text-sm">{smartDiscount.reason}</span>
              </AlertDescription>
            </Alert>
          )}
        </div>

        {/* Product Info */}
        <div className="space-y-6">
          <div>
            <div className="flex items-start justify-between mb-2">
              <h1 className="text-3xl font-bold">{product.name}</h1>
              <Button
                variant="ghost"
                size="icon"
                onClick={handleToggleFavorite}
                className={isFavorite ? 'text-red-500' : ''}
              >
                <Heart className={`h-5 w-5 ${isFavorite ? 'fill-current' : ''}`} />
              </Button>
            </div>
            
            <div className="flex items-center gap-4 mb-4">
              <Badge variant="secondary">{product.category}</Badge>
              {product.rating && (
                <div className="flex items-center gap-1">
                  <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
                  <span className="font-medium">{product.rating}</span>
                  {product.reviews_count && (
                    <span className="text-muted-foreground">({product.reviews_count} reviews)</span>
                  )}
                </div>
              )}
            </div>

            <div className="flex items-center gap-2 mb-6">
              <span className="text-3xl font-bold text-primary">
                ${smartDiscount ? 
                  (product.price * (1 - smartDiscount.discount_value / 100)).toFixed(2) : 
                  product.price
                }
              </span>
              {smartDiscount && (
                <span className="text-lg text-muted-foreground line-through">
                  ${product.price}
                </span>
              )}
            </div>
          </div>

          <p className="text-muted-foreground leading-relaxed">
            {product.description}
          </p>

          {/* Tags */}
          {(product.tags || suggestedTags).length > 0 && (
            <div className="space-y-2">
              <h3 className="font-medium flex items-center gap-2">
                <Tag className="h-4 w-4" />
                Tags
                {suggestedTags.length > 0 && (
                  <Badge variant="outline" className="text-xs">
                    <Sparkles className="h-3 w-3 mr-1" />
                    AI-Suggested
                  </Badge>
                )}
              </h3>
              <div className="flex flex-wrap gap-2">
                {product.tags?.map((tag, index) => (
                  <Badge key={index} variant="outline">{tag}</Badge>
                ))}
                {suggestedTags.map((tag, index) => (
                  <Badge key={`suggested-${index}`} variant="secondary" className="border-dashed">
                    {tag}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {/* Stock Status */}
          <div className="flex items-center gap-2">
            <Package className="h-4 w-4" />
            <span className={`font-medium ${product.stock > 0 ? 'text-green-600' : 'text-red-600'}`}>
              {product.stock > 0 ? `${product.stock} in stock` : 'Out of stock'}
            </span>
          </div>

          {/* Quantity and Add to Cart */}
          <div className="space-y-4">
            <div className="flex items-center gap-4">
              <label className="font-medium">Quantity:</label>
              <div className="flex items-center border rounded-md">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  disabled={quantity <= 1}
                >
                  -
                </Button>
                <span className="px-4 py-2 min-w-[3rem] text-center">{quantity}</span>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setQuantity(Math.min(product.stock, quantity + 1))}
                  disabled={quantity >= product.stock}
                >
                  +
                </Button>
              </div>
            </div>

            <div className="flex gap-3">
              <Button 
                className="flex-1" 
                onClick={handleAddToCart}
                disabled={product.stock === 0}
              >
                <ShoppingCart className="mr-2 h-4 w-4" />
                Add to Cart
              </Button>
              <Button variant="outline" size="icon">
                <Share2 className="h-4 w-4" />
              </Button>
            </div>
          </div>

          {/* Features */}
          <div className="grid grid-cols-2 gap-4 pt-6 border-t">
            <div className="flex items-center gap-2 text-sm">
              <Truck className="h-4 w-4 text-muted-foreground" />
              <span>Free shipping</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Shield className="h-4 w-4 text-muted-foreground" />
              <span>Secure payment</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <RotateCcw className="h-4 w-4 text-muted-foreground" />
              <span>Easy returns</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <MessageCircle className="h-4 w-4 text-muted-foreground" />
              <span>24/7 support</span>
            </div>
          </div>
        </div>
      </div>

      {/* Tabs Section */}
      <Tabs defaultValue="details" className="mb-12">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="details">Details</TabsTrigger>
          <TabsTrigger value="reviews">Reviews ({comments.length})</TabsTrigger>
          <TabsTrigger value="sentiment">
            <BarChart3 className="h-4 w-4 mr-2" />
            AI Sentiment
          </TabsTrigger>
          <TabsTrigger value="similar">Similar Products</TabsTrigger>
        </TabsList>

        <TabsContent value="details" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Product Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="font-medium">Category:</span>
                  <p className="text-muted-foreground">{product.category}</p>
                </div>
                <div>
                  <span className="font-medium">Stock:</span>
                  <p className="text-muted-foreground">{product.stock} units</p>
                </div>
                <div>
                  <span className="font-medium">Added:</span>
                  <p className="text-muted-foreground">
                    {new Date(product.created_at).toLocaleDateString()}
                  </p>
                </div>
                <div>
                  <span className="font-medium">Updated:</span>
                  <p className="text-muted-foreground">
                    {new Date(product.updated_at).toLocaleDateString()}
                  </p>
                </div>
              </div>
              <Separator />
              <div>
                <span className="font-medium">Description:</span>
                <p className="text-muted-foreground mt-2 leading-relaxed">
                  {product.description}
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="reviews" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle>Customer Reviews</CardTitle>
              <CardDescription>
                See what other customers are saying about this product
              </CardDescription>
            </CardHeader>
            <CardContent>
              {comments.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <MessageCircle className="h-12 w-12 mx-auto mb-4" />
                  <p>No reviews yet. Be the first to review this product!</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {comments.map((comment) => (
                    <div key={comment.id} className="border-b pb-4 last:border-b-0">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{comment.user_name}</span>
                          <div className="flex items-center">
                            {[...Array(5)].map((_, i) => (
                              <Star
                                key={i}
                                className={`h-4 w-4 ${
                                  i < comment.rating
                                    ? 'fill-yellow-400 text-yellow-400'
                                    : 'text-muted-foreground'
                                }`}
                              />
                            ))}
                          </div>
                        </div>
                        <span className="text-sm text-muted-foreground">
                          {new Date(comment.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      <p className="text-muted-foreground">{comment.content}</p>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="sentiment" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <BarChart3 className="h-5 w-5" />
                AI Sentiment Analysis
              </CardTitle>
              <CardDescription>
                Machine learning analysis of customer feedback and reviews
              </CardDescription>
            </CardHeader>
            <CardContent>
              {sentimentData ? (
                <div className="space-y-6">
                  <div className="text-center">
                    <div className="text-3xl font-bold mb-2">
                      {sentimentData.overall_sentiment.toUpperCase()}
                    </div>
                    <div className="text-muted-foreground">
                      Overall Sentiment Score: {(sentimentData.sentiment_score * 100).toFixed(1)}%
                    </div>
                  </div>
                  
                  <div className="space-y-4">
                    <div>
                      <div className="flex justify-between mb-2">
                        <span className="text-sm font-medium">Positive</span>
                        <span className="text-sm text-muted-foreground">
                          {sentimentData.positive_percentage.toFixed(1)}%
                        </span>
                      </div>
                      <Progress value={sentimentData.positive_percentage} className="h-2" />
                    </div>
                    
                    <div>
                      <div className="flex justify-between mb-2">
                        <span className="text-sm font-medium">Neutral</span>
                        <span className="text-sm text-muted-foreground">
                          {sentimentData.neutral_percentage.toFixed(1)}%
                        </span>
                      </div>
                      <Progress value={sentimentData.neutral_percentage} className="h-2" />
                    </div>
                    
                    <div>
                      <div className="flex justify-between mb-2">
                        <span className="text-sm font-medium">Negative</span>
                        <span className="text-sm text-muted-foreground">
                          {sentimentData.negative_percentage.toFixed(1)}%
                        </span>
                      </div>
                      <Progress value={sentimentData.negative_percentage} className="h-2" />
                    </div>
                  </div>
                  
                  {sentimentData.insights.length > 0 && (
                    <div>
                      <h4 className="font-medium mb-3">AI Insights</h4>
                      <ul className="space-y-2">
                        {sentimentData.insights.map((insight, index) => (
                          <li key={index} className="flex items-start gap-2 text-sm">
                            <Sparkles className="h-4 w-4 text-primary mt-0.5 flex-shrink-0" />
                            <span>{insight}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <BarChart3 className="h-12 w-12 mx-auto mb-4" />
                  <p>Sentiment analysis not available for this product</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="similar" className="mt-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5" />
                Similar Products
              </CardTitle>
              <CardDescription>
                AI-recommended products based on similarity analysis
              </CardDescription>
            </CardHeader>
            <CardContent>
              {similarProducts.length === 0 ? (
                <div className="text-center py-8 text-muted-foreground">
                  <Package className="h-12 w-12 mx-auto mb-4" />
                  <p>No similar products found</p>
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  {similarProducts.map((similarProduct) => (
                    <Card key={similarProduct.id} className="hover:shadow-lg transition-shadow">
                      <div className="aspect-square bg-muted rounded-t-lg flex items-center justify-center overflow-hidden">
                        {similarProduct.image_url ? (
                          <img 
                            src={similarProduct.image_url} 
                            alt={similarProduct.name}
                            className="w-full h-full object-cover"
                          />
                        ) : (
                          <Package className="h-12 w-12 text-muted-foreground" />
                        )}
                      </div>
                      <CardContent className="p-4">
                        <h4 className="font-medium line-clamp-2 mb-2">{similarProduct.name}</h4>
                        <p className="text-lg font-bold text-primary mb-3">${similarProduct.price}</p>
                        <Button size="sm" className="w-full" asChild>
                          <Link to={`/products/${similarProduct.id}`}>View Product</Link>
                        </Button>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}