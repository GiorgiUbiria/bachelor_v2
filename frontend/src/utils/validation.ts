import { z } from 'zod'

// Auth schemas
export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters').max(128, 'Password must be at most 128 characters'),
})

export const registerSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(100, 'Name must be at most 100 characters'),
  email: z.string().email('Please enter a valid email address').max(255, 'Email must be at most 255 characters'),
  password: z.string().min(8, 'Password must be at least 8 characters').max(128, 'Password must be at most 128 characters'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

export const updateProfileSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(100, 'Name must be at most 100 characters').optional(),
  phone: z.string().max(20, 'Phone must be at most 20 characters').optional(),
})

// Product schemas
export const createProductSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be at most 255 characters'),
  description: z.string().min(1, 'Description is required').max(1000, 'Description must be at most 1000 characters'),
  price: z.number().min(0.01, 'Price must be at least 0.01'),
  stock: z.number().int().min(0, 'Stock must be at least 0'),
  category: z.string().min(1, 'Category is required').max(100, 'Category must be at most 100 characters'),
  image_url: z.string().url('Invalid image URL').optional(),
})

export const updateProductSchema = z.object({
  name: z.string().min(1, 'Name is required').max(255, 'Name must be at most 255 characters').optional(),
  description: z.string().min(1, 'Description is required').max(1000, 'Description must be at most 1000 characters').optional(),
  price: z.number().min(0.01, 'Price must be at least 0.01').optional(),
  stock: z.number().int().min(0, 'Stock must be at least 0').optional(),
  category: z.string().min(1, 'Category is required').max(100, 'Category must be at most 100 characters').optional(),
  image_url: z.string().url('Invalid image URL').optional(),
})

export const productSearchSchema = z.object({
  q: z.string().min(1, 'Search query is required').optional(),
  category: z.string().optional(),
  page: z.number().int().min(1).optional(),
  limit: z.number().int().min(1).max(100).optional(),
})

// Cart schemas
export const addToCartSchema = z.object({
  product_id: z.string().uuid('Invalid product ID'),
  quantity: z.number().int().min(1, 'Quantity must be at least 1').max(100, 'Quantity must be at most 100'),
})

export const updateCartItemSchema = z.object({
  quantity: z.number().int().min(0, 'Quantity must be at least 0').max(100, 'Quantity must be at most 100'),
})

// Order schemas
export const createOrderSchema = z.object({
  cart_item_ids: z.array(z.string().uuid()).optional(),
  payment_method: z.string().min(1, 'Payment method is required'),
  shipping_address: z.string().min(1, 'Shipping address is required'),
})

export const updateOrderStatusSchema = z.object({
  status: z.enum(['pending', 'processing', 'shipped', 'delivered', 'cancelled'], {
    errorMap: () => ({ message: 'Status must be one of: pending, processing, shipped, delivered, cancelled' })
  }),
})

// Comment schemas
export const addCommentSchema = z.object({
  product_id: z.string().uuid('Invalid product ID'),
  content: z.string().min(1, 'Content is required').max(1000, 'Content must be at most 1000 characters'),
  rating: z.number().int().min(1, 'Rating must be at least 1').max(5, 'Rating must be at most 5'),
})

export const updateCommentSchema = z.object({
  content: z.string().min(1, 'Content is required').max(1000, 'Content must be at most 1000 characters').optional(),
  rating: z.number().int().min(1, 'Rating must be at least 1').max(5, 'Rating must be at most 5').optional(),
})

// Discount schemas
export const createDiscountSchema = z.object({
  discount_type: z.string().min(1, 'Discount type is required'),
  discount_value: z.number().min(0, 'Discount value must be at least 0'),
  start_date: z.string().min(1, 'Start date is required'),
  end_date: z.string().min(1, 'End date is required'),
  product_id: z.string().uuid('Invalid product ID').optional(),
  category: z.string().optional(),
  min_order_amount: z.number().min(0).optional(),
  max_discount_amount: z.number().min(0).optional(),
  usage_limit: z.number().int().min(1).optional(),
})

// Favorite schemas
export const addFavoriteSchema = z.object({
  product_id: z.string().uuid('Invalid product ID'),
})

// Tag schemas
export const createTagSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  color: z.string().optional(),
})

export const addProductTagSchema = z.object({
  product_id: z.string().uuid('Invalid product ID'),
  tag_id: z.string().uuid('Invalid tag ID'),
})

// Upvote schemas
export const addUpvoteSchema = z.object({
  product_id: z.string().uuid('Invalid product ID'),
})

// ML schemas
export const mlSearchSchema = z.object({
  q: z.string().min(1, 'Search query is required'),
  user_id: z.string().uuid().optional(),
  category: z.string().optional(),
  min_price: z.number().min(0).optional(),
  max_price: z.number().min(0).optional(),
  sort_by: z.string().optional(),
  limit: z.number().int().min(1).max(100).optional(),
})

export const generateRecommendationsSchema = z.object({
  user_id: z.string().uuid('Invalid user ID'),
  algorithm: z.string().optional(),
  limit: z.number().int().min(1).max(100).optional(),
})

// Type exports
export type LoginFormData = z.infer<typeof loginSchema>
export type RegisterFormData = z.infer<typeof registerSchema>
export type UpdateProfileData = z.infer<typeof updateProfileSchema>

export type CreateProductData = z.infer<typeof createProductSchema>
export type UpdateProductData = z.infer<typeof updateProductSchema>
export type ProductSearchData = z.infer<typeof productSearchSchema>

export type AddToCartData = z.infer<typeof addToCartSchema>
export type UpdateCartItemData = z.infer<typeof updateCartItemSchema>

export type CreateOrderData = z.infer<typeof createOrderSchema>
export type UpdateOrderStatusData = z.infer<typeof updateOrderStatusSchema>

export type AddCommentData = z.infer<typeof addCommentSchema>
export type UpdateCommentData = z.infer<typeof updateCommentSchema>

export type CreateDiscountData = z.infer<typeof createDiscountSchema>
export type AddFavoriteData = z.infer<typeof addFavoriteSchema>

export type CreateTagData = z.infer<typeof createTagSchema>
export type AddProductTagData = z.infer<typeof addProductTagSchema>

export type AddUpvoteData = z.infer<typeof addUpvoteSchema>

export type MLSearchData = z.infer<typeof mlSearchSchema>
export type GenerateRecommendationsData = z.infer<typeof generateRecommendationsSchema> 