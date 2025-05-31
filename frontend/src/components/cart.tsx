import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import { ShoppingCart, Plus, Minus, Trash2, Package, CreditCard, Truck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { useCartStore } from '../store/cart'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'

interface CartProps {
  trigger?: React.ReactNode
  showTrigger?: boolean
}

export function Cart({ trigger, showTrigger = true }: CartProps) {
  const { cart, isLoading, error, fetchCart, updateItem, removeItem, clearCart } = useCartStore()
  const { isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    if (isAuthenticated && isOpen) {
      fetchCart()
    }
  }, [isAuthenticated, isOpen, fetchCart])

  const handleUpdateQuantity = async (itemId: string, newQuantity: number) => {
    if (newQuantity <= 0) {
      handleRemoveItem(itemId)
      return
    }

    try {
      await updateItem(itemId, newQuantity)
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to update item quantity'
      })
    }
  }

  const handleRemoveItem = async (itemId: string) => {
    try {
      await removeItem(itemId)
      addToast({
        type: 'success',
        description: 'Item removed from cart'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to remove item'
      })
    }
  }

  const handleClearCart = async () => {
    try {
      await clearCart()
      addToast({
        type: 'success',
        description: 'Cart cleared'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to clear cart'
      })
    }
  }

  const cartItemsCount = cart?.total_items || 0
  const cartTotal = cart?.total_amount || 0

  const CartTrigger = trigger || (
    <Button variant="outline" size="icon" className="relative">
      <ShoppingCart className="h-4 w-4" />
      {cartItemsCount > 0 && (
        <Badge 
          variant="destructive" 
          className="absolute -top-2 -right-2 h-5 w-5 flex items-center justify-center p-0 text-xs"
        >
          {cartItemsCount > 99 ? '99+' : cartItemsCount}
        </Badge>
      )}
    </Button>
  )

  const CartContent = () => {
    if (!isAuthenticated) {
      return (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <ShoppingCart className="h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium mb-2">Sign in to view your cart</h3>
          <p className="text-muted-foreground mb-4">
            Create an account or sign in to start shopping
          </p>
          <div className="flex gap-2">
            <Button asChild>
              <Link to="/auth/login" onClick={() => setIsOpen(false)}>
                Sign In
              </Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="/auth/register" onClick={() => setIsOpen(false)}>
                Sign Up
              </Link>
            </Button>
          </div>
        </div>
      )
    }

    if (isLoading) {
      return (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="flex gap-4">
              <Skeleton className="h-16 w-16 rounded" />
              <div className="flex-1 space-y-2">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
              </div>
            </div>
          ))}
        </div>
      )
    }

    if (error) {
      return (
        <Alert>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )
    }

    if (!cart || cart.items.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Package className="h-12 w-12 text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium mb-2">Your cart is empty</h3>
          <p className="text-muted-foreground mb-4">
            Add some products to get started
          </p>
          <Button asChild>
            <Link to="/products" onClick={() => setIsOpen(false)}>
              Browse Products
            </Link>
          </Button>
        </div>
      )
    }

    return (
      <div className="flex flex-col h-full">
        {/* Cart Items */}
        <div className="flex-1 overflow-y-auto">
          <div className="space-y-4">
            {cart.items.map((item) => (
              <div key={item.id} className="flex gap-4 p-4 border rounded-lg">
                {/* Product Image */}
                <div className="w-16 h-16 bg-muted rounded-lg flex items-center justify-center overflow-hidden">
                  {item.product.image_url ? (
                    <img 
                      src={item.product.image_url} 
                      alt={item.product.name}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <Package className="h-6 w-6 text-muted-foreground" />
                  )}
                </div>

                {/* Product Details */}
                <div className="flex-1 min-w-0">
                  <h4 className="font-medium line-clamp-2 mb-1">
                    {item.product.name}
                  </h4>
                  <p className="text-sm text-muted-foreground mb-2">
                    ${item.product.price} each
                  </p>
                  
                  {/* Quantity Controls */}
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleUpdateQuantity(item.id, item.quantity - 1)}
                      disabled={isLoading}
                    >
                      <Minus className="h-3 w-3" />
                    </Button>
                    <span className="w-8 text-center text-sm font-medium">
                      {item.quantity}
                    </span>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleUpdateQuantity(item.id, item.quantity + 1)}
                      disabled={isLoading || item.quantity >= item.product.stock}
                    >
                      <Plus className="h-3 w-3" />
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleRemoveItem(item.id)}
                      disabled={isLoading}
                      className="ml-auto text-destructive hover:text-destructive"
                    >
                      <Trash2 className="h-3 w-3" />
                    </Button>
                  </div>
                </div>

                {/* Item Total */}
                <div className="text-right">
                  <p className="font-medium">
                    ${(item.product.price * item.quantity).toFixed(2)}
                  </p>
                  {item.quantity > 1 && (
                    <p className="text-xs text-muted-foreground">
                      {item.quantity} × ${item.product.price}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        <Separator className="my-4" />

        {/* Cart Summary */}
        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Subtotal ({cartItemsCount} items)</span>
              <span>${cartTotal.toFixed(2)}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span>Shipping</span>
              <span className="text-green-600">Free</span>
            </div>
            <div className="flex justify-between text-sm">
              <span>Tax</span>
              <span>${(cartTotal * 0.1).toFixed(2)}</span>
            </div>
            <Separator />
            <div className="flex justify-between font-medium">
              <span>Total</span>
              <span>${(cartTotal * 1.1).toFixed(2)}</span>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="space-y-2">
            <Button className="w-full" size="lg" asChild>
              <Link to="/checkout" onClick={() => setIsOpen(false)}>
                <CreditCard className="mr-2 h-4 w-4" />
                Proceed to Checkout
              </Link>
            </Button>
            <div className="flex gap-2">
              <Button 
                variant="outline" 
                className="flex-1"
                asChild
              >
                <Link to="/products" onClick={() => setIsOpen(false)}>
                  Continue Shopping
                </Link>
              </Button>
              <Button 
                variant="outline" 
                onClick={handleClearCart}
                disabled={isLoading}
              >
                Clear Cart
              </Button>
            </div>
          </div>

          {/* Shipping Info */}
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Truck className="h-4 w-4" />
            <span>Free shipping on orders over $50</span>
          </div>
        </div>
      </div>
    )
  }

  if (!showTrigger) {
    return <CartContent />
  }

  return (
    <Sheet open={isOpen} onOpenChange={setIsOpen}>
      <SheetTrigger asChild>
        {CartTrigger}
      </SheetTrigger>
      <SheetContent className="w-full sm:max-w-lg">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            <ShoppingCart className="h-5 w-5" />
            Shopping Cart
            {cartItemsCount > 0 && (
              <Badge variant="secondary">
                {cartItemsCount} {cartItemsCount === 1 ? 'item' : 'items'}
              </Badge>
            )}
          </SheetTitle>
        </SheetHeader>
        <div className="mt-6 h-full">
          <CartContent />
        </div>
      </SheetContent>
    </Sheet>
  )
}