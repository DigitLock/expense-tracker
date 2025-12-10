import { z } from 'zod'

/**
 * Category validation schemas for Expense Tracker
 * 
 * Matches backend API structure from internal/api/handlers/categories.go
 */

// Category types as defined in database
export const categoryTypes = ['income', 'expense'] as const

/**
 * Schema for creating a new category
 * POST /api/v1/categories
 */
export const createCategorySchema = z.object({
  name: z.string()
    .min(1, 'Category name is required')
    .max(100, 'Category name must be less than 100 characters')
    .trim(),
  
  type: z.enum(categoryTypes, {
    errorMap: () => ({ message: 'Category type must be either income or expense' })
  }),
  
  parent_id: z.string()
    .uuid('Invalid parent category ID')
    .optional()
    .nullable(),

})

/**
 * Schema for updating an existing category
 * PUT /api/v1/categories/:id
 */
export const updateCategorySchema = z.object({
  name: z.string()
    .min(1, 'Category name is required')
    .max(100, 'Category name must be less than 100 characters')
    .trim()
    .optional(),
  
  parent_id: z.string()
    .uuid('Invalid parent category ID')
    .optional()
    .nullable(),

  is_active: z.boolean()
    .optional(),
})

/**
 * Type inference from schemas
 */
export type CreateCategoryInput = z.infer<typeof createCategorySchema>
export type UpdateCategoryInput = z.infer<typeof updateCategorySchema>

/**
 * Category type for display
 */
export type CategoryType = typeof categoryTypes[number]

/**
 * Helper for category type labels
 */
export const categoryTypeLabels: Record<CategoryType, string> = {
  income: 'Income',
  expense: 'Expense',
}

/**
 * Helper for category type colors (Tailwind CSS classes)
 */
export const categoryTypeColors: Record<CategoryType, string> = {
  income: 'text-green-600 bg-green-50',
  expense: 'text-red-600 bg-red-50',
}

/**
 * Helper for category type icons (Lucide icon names)
 */
export const categoryTypeIcons: Record<CategoryType, string> = {
  income: 'TrendingUp',
  expense: 'TrendingDown',
}
