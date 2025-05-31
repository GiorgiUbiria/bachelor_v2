import { useEffect, useState } from 'react'
import { 
  BarChart3, Users, Package, ShoppingCart, TrendingUp, 
  Brain, Zap, Tag, MessageSquare, Settings, Plus,
  Search, Download, RefreshCw,
  Eye, Edit, Trash2, AlertTriangle
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import apiService from '../services/api'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'

interface DashboardStats {
  total_products: number
  total_users: number
  total_orders: number
  total_revenue: number
  active_users: number
  pending_orders: number
}

interface MLInsights {
  auto_tagging: {
    total_tags_suggested: number
    accuracy_rate: number
    recent_suggestions: Array<{
      product_name: string
      suggested_tags: string[]
      confidence: number
    }>
  }
  sentiment: {
    overall_sentiment: string
    positive_percentage: number
    negative_percentage: number
    neutral_percentage: number
    trending_topics: string[]
  }
  smart_discounts: {
    active_discounts: number
    total_savings: number
    conversion_rate: number
    recommendations: Array<{
      product_name: string
      suggested_discount: number
      reason: string
    }>
  }
}

interface Product {
  id: string
  name: string
  price: number
  stock: number
  category: string
  created_at: string
}

export default function AdminPage() {
  const { user, isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()
  
  const [dashboardStats, setDashboardStats] = useState<DashboardStats | null>(null)
  const [mlInsights, setMLInsights] = useState<MLInsights | null>(null)
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')

  const isAdmin = user?.email?.includes('ubiriagiorgi8') || user?.name?.toLowerCase().includes('admin')

  useEffect(() => {
    if (!isAuthenticated || !isAdmin) return

    const fetchAdminData = async () => {
      try {
        setLoading(true)
        
        const statsResponse = await apiService.analytics.dashboard()
        setDashboardStats(statsResponse.data)
        
        const [autoTaggingInsights, sentimentInsights, discountInsights] = await Promise.all([
          apiService.ml.autoTagging.getInsights().catch(() => ({ data: null })),
          apiService.ml.sentiment.getInsights().catch(() => ({ data: null })),
          apiService.ml.smartDiscounts.getInsights().catch(() => ({ data: null }))
        ])
        
        setMLInsights({
          auto_tagging: autoTaggingInsights.data || {
            total_tags_suggested: 0,
            accuracy_rate: 0,
            recent_suggestions: []
          },
          sentiment: sentimentInsights.data || {
            overall_sentiment: 'neutral',
            positive_percentage: 0,
            negative_percentage: 0,
            neutral_percentage: 100,
            trending_topics: []
          },
          smart_discounts: discountInsights.data || {
            active_discounts: 0,
            total_savings: 0,
            conversion_rate: 0,
            recommendations: []
          }
        })
        
        const [productsResponse] = await Promise.all([
          apiService.products.getAll({ limit: 50 }),
          Promise.resolve({ data: { users: [] } })
        ])
        
        setProducts(productsResponse.data.products || [])
      } catch (error) {
        console.error('Error fetching admin data:', error)
        addToast({
          type: 'error',
          description: 'Failed to load admin dashboard'
        })
      } finally {
        setLoading(false)
      }
    }
    
    fetchAdminData()
  }, [isAuthenticated, isAdmin, addToast])

  const handleInitializeMLServices = async () => {
    try {
      await apiService.ml.initializeServices()
      addToast({
        type: 'success',
        description: 'ML services initialized successfully'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to initialize ML services'
      })
    }
  }

  const handleRetrainModels = async () => {
    try {
      await Promise.all([
        apiService.mlService.recommendations.retrain(),
        apiService.mlService.trends.retrain()
      ])
      addToast({
        type: 'success',
        description: 'ML models retrained successfully'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to retrain ML models'
      })
    }
  }

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            Please log in to access the admin dashboard.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (!isAdmin) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertTriangle className="h-4 w-4" />
          <AlertDescription>
            You don't have permission to access the admin dashboard.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold">Admin Dashboard</h1>
          <p className="text-muted-foreground">
            Manage your ML-powered e-commerce platform
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleRetrainModels}>
            <RefreshCw className="mr-2 h-4 w-4" />
            Retrain Models
          </Button>
          <Button onClick={handleInitializeMLServices}>
            <Brain className="mr-2 h-4 w-4" />
            Initialize ML Services
          </Button>
        </div>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="ml-insights">ML Insights</TabsTrigger>
          <TabsTrigger value="products">Products</TabsTrigger>
          <TabsTrigger value="analytics">Analytics</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              {[...Array(4)].map((_, i) => (
                <Card key={i}>
                  <CardContent className="p-6">
                    <Skeleton className="h-4 w-20 mb-2" />
                    <Skeleton className="h-8 w-16" />
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : (
            <>
              {/* Stats Cards */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card>
                  <CardContent className="p-6">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Total Products</p>
                        <p className="text-2xl font-bold">{dashboardStats?.total_products || 0}</p>
                      </div>
                      <Package className="h-8 w-8 text-blue-500" />
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="p-6">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Total Users</p>
                        <p className="text-2xl font-bold">{dashboardStats?.total_users || 0}</p>
                      </div>
                      <Users className="h-8 w-8 text-green-500" />
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="p-6">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Total Orders</p>
                        <p className="text-2xl font-bold">{dashboardStats?.total_orders || 0}</p>
                      </div>
                      <ShoppingCart className="h-8 w-8 text-purple-500" />
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent className="p-6">
                    <div className="flex items-center justify-between">
                      <div>
                        <p className="text-sm font-medium text-muted-foreground">Total Revenue</p>
                        <p className="text-2xl font-bold">${dashboardStats?.total_revenue || 0}</p>
                      </div>
                      <TrendingUp className="h-8 w-8 text-orange-500" />
                    </div>
                  </CardContent>
                </Card>
              </div>

              {/* Quick Actions */}
              <Card>
                <CardHeader>
                  <CardTitle>Quick Actions</CardTitle>
                  <CardDescription>
                    Common administrative tasks and ML operations
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <Button className="h-20 flex-col gap-2">
                      <Plus className="h-6 w-6" />
                      Add Product
                    </Button>
                    <Button variant="outline" className="h-20 flex-col gap-2">
                      <Download className="h-6 w-6" />
                      Export Data
                    </Button>
                    <Button variant="outline" className="h-20 flex-col gap-2">
                      <BarChart3 className="h-6 w-6" />
                      View Reports
                    </Button>
                  </div>
                </CardContent>
              </Card>
            </>
          )}
        </TabsContent>

        {/* ML Insights Tab */}
        <TabsContent value="ml-insights" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Auto-Tagging Insights */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Tag className="h-5 w-5 text-purple-500" />
                  Auto-Tagging
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="text-center">
                  <div className="text-2xl font-bold">
                    {mlInsights?.auto_tagging.total_tags_suggested || 0}
                  </div>
                  <div className="text-sm text-muted-foreground">Tags Suggested</div>
                </div>
                <div>
                  <div className="flex justify-between mb-2">
                    <span className="text-sm">Accuracy Rate</span>
                    <span className="text-sm font-medium">
                      {(mlInsights?.auto_tagging.accuracy_rate || 0).toFixed(1)}%
                    </span>
                  </div>
                  <Progress value={mlInsights?.auto_tagging.accuracy_rate || 0} />
                </div>
                {mlInsights?.auto_tagging?.recent_suggestions && mlInsights.auto_tagging.recent_suggestions.length > 0 && (
                  <div>
                    <h4 className="font-medium mb-2">Recent Suggestions</h4>
                    <div className="space-y-2">
                      {mlInsights.auto_tagging.recent_suggestions.slice(0, 3).map((suggestion, index) => (
                        <div key={index} className="text-xs">
                          <div className="font-medium">{suggestion.product_name}</div>
                          <div className="text-muted-foreground">
                            {suggestion.suggested_tags.join(', ')}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Sentiment Analysis */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <MessageSquare className="h-5 w-5 text-blue-500" />
                  Sentiment Analysis
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="text-center">
                  <div className="text-2xl font-bold capitalize">
                    {mlInsights?.sentiment.overall_sentiment || 'Neutral'}
                  </div>
                  <div className="text-sm text-muted-foreground">Overall Sentiment</div>
                </div>
                <div className="space-y-3">
                  <div>
                    <div className="flex justify-between mb-1">
                      <span className="text-xs">Positive</span>
                      <span className="text-xs">{mlInsights?.sentiment.positive_percentage.toFixed(1)}%</span>
                    </div>
                    <Progress value={mlInsights?.sentiment.positive_percentage || 0} className="h-2" />
                  </div>
                  <div>
                    <div className="flex justify-between mb-1">
                      <span className="text-xs">Neutral</span>
                      <span className="text-xs">{mlInsights?.sentiment.neutral_percentage.toFixed(1)}%</span>
                    </div>
                    <Progress value={mlInsights?.sentiment.neutral_percentage || 0} className="h-2" />
                  </div>
                  <div>
                    <div className="flex justify-between mb-1">
                      <span className="text-xs">Negative</span>
                      <span className="text-xs">{mlInsights?.sentiment.negative_percentage.toFixed(1)}%</span>
                    </div>
                    <Progress value={mlInsights?.sentiment.negative_percentage || 0} className="h-2" />
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Smart Discounts */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Zap className="h-5 w-5 text-yellow-500" />
                  Smart Discounts
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4 text-center">
                  <div>
                    <div className="text-xl font-bold">
                      {mlInsights?.smart_discounts.active_discounts || 0}
                    </div>
                    <div className="text-xs text-muted-foreground">Active</div>
                  </div>
                  <div>
                    <div className="text-xl font-bold">
                      ${mlInsights?.smart_discounts.total_savings || 0}
                    </div>
                    <div className="text-xs text-muted-foreground">Savings</div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-2">
                    <span className="text-sm">Conversion Rate</span>
                    <span className="text-sm font-medium">
                      {(mlInsights?.smart_discounts.conversion_rate || 0).toFixed(1)}%
                    </span>
                  </div>
                  <Progress value={mlInsights?.smart_discounts.conversion_rate || 0} />
                </div>
                {mlInsights?.smart_discounts?.recommendations && mlInsights.smart_discounts.recommendations.length > 0 && (
                  <div>
                    <h4 className="font-medium mb-2">Recommendations</h4>
                    <div className="space-y-2">
                      {mlInsights.smart_discounts.recommendations.slice(0, 2).map((rec, index) => (
                        <div key={index} className="text-xs">
                          <div className="font-medium">{rec.product_name}</div>
                          <div className="text-muted-foreground">
                            {rec.suggested_discount}% off - {rec.reason}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* ML Services Status */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Brain className="h-5 w-5" />
                ML Services Status
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <div className="font-medium">Recommendation Engine</div>
                    <div className="text-sm text-muted-foreground">Status: Active</div>
                  </div>
                  <Badge variant="default">Online</Badge>
                </div>
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <div className="font-medium">Search Service</div>
                    <div className="text-sm text-muted-foreground">Status: Active</div>
                  </div>
                  <Badge variant="default">Online</Badge>
                </div>
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div>
                    <div className="font-medium">Trend Analysis</div>
                    <div className="text-sm text-muted-foreground">Status: Active</div>
                  </div>
                  <Badge variant="default">Online</Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Products Tab */}
        <TabsContent value="products" className="space-y-6">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Product Management</CardTitle>
                  <CardDescription>Manage your product catalog</CardDescription>
                </div>
                <Button>
                  <Plus className="mr-2 h-4 w-4" />
                  Add Product
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {/* Search and Filters */}
              <div className="flex gap-4 mb-6">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input placeholder="Search products..." className="pl-10" />
                </div>
                <Select>
                  <SelectTrigger className="w-48">
                    <SelectValue placeholder="Category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    <SelectItem value="electronics">Electronics</SelectItem>
                    <SelectItem value="clothing">Clothing</SelectItem>
                    <SelectItem value="books">Books</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Products Table */}
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Category</TableHead>
                    <TableHead>Price</TableHead>
                    <TableHead>Stock</TableHead>
                    <TableHead>Created</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {products.slice(0, 10).map((product) => (
                    <TableRow key={product.id}>
                      <TableCell className="font-medium">{product.name}</TableCell>
                      <TableCell>
                        <Badge variant="secondary">{product.category}</Badge>
                      </TableCell>
                      <TableCell>${product.price}</TableCell>
                      <TableCell>{product.stock}</TableCell>
                      <TableCell>
                        {new Date(product.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell>
                        <div className="flex gap-2">
                          <Button size="sm" variant="outline">
                            <Eye className="h-4 w-4" />
                          </Button>
                          <Button size="sm" variant="outline">
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button size="sm" variant="outline">
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Analytics Tab */}
        <TabsContent value="analytics" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle>Sales Analytics</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center py-8 text-muted-foreground">
                  <BarChart3 className="h-12 w-12 mx-auto mb-4" />
                  <p>Sales chart would be displayed here</p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>User Activity</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-center py-8 text-muted-foreground">
                  <Users className="h-12 w-12 mx-auto mb-4" />
                  <p>User activity chart would be displayed here</p>
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                ML Configuration
              </CardTitle>
              <CardDescription>
                Configure machine learning services and parameters
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h4 className="font-medium mb-3">Recommendation Settings</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Algorithm</span>
                      <Select defaultValue="collaborative">
                        <SelectTrigger className="w-32">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="collaborative">Collaborative</SelectItem>
                          <SelectItem value="content">Content-based</SelectItem>
                          <SelectItem value="hybrid">Hybrid</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Max Recommendations</span>
                      <Input type="number" defaultValue="10" className="w-20" />
                    </div>
                  </div>
                </div>

                <div>
                  <h4 className="font-medium mb-3">Auto-Tagging Settings</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Confidence Threshold</span>
                      <Input type="number" defaultValue="0.8" step="0.1" className="w-20" />
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Max Tags per Product</span>
                      <Input type="number" defaultValue="5" className="w-20" />
                    </div>
                  </div>
                </div>
              </div>

              <div className="pt-4 border-t">
                <Button>Save Configuration</Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
} 
