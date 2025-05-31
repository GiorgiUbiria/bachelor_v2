import { useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router'
import { 
  ArrowLeft, 
  Package, 
  Truck, 
  CheckCircle, 
  XCircle, 
  Clock, 
  MapPin, 
  CreditCard, 
  Download, 
  MessageCircle,
  AlertTriangle
} from 'lucide-react'
import { useOrdersStore } from '../store/orders'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import { Progress } from '@/components/ui/progress'

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { currentOrder, isLoading, error, fetchOrderById, cancelOrder, clearCurrentOrder } = useOrdersStore()
  const { isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()

  useEffect(() => {
    if (!isAuthenticated) {
      navigate('/auth/login')
      return
    }

    if (id) {
      fetchOrderById(id)
    }

    return () => {
      clearCurrentOrder()
    }
  }, [id, isAuthenticated, fetchOrderById, clearCurrentOrder, navigate])

  const handleCancelOrder = async () => {
    if (!currentOrder) return

    try {
      await cancelOrder(currentOrder.id)
      addToast({
        type: 'success',
        description: 'Order cancelled successfully'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to cancel order'
      })
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending':
        return <Clock className="w-5 h-5 text-yellow-500" />
      case 'processing':
        return <AlertTriangle className="w-5 h-5 text-blue-500" />
      case 'shipped':
        return <Truck className="w-5 h-5 text-purple-500" />
      case 'delivered':
        return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'cancelled':
        return <XCircle className="w-5 h-5 text-red-500" />
      default:
        return <Package className="w-5 h-5 text-gray-500" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200'
      case 'processing':
        return 'bg-blue-100 text-blue-800 border-blue-200'
      case 'shipped':
        return 'bg-purple-100 text-purple-800 border-purple-200'
      case 'delivered':
        return 'bg-green-100 text-green-800 border-green-200'
      case 'cancelled':
        return 'bg-red-100 text-red-800 border-red-200'
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200'
    }
  }

  const getOrderProgress = (status: string) => {
    switch (status) {
      case 'pending':
        return 25
      case 'processing':
        return 50
      case 'shipped':
        return 75
      case 'delivered':
        return 100
      case 'cancelled':
        return 0
      default:
        return 0
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const calculateSubtotal = () => {
    if (!currentOrder) return 0
    return currentOrder.order_items.reduce((sum, item) => sum + (item.price * item.quantity), 0)
  }

  const calculateTax = () => {
    return calculateSubtotal() * 0.1 // 10% tax
  }

  const calculateShipping = () => {
    return calculateSubtotal() > 50 ? 0 : 9.99 // Free shipping over $50
  }

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            Please log in to view order details.
          </AlertDescription>
        </Alert>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="space-y-6">
          <Skeleton className="h-8 w-64" />
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 space-y-6">
              <Skeleton className="h-64" />
              <Skeleton className="h-48" />
            </div>
            <div className="space-y-6">
              <Skeleton className="h-32" />
              <Skeleton className="h-48" />
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error || !currentOrder) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            {error || 'Order not found'}
          </AlertDescription>
        </Alert>
        <Button asChild className="mt-4">
          <Link to="/orders">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Orders
          </Link>
        </Button>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="outline" asChild>
              <Link to="/orders">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Orders
              </Link>
            </Button>
            <div>
              <h1 className="text-3xl font-bold">Order #{currentOrder.id.slice(0, 8)}</h1>
              <p className="text-muted-foreground">
                Placed on {formatDate(currentOrder.created_at)}
              </p>
            </div>
          </div>
          
          <div className="flex items-center gap-2">
            <Badge className={`${getStatusColor(currentOrder.status)} border text-base px-3 py-1`}>
              <div className="flex items-center gap-2">
                {getStatusIcon(currentOrder.status)}
                {currentOrder.status.charAt(0).toUpperCase() + currentOrder.status.slice(1)}
              </div>
            </Badge>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Order Progress */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Truck className="w-5 h-5" />
                  Order Status
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <Progress value={getOrderProgress(currentOrder.status)} className="h-2" />
                  
                  <div className="grid grid-cols-4 gap-4 text-center">
                    <div className={`flex flex-col items-center gap-2 ${
                      ['pending', 'processing', 'shipped', 'delivered'].includes(currentOrder.status) 
                        ? 'text-primary' : 'text-muted-foreground'
                    }`}>
                      <Clock className="w-6 h-6" />
                      <span className="text-sm font-medium">Pending</span>
                    </div>
                    <div className={`flex flex-col items-center gap-2 ${
                      ['processing', 'shipped', 'delivered'].includes(currentOrder.status) 
                        ? 'text-primary' : 'text-muted-foreground'
                    }`}>
                      <AlertTriangle className="w-6 h-6" />
                      <span className="text-sm font-medium">Processing</span>
                    </div>
                    <div className={`flex flex-col items-center gap-2 ${
                      ['shipped', 'delivered'].includes(currentOrder.status) 
                        ? 'text-primary' : 'text-muted-foreground'
                    }`}>
                      <Truck className="w-6 h-6" />
                      <span className="text-sm font-medium">Shipped</span>
                    </div>
                    <div className={`flex flex-col items-center gap-2 ${
                      currentOrder.status === 'delivered' 
                        ? 'text-primary' : 'text-muted-foreground'
                    }`}>
                      <CheckCircle className="w-6 h-6" />
                      <span className="text-sm font-medium">Delivered</span>
                    </div>
                  </div>

                  {currentOrder.status === 'cancelled' && (
                    <div className="text-center p-4 bg-red-50 rounded-lg border border-red-200">
                      <XCircle className="w-8 h-8 text-red-500 mx-auto mb-2" />
                      <p className="text-red-800 font-medium">Order Cancelled</p>
                      <p className="text-red-600 text-sm">This order has been cancelled</p>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Order Items */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Package className="w-5 h-5" />
                  Order Items ({currentOrder.order_items.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {currentOrder.order_items.map((item) => (
                    <div key={item.id} className="flex gap-4 p-4 border rounded-lg">
                      <div className="w-16 h-16 bg-muted rounded-lg flex items-center justify-center overflow-hidden">
                        {item.product.image_url ? (
                          <img 
                            src={item.product.image_url} 
                            alt={item.product.name}
                            className="w-full h-full object-cover"
                          />
                        ) : (
                          <Package className="w-6 h-6 text-muted-foreground" />
                        )}
                      </div>
                      
                      <div className="flex-1">
                        <h4 className="font-medium mb-1">{item.product.name}</h4>
                        <p className="text-sm text-muted-foreground mb-2">
                          Category: {item.product.category}
                        </p>
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-4 text-sm">
                            <span>Quantity: {item.quantity}</span>
                            <span>Price: {formatCurrency(item.price)}</span>
                          </div>
                          <p className="font-medium">
                            {formatCurrency(item.price * item.quantity)}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Order Summary */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CreditCard className="w-5 h-5" />
                  Order Summary
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span>Subtotal</span>
                    <span>{formatCurrency(calculateSubtotal())}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Shipping</span>
                    <span className={calculateShipping() === 0 ? 'text-green-600' : ''}>
                      {calculateShipping() === 0 ? 'Free' : formatCurrency(calculateShipping())}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span>Tax</span>
                    <span>{formatCurrency(calculateTax())}</span>
                  </div>
                  <Separator />
                  <div className="flex justify-between font-medium text-lg">
                    <span>Total</span>
                    <span>{formatCurrency(currentOrder.total)}</span>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Payment & Shipping Info */}
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CreditCard className="w-5 h-5" />
                  Payment & Shipping
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <h4 className="font-medium mb-2 flex items-center gap-2">
                    <CreditCard className="w-4 h-4" />
                    Payment Method
                  </h4>
                  <p className="text-sm text-muted-foreground">
                    Credit Card ending in ****
                  </p>
                </div>
                
                <Separator />
                
                <div>
                  <h4 className="font-medium mb-2 flex items-center gap-2">
                    <MapPin className="w-4 h-4" />
                    Shipping Address
                  </h4>
                  <p className="text-sm text-muted-foreground">
                    Default shipping address
                  </p>
                </div>
              </CardContent>
            </Card>

            {/* Actions */}
            <Card>
              <CardHeader>
                <CardTitle>Actions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {(currentOrder.status === 'pending' || currentOrder.status === 'processing') && (
                  <Button 
                    variant="outline" 
                    onClick={handleCancelOrder}
                    disabled={isLoading}
                    className="w-full text-red-600 hover:text-red-700 hover:bg-red-50"
                  >
                    <XCircle className="w-4 h-4 mr-2" />
                    Cancel Order
                  </Button>
                )}
                
                <Button variant="outline" className="w-full">
                  <Download className="w-4 h-4 mr-2" />
                  Download Invoice
                </Button>
                
                <Button variant="outline" className="w-full">
                  <MessageCircle className="w-4 h-4 mr-2" />
                  Contact Support
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
} 