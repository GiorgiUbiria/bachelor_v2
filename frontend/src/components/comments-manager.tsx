import { useState, useEffect } from 'react'
import { 
  MessageCircle, 
  Star, 
  Edit, 
  Trash2, 
  Send, 
  X, 
  Save,
  AlertCircle,
  User
} from 'lucide-react'
import { useCommentsStore, type Comment } from '../store/comments'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface CommentsManagerProps {
  productId: string
  productName?: string
}

interface CommentFormData {
  content: string
  rating: number
}

export function CommentsManager({ productId, productName }: CommentsManagerProps) {
  const { 
    commentsByProduct, 
    isLoading, 
    error, 
    isSubmitting,
    fetchComments, 
    addComment, 
    updateComment, 
    deleteComment,
    clearError 
  } = useCommentsStore()
  
  const { user, isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()

  const [showAddForm, setShowAddForm] = useState(false)
  const [editingComment, setEditingComment] = useState<Comment | null>(null)
  const [deleteConfirmComment, setDeleteConfirmComment] = useState<Comment | null>(null)
  
  const [formData, setFormData] = useState<CommentFormData>({
    content: '',
    rating: 5
  })

  const comments = commentsByProduct[productId] || []

  useEffect(() => {
    fetchComments(productId)
  }, [productId, fetchComments])

  useEffect(() => {
    if (error) {
      addToast({
        type: 'error',
        description: error
      })
      clearError()
    }
  }, [error, addToast, clearError])

  const handleSubmitComment = async () => {
    if (!formData.content.trim()) {
      addToast({
        type: 'error',
        description: 'Please enter a comment'
      })
      return
    }

    try {
      await addComment({
        product_id: productId,
        content: formData.content.trim(),
        rating: formData.rating
      })

      setFormData({ content: '', rating: 5 })
      setShowAddForm(false)
      
      addToast({
        type: 'success',
        description: 'Comment added successfully!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const handleEditComment = async () => {
    if (!editingComment || !formData.content.trim()) return

    try {
      await updateComment(editingComment.id, {
        content: formData.content.trim(),
        rating: formData.rating
      })

      setEditingComment(null)
      setFormData({ content: '', rating: 5 })
      
      addToast({
        type: 'success',
        description: 'Comment updated successfully!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const handleDeleteComment = async () => {
    if (!deleteConfirmComment) return

    try {
      await deleteComment(deleteConfirmComment.id, productId)
      setDeleteConfirmComment(null)
      
      addToast({
        type: 'success',
        description: 'Comment deleted successfully!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const startEdit = (comment: Comment) => {
    setEditingComment(comment)
    setFormData({
      content: comment.content,
      rating: comment.rating
    })
    setShowAddForm(false)
  }

  const cancelEdit = () => {
    setEditingComment(null)
    setFormData({ content: '', rating: 5 })
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const canEditComment = (comment: Comment) => {
    return isAuthenticated && user && comment.user_id === user.id
  }

  const StarRating = ({ 
    rating, 
    onRatingChange, 
    readonly = false 
  }: { 
    rating: number; 
    onRatingChange?: (rating: number) => void; 
    readonly?: boolean 
  }) => (
    <div className="flex items-center gap-1">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          onClick={() => !readonly && onRatingChange?.(star)}
          disabled={readonly}
          className={`${readonly ? 'cursor-default' : 'cursor-pointer hover:scale-110'} transition-transform`}
        >
          <Star
            className={`h-4 w-4 ${
              star <= rating
                ? 'fill-yellow-400 text-yellow-400'
                : 'text-muted-foreground'
            }`}
          />
        </button>
      ))}
    </div>
  )

  const CommentForm = ({ 
    isEdit = false, 
    onSubmit, 
    onCancel 
  }: { 
    isEdit?: boolean; 
    onSubmit: () => void; 
    onCancel: () => void 
  }) => (
    <Card className="border-primary/20">
      <CardContent className="p-4">
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium mb-2 block">Rating</label>
            <StarRating 
              rating={formData.rating} 
              onRatingChange={(rating) => setFormData(prev => ({ ...prev, rating }))}
            />
          </div>
          
          <div>
            <label className="text-sm font-medium mb-2 block">Your Review</label>
            <Textarea
              value={formData.content}
              onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
              placeholder={`Share your thoughts about ${productName || 'this product'}...`}
              rows={4}
              maxLength={1000}
            />
            <div className="text-xs text-muted-foreground mt-1">
              {formData.content.length}/1000 characters
            </div>
          </div>

          <div className="flex gap-2">
            <Button 
              onClick={onSubmit}
              disabled={isSubmitting || !formData.content.trim()}
              className="flex-1"
            >
              {isSubmitting ? (
                <>Loading...</>
              ) : (
                <>
                  {isEdit ? <Save className="h-4 w-4 mr-2" /> : <Send className="h-4 w-4 mr-2" />}
                  {isEdit ? 'Update Review' : 'Post Review'}
                </>
              )}
            </Button>
            <Button variant="outline" onClick={onCancel}>
              <X className="h-4 w-4 mr-2" />
              Cancel
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <MessageCircle className="h-5 w-5" />
          <h3 className="text-lg font-semibold">
            Customer Reviews ({comments.length})
          </h3>
        </div>
        
        {isAuthenticated && !showAddForm && !editingComment && (
          <Button onClick={() => setShowAddForm(true)}>
            <MessageCircle className="h-4 w-4 mr-2" />
            Write Review
          </Button>
        )}
      </div>

      {/* Authentication Notice */}
      {!isAuthenticated && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Please <strong>sign in</strong> to write a review for this product.
          </AlertDescription>
        </Alert>
      )}

      {/* Add Comment Form */}
      {showAddForm && (
        <CommentForm
          onSubmit={handleSubmitComment}
          onCancel={() => {
            setShowAddForm(false)
            setFormData({ content: '', rating: 5 })
          }}
        />
      )}

      {/* Edit Comment Form */}
      {editingComment && (
        <CommentForm
          isEdit
          onSubmit={handleEditComment}
          onCancel={cancelEdit}
        />
      )}

      {/* Loading State */}
      {isLoading && comments.length === 0 && (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <Card key={i}>
              <CardContent className="p-4">
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-4 w-24" />
                  </div>
                  <Skeleton className="h-4 w-20" />
                  <Skeleton className="h-16 w-full" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Comments List */}
      {comments.length === 0 && !isLoading ? (
        <Card>
          <CardContent className="p-12 text-center">
            <MessageCircle className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h4 className="text-lg font-medium mb-2">No reviews yet</h4>
            <p className="text-muted-foreground mb-4">
              Be the first to review this product and help other customers!
            </p>
            {isAuthenticated && (
              <Button onClick={() => setShowAddForm(true)}>
                Write the First Review
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4">
          {comments.map((comment) => (
            <Card key={comment.id} className="relative">
              <CardContent className="p-4">
                <div className="space-y-3">
                  {/* Comment Header */}
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                        <User className="h-4 w-4 text-primary" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{comment.user_name}</span>
                          <StarRating rating={comment.rating} readonly />
                        </div>
                        <span className="text-sm text-muted-foreground">
                          {formatDate(comment.created_at)}
                          {comment.updated_at !== comment.created_at && (
                            <span className="ml-2">(edited)</span>
                          )}
                        </span>
                      </div>
                    </div>

                    {/* Action Buttons */}
                    {canEditComment(comment) && (
                      <div className="flex gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => startEdit(comment)}
                          disabled={isSubmitting}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => setDeleteConfirmComment(comment)}
                          disabled={isSubmitting}
                          className="text-destructive hover:text-destructive"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    )}
                  </div>

                  {/* Comment Content */}
                  <div className="pl-11">
                    <p className="text-muted-foreground leading-relaxed">
                      {comment.content}
                    </p>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={!!deleteConfirmComment} onOpenChange={() => setDeleteConfirmComment(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Review</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this review? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteConfirmComment(null)}>
              Cancel
            </Button>
            <Button 
              variant="destructive" 
              onClick={handleDeleteComment}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Deleting...' : 'Delete Review'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
} 