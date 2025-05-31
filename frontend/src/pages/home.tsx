import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { TrendingUp, ShoppingBag, Brain, BarChart3, Search, Tag, Heart, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import apiService from '../services/api'
import { useAuthStore } from '../store/auth'

interface Product {
  id: string
  name: string
  price: number
  image_url?: string
  category: string
  description: string
}

interface TrendData {
  popular_products?: Product[]
  trending_categories?: string[]
  search_trends?: { query: string; count: number }[]
}

export default function HomePage() {
  const { isAuthenticated } = useAuthStore()
  const [trendData, setTrendData] = useState<TrendData>({})
  const [recommendations, setRecommendations] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [analytics, setAnalytics] = useState<any>(null)

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch trending data from ML service
        const [analyticsResponse] = await Promise.all([
          apiService.analytics.dashboard()
        ])
        
        setAnalytics(analyticsResponse.data)
        
        // Set some mock trend data for now since the ML service endpoint doesn't exist
        setTrendData({
          popular_products: [],
          trending_categories: [],
          search_trends: []
        })

        // Fetch recommendations if authenticated
        if (isAuthenticated) {
          try {
            const recsResponse = await apiService.products.getRecommendations({ limit: 6 })
            setRecommendations(recsResponse.data.products || [])
          } catch (error) {
            console.log('Recommendations not available')
          }
        }
      } catch (error) {
        console.error('Error fetching data:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchData()
  }, [isAuthenticated])

  const features = [
    {
      icon: Brain,
      title: 'AI-Powered Recommendations',
      description: 'Machine learning algorithms analyze your behavior to suggest products you\'ll love',
      color: 'text-blue-500'
    },
    {
      icon: Search,
      title: 'Smart Search',
      description: 'Advanced search with auto-suggestions and semantic understanding',
      color: 'text-green-500'
    },
    {
      icon: Tag,
      title: 'Auto-Tagging',
      description: 'Products are automatically categorized and tagged using ML',
      color: 'text-purple-500'
    },
    {
      icon: BarChart3,
      title: 'Sentiment Analysis',
      description: 'Real-time analysis of customer feedback and product sentiment',
      color: 'text-orange-500'
    },
    {
      icon: Zap,
      title: 'Smart Discounts',
      description: 'Dynamic pricing and discount suggestions based on market trends',
      color: 'text-red-500'
    },
    {
      icon: TrendingUp,
      title: 'Trend Forecasting',
      description: 'Predict future demand and identify emerging product trends',
      color: 'text-indigo-500'
    }
  ]

  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="bg-gradient-to-br from-primary/10 via-background to-secondary/10 py-20">
        <div className="container mx-auto px-4">
          <div className="text-center max-w-4xl mx-auto">
            <h1 className="text-4xl md:text-6xl font-bold mb-6 bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
              ML-Powered E-Commerce Platform
            </h1>
            <p className="text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
              Experience the future of online shopping with our AI-driven platform that learns, adapts, and personalizes your shopping journey.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button size="lg" asChild>
                <Link to="/products">
                  <ShoppingBag className="mr-2 h-5 w-5" />
                  Explore Products
                </Link>
              </Button>
              {!isAuthenticated && (
                <Button variant="outline" size="lg" asChild>
                  <Link to="/auth/register">Get Started</Link>
                </Button>
              )}
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-background">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold mb-4">Powered by Machine Learning</h2>
            <p className="text-muted-foreground max-w-2xl mx-auto">
              Our platform leverages cutting-edge AI technologies to provide intelligent shopping experiences
            </p>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <Card key={index} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <feature.icon className={`h-12 w-12 ${feature.color} mb-4`} />
                  <CardTitle>{feature.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription>{feature.description}</CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Analytics Dashboard Preview */}
      {analytics && (
        <section className="py-20 bg-muted/30">
          <div className="container mx-auto px-4">
            <div className="text-center mb-16">
              <h2 className="text-3xl font-bold mb-4">Real-Time Analytics</h2>
              <p className="text-muted-foreground">Live insights from our ML-powered analytics engine</p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">Total Products</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{analytics.total_products || 0}</div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">Active Users</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{analytics.active_users || 0}</div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">Total Orders</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{analytics.total_orders || 0}</div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">Revenue</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">${analytics.total_revenue || 0}</div>
                </CardContent>
              </Card>
            </div>
          </div>
        </section>
      )}

      {/* Trending Products */}
      <section className="py-20">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-between mb-12">
            <div>
              <h2 className="text-3xl font-bold mb-4">Trending Now</h2>
              <p className="text-muted-foreground">Discover what's popular based on ML analysis</p>
            </div>
            <Button variant="outline" asChild>
              <Link to="/products">View All</Link>
            </Button>
          </div>
          
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[...Array(6)].map((_, i) => (
                <Card key={i}>
                  <Skeleton className="h-48 w-full" />
                  <CardContent className="p-4">
                    <Skeleton className="h-4 w-3/4 mb-2" />
                    <Skeleton className="h-4 w-1/2" />
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {(trendData.popular_products || []).slice(0, 6).map((product) => (
                <Card key={product.id} className="hover:shadow-lg transition-shadow">
                  <div className="aspect-square bg-muted rounded-t-lg flex items-center justify-center">
                    {product.image_url ? (
                      <img 
                        src={product.image_url} 
                        alt={product.name}
                        className="w-full h-full object-cover rounded-t-lg"
                      />
                    ) : (
                      <ShoppingBag className="h-16 w-16 text-muted-foreground" />
                    )}
                  </div>
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between mb-2">
                      <h3 className="font-semibold line-clamp-1">{product.name}</h3>
                      <Badge variant="secondary">{product.category}</Badge>
                    </div>
                    <p className="text-muted-foreground text-sm mb-3 line-clamp-2">
                      {product.description}
                    </p>
                    <div className="flex items-center justify-between">
                      <span className="text-lg font-bold">${product.price}</span>
                      <Button size="sm" asChild>
                        <Link to={`/products/${product.id}`}>View</Link>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </section>

      {/* Personalized Recommendations */}
      {isAuthenticated && recommendations.length > 0 && (
        <section className="py-20 bg-muted/30">
          <div className="container mx-auto px-4">
            <div className="text-center mb-12">
              <h2 className="text-3xl font-bold mb-4">
                <Heart className="inline-block mr-2 h-8 w-8 text-red-500" />
                Just For You
              </h2>
              <p className="text-muted-foreground">Personalized recommendations based on your preferences</p>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {recommendations.map((product) => (
                <Card key={product.id} className="hover:shadow-lg transition-shadow">
                  <div className="aspect-square bg-muted rounded-t-lg flex items-center justify-center">
                    {product.image_url ? (
                      <img 
                        src={product.image_url} 
                        alt={product.name}
                        className="w-full h-full object-cover rounded-t-lg"
                      />
                    ) : (
                      <ShoppingBag className="h-16 w-16 text-muted-foreground" />
                    )}
                  </div>
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between mb-2">
                      <h3 className="font-semibold line-clamp-1">{product.name}</h3>
                      <Badge variant="outline">Recommended</Badge>
                    </div>
                    <p className="text-muted-foreground text-sm mb-3 line-clamp-2">
                      {product.description}
                    </p>
                    <div className="flex items-center justify-between">
                      <span className="text-lg font-bold">${product.price}</span>
                      <Button size="sm" asChild>
                        <Link to={`/products/${product.id}`}>View</Link>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Search Trends */}
      {trendData.search_trends && trendData.search_trends.length > 0 && (
        <section className="py-20">
          <div className="container mx-auto px-4">
            <div className="text-center mb-12">
              <h2 className="text-3xl font-bold mb-4">What People Are Searching</h2>
              <p className="text-muted-foreground">Real-time search trends powered by ML analytics</p>
            </div>
            
            <div className="flex flex-wrap gap-3 justify-center">
              {trendData.search_trends.slice(0, 10).map((trend, index) => (
                <Badge key={index} variant="outline" className="text-sm py-2 px-4">
                  {trend.query} ({trend.count})
                </Badge>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* CTA Section */}
      <section className="py-20 bg-primary text-primary-foreground">
        <div className="container mx-auto px-4 text-center">
          <h2 className="text-3xl font-bold mb-4">Ready to Experience the Future?</h2>
          <p className="text-xl mb-8 opacity-90 max-w-2xl mx-auto">
            Join thousands of users who are already enjoying personalized shopping with AI
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            {!isAuthenticated ? (
              <>
                <Button size="lg" variant="secondary" asChild>
                  <Link to="/auth/register">Sign Up Now</Link>
                </Button>
                <Button size="lg" variant="outline" asChild>
                  <Link to="/auth/login">Sign In</Link>
                </Button>
              </>
            ) : (
              <Button size="lg" variant="secondary" asChild>
                <Link to="/products">Start Shopping</Link>
              </Button>
            )}
          </div>
        </div>
      </section>
    </div>
  )
}
