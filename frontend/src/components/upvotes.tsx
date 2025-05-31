import { useState, useEffect } from 'react'
import { ThumbsUp, Heart } from 'lucide-react'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import apiService from '../services/api'

interface UpvotesProps {
  productId: string
  variant?: 'button' | 'compact' | 'minimal'
  showCount?: boolean
  className?: string
}

interface UpvoteData {
  total_upvotes: number
  user_upvoted: boolean
  upvotes: Array<{
    id: string
    user_id: string
    product_id: string
    created_at: string
  }>
}

export function Upvotes({ 
  productId, 
  variant = 'button', 
  showCount = true, 
  className = '' 
}: UpvotesProps) {
  const { isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()

  const [upvoteData, setUpvoteData] = useState<UpvoteData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isToggling, setIsToggling] = useState(false)

  useEffect(() => {
    fetchUpvotes()
  }, [productId])

  const fetchUpvotes = async () => {
    try {
      setIsLoading(true)
      const response = await apiService.upvotes.getByProduct(productId)
      setUpvoteData(response.data)
    } catch (error) {
      console.error('Error fetching upvotes:', error)
      // Set default data if fetch fails
      setUpvoteData({
        total_upvotes: 0,
        user_upvoted: false,
        upvotes: []
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleToggleUpvote = async () => {
    if (!isAuthenticated) {
      addToast({
        type: 'error',
        description: 'Please sign in to upvote products'
      })
      return
    }

    if (!upvoteData || isToggling) return

    setIsToggling(true)

    try {
      if (upvoteData.user_upvoted) {
        // Remove upvote
        await apiService.upvotes.remove(productId)
        setUpvoteData(prev => prev ? {
          ...prev,
          total_upvotes: prev.total_upvotes - 1,
          user_upvoted: false
        } : null)
        
        addToast({
          type: 'success',
          description: 'Upvote removed'
        })
      } else {
        // Add upvote
        await apiService.upvotes.add(productId)
        setUpvoteData(prev => prev ? {
          ...prev,
          total_upvotes: prev.total_upvotes + 1,
          user_upvoted: true
        } : null)
        
        addToast({
          type: 'success',
          description: 'Product upvoted!'
        })
      }
    } catch (error: any) {
      addToast({
        type: 'error',
        description: error.response?.data?.error || 'Failed to update upvote'
      })
    } finally {
      setIsToggling(false)
    }
  }

  if (isLoading) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <Skeleton className="h-8 w-8" />
        {showCount && <Skeleton className="h-4 w-8" />}
      </div>
    )
  }

  if (!upvoteData) {
    return null
  }

  const { total_upvotes, user_upvoted } = upvoteData

  // Minimal variant - just the icon and count
  if (variant === 'minimal') {
    return (
      <button
        onClick={handleToggleUpvote}
        disabled={isToggling}
        className={`flex items-center gap-1 text-sm transition-colors hover:text-primary ${
          user_upvoted ? 'text-primary' : 'text-muted-foreground'
        } ${className}`}
      >
        <ThumbsUp 
          className={`h-4 w-4 ${user_upvoted ? 'fill-current' : ''} ${
            isToggling ? 'animate-pulse' : ''
          }`} 
        />
        {showCount && <span>{total_upvotes}</span>}
      </button>
    )
  }

  // Compact variant - smaller button with badge
  if (variant === 'compact') {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <Button
          variant={user_upvoted ? 'default' : 'outline'}
          size="sm"
          onClick={handleToggleUpvote}
          disabled={isToggling}
          className="h-8 px-2"
        >
          <ThumbsUp 
            className={`h-3 w-3 ${user_upvoted ? 'fill-current' : ''} ${
              isToggling ? 'animate-pulse' : ''
            }`} 
          />
        </Button>
        {showCount && (
          <Badge variant="secondary" className="text-xs">
            {total_upvotes}
          </Badge>
        )}
      </div>
    )
  }

  // Default button variant
  return (
    <Button
      variant={user_upvoted ? 'default' : 'outline'}
      onClick={handleToggleUpvote}
      disabled={isToggling}
      className={`flex items-center gap-2 ${className}`}
    >
      <ThumbsUp 
        className={`h-4 w-4 ${user_upvoted ? 'fill-current' : ''} ${
          isToggling ? 'animate-pulse' : ''
        }`} 
      />
      {showCount && <span>{total_upvotes}</span>}
      <span className="hidden sm:inline">
        {user_upvoted ? 'Upvoted' : 'Upvote'}
      </span>
    </Button>
  )
}

// Alternative heart-style upvote component
export function HeartUpvotes({ 
  productId, 
  variant = 'button', 
  showCount = true, 
  className = '' 
}: UpvotesProps) {
  const { isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()

  const [upvoteData, setUpvoteData] = useState<UpvoteData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [isToggling, setIsToggling] = useState(false)

  useEffect(() => {
    fetchUpvotes()
  }, [productId])

  const fetchUpvotes = async () => {
    try {
      setIsLoading(true)
      const response = await apiService.upvotes.getByProduct(productId)
      setUpvoteData(response.data)
    } catch (error) {
      console.error('Error fetching upvotes:', error)
      setUpvoteData({
        total_upvotes: 0,
        user_upvoted: false,
        upvotes: []
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleToggleUpvote = async () => {
    if (!isAuthenticated) {
      addToast({
        type: 'error',
        description: 'Please sign in to like products'
      })
      return
    }

    if (!upvoteData || isToggling) return

    setIsToggling(true)

    try {
      if (upvoteData.user_upvoted) {
        await apiService.upvotes.remove(productId)
        setUpvoteData(prev => prev ? {
          ...prev,
          total_upvotes: prev.total_upvotes - 1,
          user_upvoted: false
        } : null)
      } else {
        await apiService.upvotes.add(productId)
        setUpvoteData(prev => prev ? {
          ...prev,
          total_upvotes: prev.total_upvotes + 1,
          user_upvoted: true
        } : null)
      }
    } catch (error: any) {
      addToast({
        type: 'error',
        description: error.response?.data?.error || 'Failed to update like'
      })
    } finally {
      setIsToggling(false)
    }
  }

  if (isLoading) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <Skeleton className="h-8 w-8" />
        {showCount && <Skeleton className="h-4 w-8" />}
      </div>
    )
  }

  if (!upvoteData) {
    return null
  }

  const { total_upvotes, user_upvoted } = upvoteData

  return (
    <button
      onClick={handleToggleUpvote}
      disabled={isToggling}
      className={`flex items-center gap-2 transition-all hover:scale-105 ${className}`}
    >
      <Heart 
        className={`h-5 w-5 transition-colors ${
          user_upvoted 
            ? 'fill-red-500 text-red-500' 
            : 'text-muted-foreground hover:text-red-500'
        } ${isToggling ? 'animate-pulse' : ''}`} 
      />
      {showCount && (
        <span className={`text-sm ${user_upvoted ? 'text-red-500' : 'text-muted-foreground'}`}>
          {total_upvotes}
        </span>
      )}
    </button>
  )
} 