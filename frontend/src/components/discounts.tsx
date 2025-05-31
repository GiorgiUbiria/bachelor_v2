import { useState, useEffect } from 'react'
import { 
  Percent, 
  Plus, 
  DollarSign, 
  Gift,
  Sparkles,
  Copy,
  Check
} from 'lucide-react'
import { useDiscountsStore, type Discount, type SmartDiscountSuggestion } from '../store/discounts'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface DiscountsProps {
  mode?: 'display' | 'apply' | 'admin'
  productId?: string
  category?: string
  orderTotal?: number
  onDiscountApplied?: (discountAmount: number, finalTotal: number) => void
}

interface CreateDiscountFormData {
  discount_type: string
  discount_value: number
  start_date: string
  end_date: string
  product_id?: string
  category?: string
  min_order_amount?: number
  max_discount_amount?: number
  usage_limit?: number
}

export function Discounts({ 
  mode = 'display', 
  productId, 
  category,
  orderTotal = 0,
  onDiscountApplied 
}: DiscountsProps) {
  const { 
    activeDiscounts, 
    allDiscounts,
    smartSuggestions,
    isLoading, 
    error, 
    isSubmitting,
    fetchActiveDiscounts,
    fetchAllDiscounts,
    fetchSmartSuggestions,
    createDiscount,
    applyDiscount,
    clearError 
  } = useDiscountsStore()
  
  const { user } = useAuthStore()
  const { addToast } = useUIStore()

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [discountCode, setDiscountCode] = useState('')
  const [appliedDiscount, setAppliedDiscount] = useState<{ discountAmount: number; finalTotal: number } | null>(null)
  const [copiedCode, setCopiedCode] = useState<string | null>(null)
  
  const [createFormData, setCreateFormData] = useState<CreateDiscountFormData>({
    discount_type: 'percentage',
    discount_value: 10,
    start_date: new Date().toISOString().split('T')[0],
    end_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
    min_order_amount: undefined,
    max_discount_amount: undefined,
    usage_limit: undefined
  })

  const isAdmin = user?.email?.includes('ubiriagiorgi8') || user?.name?.toLowerCase().includes('admin')

  useEffect(() => {
    if (mode === 'admin') {
      fetchAllDiscounts()
    } else {
      fetchActiveDiscounts()
    }
    
    if (productId || category) {
      fetchSmartSuggestions(productId, category)
    }
  }, [mode, productId, category, fetchActiveDiscounts, fetchAllDiscounts, fetchSmartSuggestions])

  useEffect(() => {
    if (error) {
      addToast({
        type: 'error',
        description: error
      })
      clearError()
    }
  }, [error, addToast, clearError])

  const handleCreateDiscount = async () => {
    if (createFormData.discount_value <= 0) {
      addToast({
        type: 'error',
        description: 'Please enter a valid discount value'
      })
      return
    }

    try {
      await createDiscount({
        ...createFormData,
        product_id: createFormData.product_id || undefined,
        category: createFormData.category || undefined
      })

      setCreateFormData({
        discount_type: 'percentage',
        discount_value: 10,
        start_date: new Date().toISOString().split('T')[0],
        end_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        min_order_amount: undefined,
        max_discount_amount: undefined,
        usage_limit: undefined
      })
      setShowCreateDialog(false)
      
      addToast({
        type: 'success',
        description: 'Discount created successfully!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const handleApplyDiscount = async () => {
    if (!discountCode.trim()) {
      addToast({
        type: 'error',
        description: 'Please enter a discount code'
      })
      return
    }

    try {
      const result = await applyDiscount(discountCode.trim(), orderTotal)
      setAppliedDiscount(result)
      
      if (onDiscountApplied) {
        onDiscountApplied(result.discountAmount, result.finalTotal)
      }
      
      addToast({
        type: 'success',
        description: `Discount applied! You saved $${result.discountAmount.toFixed(2)}`
      })
    } catch (error: any) {
      addToast({
        type: 'error',
        description: error.message || 'Failed to apply discount'
      })
    }
  }

  const handleCopyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code)
      setCopiedCode(code)
      setTimeout(() => setCopiedCode(null), 2000)
      
      addToast({
        type: 'success',
        description: 'Discount code copied!'
      })
    } catch (error) {
      addToast({
        type: 'error',
        description: 'Failed to copy code'
      })
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    })
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const getDiscountTypeIcon = (type: string) => {
    switch (type) {
      case 'percentage':
        return <Percent className="h-4 w-4" />
      case 'fixed':
        return <DollarSign className="h-4 w-4" />
      case 'buy_one_get_one':
        return <Gift className="h-4 w-4" />
      default:
        return <Percent className="h-4 w-4" />
    }
  }

  const getDiscountTypeLabel = (type: string) => {
    switch (type) {
      case 'percentage':
        return 'Percentage'
      case 'fixed':
        return 'Fixed Amount'
      case 'buy_one_get_one':
        return 'BOGO'
      default:
        return type
    }
  }

  const getDiscountStatus = (discount: Discount) => {
    const now = new Date()
    const startDate = new Date(discount.start_date)
    const endDate = new Date(discount.end_date)

    if (now < startDate) return { status: 'upcoming', color: 'bg-blue-100 text-blue-800' }
    if (now > endDate) return { status: 'expired', color: 'bg-gray-100 text-gray-800' }
    if (!discount.is_active) return { status: 'inactive', color: 'bg-red-100 text-red-800' }
    return { status: 'active', color: 'bg-green-100 text-green-800' }
  }

  const DiscountCard = ({ discount, showActions = false }: { discount: Discount; showActions?: boolean }) => {
    const status = getDiscountStatus(discount)
    const usagePercentage = discount.usage_limit ? (discount.usage_count / discount.usage_limit) * 100 : 0

    return (
      <Card className="hover:shadow-md transition-shadow">
        <CardContent className="p-4">
          <div className="flex items-start justify-between mb-3">
            <div className="flex items-center gap-2">
              {getDiscountTypeIcon(discount.discount_type)}
              <div>
                <div className="font-medium">
                  {discount.discount_type === 'percentage' 
                    ? `${discount.discount_value}% OFF`
                    : discount.discount_type === 'fixed'
                    ? `$${discount.discount_value} OFF`
                    : 'BOGO'
                  }
                </div>
                <div className="text-sm text-muted-foreground">
                  {getDiscountTypeLabel(discount.discount_type)}
                </div>
              </div>
            </div>
            
            <div className="flex items-center gap-2">
              <Badge className={status.color}>
                {status.status}
              </Badge>
              {showActions && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleCopyCode(discount.id)}
                >
                  {copiedCode === discount.id ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              )}
            </div>
          </div>

          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Valid:</span>
              <span>{formatDate(discount.start_date)} - {formatDate(discount.end_date)}</span>
            </div>
            
            {discount.min_order_amount && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Min Order:</span>
                <span>{formatCurrency(discount.min_order_amount)}</span>
              </div>
            )}
            
            {discount.max_discount_amount && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Max Discount:</span>
                <span>{formatCurrency(discount.max_discount_amount)}</span>
              </div>
            )}
            
            {discount.usage_limit && (
              <div>
                <div className="flex justify-between mb-1">
                  <span className="text-muted-foreground">Usage:</span>
                  <span>{discount.usage_count} / {discount.usage_limit}</span>
                </div>
                <Progress value={usagePercentage} className="h-2" />
              </div>
            )}
          </div>

          {discount.product_id && (
            <Badge variant="outline" className="mt-2">
              Product-specific
            </Badge>
          )}
          
          {discount.category && (
            <Badge variant="outline" className="mt-2">
              {discount.category}
            </Badge>
          )}
        </CardContent>
      </Card>
    )
  }

  const SmartSuggestionCard = ({ suggestion }: { suggestion: SmartDiscountSuggestion }) => (
    <Card className="border-purple-200 bg-purple-50">
      <CardContent className="p-4">
        <div className="flex items-start justify-between mb-3">
          <div className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-purple-600" />
            <div>
              <div className="font-medium text-purple-900">
                AI Suggested: {suggestion.discount_value}% OFF
              </div>
              <div className="text-sm text-purple-700">
                Confidence: {(suggestion.confidence * 100).toFixed(0)}%
              </div>
            </div>
          </div>
          
          {isAdmin && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setCreateFormData(prev => ({
                  ...prev,
                  discount_type: suggestion.discount_type,
                  discount_value: suggestion.discount_value,
                  product_id: suggestion.product_id,
                  category: suggestion.category
                }))
                setShowCreateDialog(true)
              }}
            >
              Create
            </Button>
          )}
        </div>
        
        <p className="text-sm text-purple-800 mb-2">{suggestion.reason}</p>
        
        <div className="flex items-center gap-2">
          <Progress value={suggestion.confidence * 100} className="flex-1 h-2" />
          <span className="text-xs text-purple-600">
            {(suggestion.confidence * 100).toFixed(0)}%
          </span>
        </div>
      </CardContent>
    </Card>
  )

  // Display mode - show active discounts
  if (mode === 'display') {
    return (
      <div className="space-y-4">
        {/* Smart Suggestions */}
        {smartSuggestions.length > 0 && (
          <div>
            <h4 className="font-medium mb-3 flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-purple-500" />
              AI-Suggested Discounts
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {smartSuggestions.map((suggestion, index) => (
                <SmartSuggestionCard key={index} suggestion={suggestion} />
              ))}
            </div>
          </div>
        )}

        {/* Active Discounts */}
        <div>
          <h4 className="font-medium mb-3 flex items-center gap-2">
            <Percent className="h-4 w-4" />
            Available Discounts
          </h4>
          
          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {[...Array(4)].map((_, i) => (
                <Skeleton key={i} className="h-32" />
              ))}
            </div>
          ) : activeDiscounts.length === 0 ? (
            <Card>
              <CardContent className="p-8 text-center">
                <Percent className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h4 className="text-lg font-medium mb-2">No active discounts</h4>
                <p className="text-muted-foreground">
                  Check back later for new deals and offers!
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {activeDiscounts.map((discount) => (
                <DiscountCard key={discount.id} discount={discount} showActions />
              ))}
            </div>
          )}
        </div>
      </div>
    )
  }

  // Apply mode - discount code application
  if (mode === 'apply') {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Gift className="h-5 w-5" />
            Apply Discount Code
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Input
              placeholder="Enter discount code..."
              value={discountCode}
              onChange={(e) => setDiscountCode(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleApplyDiscount()}
            />
            <Button 
              onClick={handleApplyDiscount}
              disabled={isSubmitting || !discountCode.trim()}
            >
              {isSubmitting ? 'Applying...' : 'Apply'}
            </Button>
          </div>

          {appliedDiscount && (
            <Alert className="border-green-200 bg-green-50">
              <Check className="h-4 w-4 text-green-600" />
              <AlertDescription className="text-green-800">
                <div className="flex justify-between items-center">
                  <span>Discount Applied!</span>
                  <div className="text-right">
                    <div className="text-sm">Savings: {formatCurrency(appliedDiscount.discountAmount)}</div>
                    <div className="font-medium">New Total: {formatCurrency(appliedDiscount.finalTotal)}</div>
                  </div>
                </div>
              </AlertDescription>
            </Alert>
          )}

          {activeDiscounts.length > 0 && (
            <div>
              <Separator className="my-4" />
              <h4 className="font-medium mb-3">Available Discounts</h4>
              <div className="space-y-2">
                {activeDiscounts.slice(0, 3).map((discount) => (
                  <div key={discount.id} className="flex items-center justify-between p-2 border rounded">
                    <div className="text-sm">
                      <div className="font-medium">
                        {discount.discount_type === 'percentage' 
                          ? `${discount.discount_value}% OFF`
                          : `$${discount.discount_value} OFF`
                        }
                      </div>
                      {discount.min_order_amount && (
                        <div className="text-muted-foreground">
                          Min order: {formatCurrency(discount.min_order_amount)}
                        </div>
                      )}
                    </div>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setDiscountCode(discount.id)
                        handleApplyDiscount()
                      }}
                    >
                      Apply
                    </Button>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    )
  }

  // Admin mode - full discount management
  if (mode === 'admin' && isAdmin) {
    return (
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Percent className="h-5 w-5" />
            <h3 className="text-lg font-semibold">Discounts Management</h3>
          </div>
          
          <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                Create Discount
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl">
              <DialogHeader>
                <DialogTitle>Create New Discount</DialogTitle>
                <DialogDescription>
                  Create a new discount that customers can apply to their orders.
                </DialogDescription>
              </DialogHeader>
              
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="discount-type">Discount Type</Label>
                  <Select 
                    value={createFormData.discount_type} 
                    onValueChange={(value) => setCreateFormData(prev => ({ ...prev, discount_type: value }))}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="percentage">Percentage</SelectItem>
                      <SelectItem value="fixed">Fixed Amount</SelectItem>
                      <SelectItem value="buy_one_get_one">Buy One Get One</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                
                <div>
                  <Label htmlFor="discount-value">
                    {createFormData.discount_type === 'percentage' ? 'Percentage (%)' : 'Amount ($)'}
                  </Label>
                  <Input
                    id="discount-value"
                    type="number"
                    value={createFormData.discount_value}
                    onChange={(e) => setCreateFormData(prev => ({ ...prev, discount_value: Number(e.target.value) }))}
                    min="0"
                    step={createFormData.discount_type === 'percentage' ? '1' : '0.01'}
                  />
                </div>
                
                <div>
                  <Label htmlFor="start-date">Start Date</Label>
                  <Input
                    id="start-date"
                    type="date"
                    value={createFormData.start_date}
                    onChange={(e) => setCreateFormData(prev => ({ ...prev, start_date: e.target.value }))}
                  />
                </div>
                
                <div>
                  <Label htmlFor="end-date">End Date</Label>
                  <Input
                    id="end-date"
                    type="date"
                    value={createFormData.end_date}
                    onChange={(e) => setCreateFormData(prev => ({ ...prev, end_date: e.target.value }))}
                  />
                </div>
                
                <div>
                  <Label htmlFor="min-order">Min Order Amount ($)</Label>
                  <Input
                    id="min-order"
                    type="number"
                    value={createFormData.min_order_amount || ''}
                    onChange={(e) => setCreateFormData(prev => ({ 
                      ...prev, 
                      min_order_amount: e.target.value ? Number(e.target.value) : undefined 
                    }))}
                    min="0"
                    step="0.01"
                    placeholder="Optional"
                  />
                </div>
                
                <div>
                  <Label htmlFor="max-discount">Max Discount Amount ($)</Label>
                  <Input
                    id="max-discount"
                    type="number"
                    value={createFormData.max_discount_amount || ''}
                    onChange={(e) => setCreateFormData(prev => ({ 
                      ...prev, 
                      max_discount_amount: e.target.value ? Number(e.target.value) : undefined 
                    }))}
                    min="0"
                    step="0.01"
                    placeholder="Optional"
                  />
                </div>
                
                <div>
                  <Label htmlFor="usage-limit">Usage Limit</Label>
                  <Input
                    id="usage-limit"
                    type="number"
                    value={createFormData.usage_limit || ''}
                    onChange={(e) => setCreateFormData(prev => ({ 
                      ...prev, 
                      usage_limit: e.target.value ? Number(e.target.value) : undefined 
                    }))}
                    min="1"
                    placeholder="Optional"
                  />
                </div>
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={() => setShowCreateDialog(false)}>
                  Cancel
                </Button>
                <Button 
                  onClick={handleCreateDiscount}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? 'Creating...' : 'Create Discount'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>

        {/* Smart Suggestions */}
        {smartSuggestions.length > 0 && (
          <div>
            <h4 className="font-medium mb-3 flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-purple-500" />
              AI-Suggested Discounts
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {smartSuggestions.map((suggestion, index) => (
                <SmartSuggestionCard key={index} suggestion={suggestion} />
              ))}
            </div>
          </div>
        )}

        {/* All Discounts */}
        <div>
          <h4 className="font-medium mb-3">All Discounts</h4>
          
          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {[...Array(6)].map((_, i) => (
                <Skeleton key={i} className="h-40" />
              ))}
            </div>
          ) : allDiscounts.length === 0 ? (
            <Card>
              <CardContent className="p-12 text-center">
                <Percent className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h4 className="text-lg font-medium mb-2">No discounts created</h4>
                <p className="text-muted-foreground mb-4">
                  Create your first discount to start offering deals to customers.
                </p>
                <Button onClick={() => setShowCreateDialog(true)}>
                  Create First Discount
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {allDiscounts.map((discount) => (
                <DiscountCard key={discount.id} discount={discount} showActions />
              ))}
            </div>
          )}
        </div>
      </div>
    )
  }

  return null
} 