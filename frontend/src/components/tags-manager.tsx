import { useState, useEffect } from 'react'
import { 
  Tag, 
  Plus, 
  X, 
  Search, 
  Filter,
  Sparkles,
  Hash,
  Edit,
  Trash2,
  Check
} from 'lucide-react'
import { useTagsStore, type Tag as TagType } from '../store/tags'
import { useAuthStore } from '../store/auth'
import { useUIStore } from '../store/ui'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
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
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'

interface TagsManagerProps {
  productId?: string
  mode?: 'admin' | 'product' | 'filter'
  onTagSelect?: (tagId: string) => void
  selectedTags?: string[]
  showCreateButton?: boolean
}

interface CreateTagFormData {
  name: string
  description: string
  color: string
}

const TAG_COLORS = [
  '#ef4444', '#f97316', '#f59e0b', '#eab308', '#84cc16',
  '#22c55e', '#10b981', '#14b8a6', '#06b6d4', '#0ea5e9',
  '#3b82f6', '#6366f1', '#8b5cf6', '#a855f7', '#d946ef',
  '#ec4899', '#f43f5e', '#64748b', '#6b7280', '#374151'
]

export function TagsManager({ 
  productId, 
  mode = 'admin', 
  onTagSelect,
  selectedTags = [],
  showCreateButton = true 
}: TagsManagerProps) {
  const { 
    tags, 
    productTags, 
    isLoading, 
    error, 
    isSubmitting,
    fetchTags, 
    fetchProductTags,
    createTag, 
    addTagToProduct,
    removeTagFromProduct,
    clearError 
  } = useTagsStore()
  
  const { user, isAuthenticated } = useAuthStore()
  const { addToast } = useUIStore()

  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showAddTagDialog, setShowAddTagDialog] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterColor, setFilterColor] = useState<string>('all')
  
  const [createFormData, setCreateFormData] = useState<CreateTagFormData>({
    name: '',
    description: '',
    color: TAG_COLORS[0]
  })

  const currentProductTags = productId ? productTags[productId] || [] : []
  const isAdmin = user?.email?.includes('ubiriagiorgi8') || user?.name?.toLowerCase().includes('admin')

  useEffect(() => {
    fetchTags()
    if (productId) {
      fetchProductTags(productId)
    }
  }, [productId, fetchTags, fetchProductTags])

  useEffect(() => {
    if (error) {
      addToast({
        type: 'error',
        description: error
      })
      clearError()
    }
  }, [error, addToast, clearError])

  const handleCreateTag = async () => {
    if (!createFormData.name.trim()) {
      addToast({
        type: 'error',
        description: 'Please enter a tag name'
      })
      return
    }

    try {
      await createTag({
        name: createFormData.name.trim(),
        description: createFormData.description.trim() || undefined,
        color: createFormData.color
      })

      setCreateFormData({ name: '', description: '', color: TAG_COLORS[0] })
      setShowCreateDialog(false)
      
      addToast({
        type: 'success',
        description: 'Tag created successfully!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const handleAddTagToProduct = async (tagId: string) => {
    if (!productId) return

    try {
      await addTagToProduct(productId, tagId)
      setShowAddTagDialog(false)
      
      addToast({
        type: 'success',
        description: 'Tag added to product!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const handleRemoveTagFromProduct = async (tagId: string) => {
    if (!productId) return

    try {
      await removeTagFromProduct(productId, tagId)
      
      addToast({
        type: 'success',
        description: 'Tag removed from product!'
      })
    } catch (error) {
      // Error handled by store and useEffect
    }
  }

  const handleTagSelect = (tagId: string) => {
    if (onTagSelect) {
      onTagSelect(tagId)
    }
  }

  // Filter tags based on search and color
  const filteredTags = tags.filter(tag => {
    const matchesSearch = tag.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         tag.description?.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesColor = filterColor === 'all' || tag.color === filterColor
    
    return matchesSearch && matchesColor
  })

  // Get available tags for product (not already assigned)
  const availableTagsForProduct = filteredTags.filter(tag => 
    !currentProductTags.some(pt => pt.tag_id === tag.id)
  )

  const TagBadge = ({ 
    tag, 
    onRemove, 
    onClick, 
    isSelected = false,
    showRemove = false 
  }: { 
    tag: TagType; 
    onRemove?: () => void; 
    onClick?: () => void;
    isSelected?: boolean;
    showRemove?: boolean;
  }) => (
    <Badge
      variant={isSelected ? "default" : "outline"}
      className={`cursor-pointer transition-all hover:scale-105 ${
        tag.color ? `border-[${tag.color}] text-[${tag.color}]` : ''
      } ${isSelected ? 'ring-2 ring-primary' : ''}`}
      style={{
        borderColor: tag.color,
        color: isSelected ? 'white' : tag.color,
        backgroundColor: isSelected ? tag.color : 'transparent'
      }}
      onClick={onClick}
    >
      <Hash className="h-3 w-3 mr-1" />
      {tag.name}
      {showRemove && onRemove && (
        <button
          onClick={(e) => {
            e.stopPropagation()
            onRemove()
          }}
          className="ml-1 hover:bg-black/20 rounded-full p-0.5"
        >
          <X className="h-3 w-3" />
        </button>
      )}
      {isSelected && (
        <Check className="h-3 w-3 ml-1" />
      )}
    </Badge>
  )

  // Product mode - tag assignment for specific product
  if (mode === 'product' && productId) {
    return (
      <div className="space-y-4">
        {/* Current Product Tags */}
        <div>
          <div className="flex items-center justify-between mb-3">
            <h4 className="font-medium">Product Tags</h4>
            {isAdmin && (
              <Dialog open={showAddTagDialog} onOpenChange={setShowAddTagDialog}>
                <DialogTrigger asChild>
                  <Button variant="outline" size="sm">
                    <Plus className="h-4 w-4 mr-2" />
                    Add Tag
                  </Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Add Tag to Product</DialogTitle>
                    <DialogDescription>
                      Select a tag to add to this product.
                    </DialogDescription>
                  </DialogHeader>
                  
                  <div className="space-y-4">
                    <div className="relative">
                      <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input
                        placeholder="Search available tags..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="pl-10"
                      />
                    </div>
                    
                    <div className="max-h-60 overflow-y-auto">
                      {availableTagsForProduct.length === 0 ? (
                        <p className="text-center text-muted-foreground py-4">
                          No available tags found
                        </p>
                      ) : (
                        <div className="flex flex-wrap gap-2">
                          {availableTagsForProduct.map((tag) => (
                            <TagBadge
                              key={tag.id}
                              tag={tag}
                              onClick={() => handleAddTagToProduct(tag.id)}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                </DialogContent>
              </Dialog>
            )}
          </div>
          
          {currentProductTags.length === 0 ? (
            <p className="text-sm text-muted-foreground">No tags assigned to this product</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {currentProductTags.map((productTag) => (
                <TagBadge
                  key={productTag.id}
                  tag={productTag.tag}
                  showRemove={isAdmin}
                  onRemove={() => handleRemoveTagFromProduct(productTag.tag_id)}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    )
  }

  return null
} 