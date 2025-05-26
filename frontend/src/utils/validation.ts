import { z } from 'zod'

// Auth validation schemas
export const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
})

export const registerSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
})

// Product validation schemas
export const productSearchSchema = z.object({
  query: z.string().min(1, 'Search query is required'),
  category: z.string().optional(),
  minPrice: z.number().min(0).optional(),
  maxPrice: z.number().min(0).optional(),
})

// Cart validation schemas
export const addToCartSchema = z.object({
  productId: z.string().uuid('Invalid product ID'),
  quantity: z.number().min(1, 'Quantity must be at least 1'),
})

// Profile validation schemas
export const updateProfileSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
})

// Type exports
export type LoginFormData = z.infer<typeof loginSchema>
export type RegisterFormData = z.infer<typeof registerSchema>
export type ProductSearchData = z.infer<typeof productSearchSchema>
export type AddToCartData = z.infer<typeof addToCartSchema>
export type UpdateProfileData = z.infer<typeof updateProfileSchema> 