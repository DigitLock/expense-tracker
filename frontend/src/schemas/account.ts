import { z } from 'zod'

/**
 * Account validation schemas for Expense Tracker
 * 
 * Matches backend API structure from internal/api/handlers/accounts.go
 */

// Account types as defined in database
export const accountTypes = ['cash', 'bank', 'credit', 'savings', 'investment'] as const

// Currency codes (ISO 4217)
export const currencyCodes = ['RSD', 'EUR', 'USD'] as const

/**
 * Schema for creating a new account
 * POST /api/v1/accounts
 */
export const createAccountSchema = z.object({
  name: z.string()
    .min(1, 'Account name is required')
    .max(100, 'Account name must be less than 100 characters')
    .trim(),
  
  type: z.enum(accountTypes, {
    errorMap: () => ({ message: 'Invalid account type' })
  }),
  
  currency: z.string()
    .length(3, 'Currency must be a 3-letter ISO code')
    .toUpperCase()
    .refine(
      (val) => currencyCodes.includes(val as typeof currencyCodes[number]),
      { message: 'Unsupported currency' }
    ),
  
  initial_balance: z.number()
    .default(0)
    .refine(
      (val) => !isNaN(val) && isFinite(val),
      { message: 'Initial balance must be a valid number' }
    ),
  
  description: z.string()
    .max(500, 'Description must be less than 500 characters')
    .optional()
    .nullable(),
})

/**
 * Schema for updating an existing account
 * PUT /api/v1/accounts/:id
 */
export const updateAccountSchema = z.object({
  name: z.string()
    .min(1, 'Account name is required')
    .max(100, 'Account name must be less than 100 characters')
    .trim()
    .optional(),
  
  description: z.string()
    .max(500, 'Description must be less than 500 characters')
    .optional()
    .nullable(),
  
  is_active: z.boolean()
    .optional(),
})

/**
 * Type inference from schemas
 */
export type CreateAccountInput = z.infer<typeof createAccountSchema>
export type UpdateAccountInput = z.infer<typeof updateAccountSchema>

/**
 * Account type for display
 */
export type AccountType = typeof accountTypes[number]
export type CurrencyCode = typeof currencyCodes[number]

/**
 * Helper for account type labels
 */
export const accountTypeLabels: Record<AccountType, string> = {
  cash: 'Cash',
  bank: 'Bank Account',
  credit: 'Credit Card',
  savings: 'Savings Account',
  investment: 'Investment Account',
}

/**
 * Helper for currency symbols
 */
export const currencySymbols: Record<CurrencyCode, string> = {
  RSD: 'RSD',
  EUR: '€',
  USD: '$',
}
