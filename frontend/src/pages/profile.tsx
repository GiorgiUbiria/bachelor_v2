import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { 
  User, 
  Package, 
  Heart, 
  ShoppingCart, 
  Star, 
  TrendingUp, 
  Edit, 
  Save, 
  X, 
  DollarSign,
  Activity
} from 'lucide-react'
import { useAuthStore } from '../store/auth'
import { useFavoritesStore } from '../store/favorites'
import { useOrdersStore } from '../store/orders'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import apiService from '../services/api'

interface UserProfileData {
  id: string
  email: string
  name: string
  phone?: string
  created_at: string
  updated_at: string
  comments: Comment[]
  favorites: any[]
  orders: any[]
  upvotes: any[]
  recommendations: any[]
  shopping_cart: any
  statistics: {
    average_order_value: number
    cart_items_count: number
    cart_total_value: number
    favorite_category: string
    recommendations_clicked: number
    recommendations_received: number
    total_interactions: number
    total_orders: number
    total_spent: number
  }
  recent_activity: {
    recent_interactions: any[]
    recent_orders: any[]
    recent_searches: any[]
    recent_views: any[]
  }
}

export default function ProfilePage() {
  const { user, isAuthenticated } = useAuthStore()
  const { favorites, fetchFavorites } = useFavoritesStore()
  const { orders, fetchOrders, fetchOrderStats } = useOrdersStore()
  const { addToast } = useUIStore()

  const [profileData, setProfileData] = useState<UserProfileData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isEditing, setIsEditing] = useState(false)
  const [editForm, setEditForm] = useState({
    name: '',
    phone: ''
  })
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isAuthenticated) {
      fetchProfileData()
      fetchFavorites()
      fetchOrders({ limit: 5 })
      fetchOrderStats()
    }
  }, [isAuthenticated])

  const fetchProfileData = async () => {
    try {
      setIsLoading(true)
      const response = await apiService.auth.profile()
      setProfileData(response.data.user)
      setEditForm({
        name: response.data.user.name || '',
        phone: response.data.user.phone || ''
      })
    } catch (error: any) {
      setError(error.response?.data?.error || 'Failed to load profile')
    } finally {
      setIsLoading(false)
    }
  }

  const handleSaveProfile = async () => {
    try {
      await apiService.auth.updateProfile(editForm)
      await fetchProfileData()
      setIsEditing(false)
      addToast({
        type: 'success',
        description: 'Profile updated successfully'
      })
    } catch (error: any) {
      addToast({
        type: 'error',
        description: error.response?.data?.error || 'Failed to update profile'
      })
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            Please log in to view your profile.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <Skeleton className="h-8 w-48" />
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {[...Array(6)].map((_, i) => (
              <Skeleton key={i} className="h-32" />
            ))}
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 bg-primary rounded-full flex items-center justify-center text-primary-foreground text-xl font-bold">
              {user?.name?.charAt(0).toUpperCase()}
            </div>
            <div>
              <h1 className="text-3xl font-bold">My Profile</h1>
              <p className="text-muted-foreground">
                Member since {formatDate(profileData?.created_at || '')}
              </p>
            </div>
          </div>
          <Button
            variant={isEditing ? "outline" : "default"}
            onClick={() => setIsEditing(!isEditing)}
          >
            {isEditing ? <X className="w-4 h-4 mr-2" /> : <Edit className="w-4 h-4 mr-2" />}
            {isEditing ? 'Cancel' : 'Edit Profile'}
          </Button>
        </div>

        {/* Profile Information */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <User className="w-5 h-5" />
              Profile Information
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isEditing ? (
              <div className="space-y-4">
                <div>
                  <Label htmlFor="name">Name</Label>
                  <Input
                    id="name"
                    value={editForm.name}
                    onChange={(e) => setEditForm(prev => ({ ...prev, name: e.target.value }))}
                  />
                </div>
                <div>
                  <Label htmlFor="phone">Phone</Label>
                  <Input
                    id="phone"
                    value={editForm.phone}
                    onChange={(e) => setEditForm(prev => ({ ...prev, phone: e.target.value }))}
                    placeholder="Enter your phone number"
                  />
                </div>
                <div>
                  <Label>Email</Label>
                  <Input value={profileData?.email} disabled />
                  <p className="text-sm text-muted-foreground mt-1">
                    Email cannot be changed
                  </p>
                </div>
                <Button onClick={handleSaveProfile} className="w-full">
                  <Save className="w-4 h-4 mr-2" />
                  Save Changes
                </Button>
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label>Name</Label>
                  <p className="text-lg">{profileData?.name}</p>
                </div>
                <div>
                  <Label>Email</Label>
                  <p className="text-lg">{profileData?.email}</p>
                </div>
                <div>
                  <Label>Phone</Label>
                  <p className="text-lg">{profileData?.phone || 'Not provided'}</p>
                </div>
                <div>
                  <Label>Member Since</Label>
                  <p className="text-lg">{formatDate(profileData?.created_at || '')}</p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        {/* Statistics */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <Package className="w-8 h-8 text-blue-500" />
                <div>
                  <p className="text-sm text-muted-foreground">Total Orders</p>
                  <p className="text-2xl font-bold">{profileData?.statistics?.total_orders || 0}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <DollarSign className="w-8 h-8 text-green-500" />
                <div>
                  <p className="text-sm text-muted-foreground">Total Spent</p>
                  <p className="text-2xl font-bold">
                    {formatCurrency(profileData?.statistics?.total_spent || 0)}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <Heart className="w-8 h-8 text-red-500" />
                <div>
                  <p className="text-sm text-muted-foreground">Favorites</p>
                  <p className="text-2xl font-bold">{favorites.length}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-6">
              <div className="flex items-center gap-4">
                <TrendingUp className="w-8 h-8 text-purple-500" />
                <div>
                  <p className="text-sm text-muted-foreground">Avg Order Value</p>
                  <p className="text-2xl font-bold">
                    {formatCurrency(profileData?.statistics?.average_order_value || 0)}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Detailed Information Tabs */}
        <Tabs defaultValue="orders" className="space-y-4">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="orders">Recent Orders</TabsTrigger>
            <TabsTrigger value="favorites">Favorites</TabsTrigger>
            <TabsTrigger value="activity">Activity</TabsTrigger>
            <TabsTrigger value="recommendations">Recommendations</TabsTrigger>
          </TabsList>

          <TabsContent value="orders" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Recent Orders</CardTitle>
              </CardHeader>
              <CardContent>
                {orders.length > 0 ? (
                  <div className="space-y-4">
                    {orders.slice(0, 5).map((order) => (
                      <div key={order.id} className="flex items-center justify-between p-4 border rounded-lg">
                        <div>
                          <p className="font-medium">Order #{order.id.slice(0, 8)}</p>
                          <p className="text-sm text-muted-foreground">
                            {formatDate(order.created_at)}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="font-medium">{formatCurrency(order.total)}</p>
                          <Badge variant={order.status === 'delivered' ? 'default' : 'secondary'}>
                            {order.status}
                          </Badge>
                        </div>
                      </div>
                    ))}
                    <Button asChild variant="outline" className="w-full">
                      <Link to="/orders">View All Orders</Link>
                    </Button>
                  </div>
                ) : (
                  <p className="text-center text-muted-foreground py-8">
                    No orders yet. <Link to="/products" className="text-primary hover:underline">Start shopping!</Link>
                  </p>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="favorites" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Favorite Products</CardTitle>
              </CardHeader>
              <CardContent>
                {favorites.length > 0 ? (
                  <div className="space-y-4">
                    {favorites.slice(0, 5).map((favorite) => (
                      <div key={favorite.id} className="flex items-center gap-4 p-4 border rounded-lg">
                        <div className="w-16 h-16 bg-muted rounded-lg flex items-center justify-center">
                          {favorite.product.image_url ? (
                            <img 
                              src={favorite.product.image_url} 
                              alt={favorite.product.name}
                              className="w-full h-full object-cover rounded-lg"
                            />
                          ) : (
                            <Package className="w-6 h-6 text-muted-foreground" />
                          )}
                        </div>
                        <div className="flex-1">
                          <p className="font-medium">{favorite.product.name}</p>
                          <p className="text-sm text-muted-foreground">{favorite.product.category}</p>
                        </div>
                        <p className="font-medium">{formatCurrency(favorite.product.price)}</p>
                      </div>
                    ))}
                    <Button asChild variant="outline" className="w-full">
                      <Link to="/favorites">View All Favorites</Link>
                    </Button>
                  </div>
                ) : (
                  <p className="text-center text-muted-foreground py-8">
                    No favorites yet. <Link to="/products" className="text-primary hover:underline">Browse products!</Link>
                  </p>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="activity" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Recent Activity</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center gap-4 p-4 border rounded-lg">
                    <Activity className="w-8 h-8 text-blue-500" />
                    <div>
                      <p className="font-medium">Total Interactions</p>
                      <p className="text-sm text-muted-foreground">
                        {profileData?.statistics?.total_interactions || 0} product interactions
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex items-center gap-4 p-4 border rounded-lg">
                    <Star className="w-8 h-8 text-yellow-500" />
                    <div>
                      <p className="font-medium">Favorite Category</p>
                      <p className="text-sm text-muted-foreground">
                        {profileData?.statistics?.favorite_category || 'No preference yet'}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-4 p-4 border rounded-lg">
                    <ShoppingCart className="w-8 h-8 text-green-500" />
                    <div>
                      <p className="font-medium">Current Cart</p>
                      <p className="text-sm text-muted-foreground">
                        {profileData?.statistics?.cart_items_count || 0} items • {formatCurrency(profileData?.statistics?.cart_total_value || 0)}
                      </p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="recommendations" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>ML Recommendations</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <p className="font-medium">Recommendations Received</p>
                      <p className="text-sm text-muted-foreground">
                        AI-powered product suggestions
                      </p>
                    </div>
                    <Badge variant="secondary">
                      {profileData?.statistics?.recommendations_received || 0}
                    </Badge>
                  </div>
                  
                  <div className="flex items-center justify-between p-4 border rounded-lg">
                    <div>
                      <p className="font-medium">Recommendations Clicked</p>
                      <p className="text-sm text-muted-foreground">
                        Engagement with AI suggestions
                      </p>
                    </div>
                    <Badge variant="secondary">
                      {profileData?.statistics?.recommendations_clicked || 0}
                    </Badge>
                  </div>

                  <Button asChild variant="outline" className="w-full">
                    <Link to="/products?tab=recommendations">View Recommendations</Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
} 