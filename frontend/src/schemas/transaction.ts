import { z } from 'zod'

/**
 * Transaction validation schemas for Expense Tracker
 *
 * Matches backend API structure from internal/api/handlers/transactions.go
 */

// Transaction types as defined in database
export const transactionTypes = ['income', 'expense'] as const

/**
 * Schema for creating a new transaction
 * POST /api/v1/transactions
 */
export const createTransactionSchema = z.object({
    type: z.enum(transactionTypes, {
        errorMap: () => ({ message: 'Transaction type must be income or expense' })
    }),

    amount: z.number()
        .positive('Amount must be greater than zero')
        .finite('Amount must be a valid number')
        .refine(
            (val) => Number(val.toFixed(2)) === val || val.toString().split('.')[1]?.length <= 2,
            { message: 'Amount can have maximum 2 decimal places' }
        ),

    currency: z.enum(['RSD', 'EUR', 'USD'], {
        errorMap: () => ({ message: 'Invalid currency' })
    }),

    account_id: z.string()
        .uuid('Invalid account ID'),

    category_id: z.string()
        .uuid('Invalid category ID'),

    date: z.string()
        .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
        .refine(
            (date) => {
                const parsed = new Date(date)
                return !isNaN(parsed.getTime())
            },
            { message: 'Invalid date' }
        )
        .refine(
            (date) => {
                const parsed = new Date(date)
                const now = new Date()
                now.setHours(23, 59, 59, 999) // End of today
                return parsed <= now
            },
            { message: 'Transaction date cannot be in the future' }
        ),

    description: z.string()
        .max(500, 'Description must be less than 500 characters')
        .optional()
        .or(z.literal('')),
})

/**
 * Schema for updating an existing transaction
 * PATCH /api/v1/transactions/:id
 */
export const updateTransactionSchema = z.object({
    amount: z.number()
        .positive('Amount must be greater than zero')
        .finite('Amount must be a valid number')
        .refine(
            (val) => Number(val.toFixed(2)) === val || val.toString().split('.')[1]?.length <= 2,
            { message: 'Amount can have maximum 2 decimal places' }
        )
        .optional(),

    currency: z.enum(['RSD', 'EUR', 'USD']).optional(),

    category_id: z.string()
        .uuid('Invalid category ID')
        .optional(),

    date: z.string()
        .regex(/^\d{4}-\d{2}-\d{2}$/, 'Date must be in YYYY-MM-DD format')
        .refine(
            (date) => {
                const parsed = new Date(date)
                return !isNaN(parsed.getTime())
            },
            { message: 'Invalid date' }
        )
        .refine(
            (date) => {
                const parsed = new Date(date)
                const now = new Date()
                now.setHours(23, 59, 59, 999)
                return parsed <= now
            },
            { message: 'Transaction date cannot be in the future' }
        )
        .optional(),

    description: z.string()
        .max(500, 'Description must be less than 500 characters')
        .optional()
        .or(z.literal('')),
})

/**
 * Type inference from schemas
 */
export type CreateTransactionInput = z.infer<typeof createTransactionSchema>
export type UpdateTransactionInput = z.infer<typeof updateTransactionSchema>

/**
 * Transaction type for display
 */
export type TransactionType = typeof transactionTypes[number]

/**
 * Helper for transaction type labels
 */
export const transactionTypeLabels: Record<TransactionType, string> = {
    income: 'Income',
    expense: 'Expense',
}

/**
 * Helper for transaction type colors (Tailwind CSS classes)
 */
export const transactionTypeColors: Record<TransactionType, string> = {
    income: 'text-green-600 bg-green-50',
    expense: 'text-red-600 bg-red-50',
}

/**
 * Helper for transaction type icons (Lucide icon names)
 */
export const transactionTypeIcons: Record<TransactionType, string> = {
    income: 'ArrowDownLeft',
    expense: 'ArrowUpRight',
}

/**
 * Helper function to format date for API (YYYY-MM-DD)
 */
export function formatDateForAPI(date: Date): string {
    const year = date.getFullYear()
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const day = String(date.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
}

/**
 * Helper function to get today's date in YYYY-MM-DD format
 */
export function getTodayDateString(): string {
    return formatDateForAPI(new Date())
}

/**
 * Helper function to format date for display
 */
export function formatDateForDisplay(dateStr: string): string {
    const date = new Date(dateStr)
    return new Intl.DateTimeFormat('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
    }).format(date)
}

/**
 * Helper function to format amount with currency
 */
export function formatAmount(amount: string | number, currency: string = 'RSD'): string {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount

    return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
    }).format(num) + ' ' + currency
}