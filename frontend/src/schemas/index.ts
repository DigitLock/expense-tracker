/**
 * Validation schemas for Expense Tracker
 * 
 * This module exports Zod validation schemas for all entities in the application.
 * These schemas are used with VeeValidate for form validation on the frontend.
 * 
 * @module schemas
 */

// ============================================
// Account Schemas
// ============================================
export {
  createAccountSchema,
  updateAccountSchema,
  accountTypes,
  currencyCodes,
  accountTypeLabels,
  currencySymbols,
  type CreateAccountInput,
  type UpdateAccountInput,
  type AccountType,
  type CurrencyCode,
} from './account'

// ============================================
// Category Schemas
// ============================================
export {
  createCategorySchema,
  updateCategorySchema,
  categoryTypes,
  categoryTypeLabels,
  categoryTypeColors,
  categoryTypeIcons,
  type CreateCategoryInput,
  type UpdateCategoryInput,
  type CategoryType,
} from './category'

// ============================================
// Transaction Schemas
// ============================================
export {
    createTransactionSchema,
    updateTransactionSchema,
    transactionTypes,
    transactionTypeLabels,
    transactionTypeColors,
    transactionTypeIcons,
    formatDateForAPI,
    getTodayDateString,
    formatDateForDisplay,  // ← ДОБАВЬ (есть в файле)
    formatAmount,          // ← ДОБАВЬ (есть в файле)
    type CreateTransactionInput,
    type UpdateTransactionInput,
    type TransactionType,
} from './transaction'

/**
 * Usage examples:
 * 
 * @example
 * ```typescript
 * import { createAccountSchema } from '@/schemas'
 * import { useForm } from 'vee-validate'
 * import { toTypedSchema } from '@vee-validate/zod'
 * 
 * const { handleSubmit } = useForm({
 *   validationSchema: toTypedSchema(createAccountSchema)
 * })
 * 
 * const onSubmit = handleSubmit(async (values) => {
 *   // values is typed as CreateAccountInput
 *   await api.createAccount(values)
 * })
 * ```
 * 
 * @example
 * ```typescript
 * import { accountTypes, accountTypeLabels } from '@/schemas'
 * 
 * // Generate options for select dropdown
 * const accountTypeOptions = accountTypes.map(type => ({
 *   value: type,
 *   label: accountTypeLabels[type]
 * }))
 * ```
 * 
 * @example
 * ```typescript
 * import { formatDateForAPI, getTodayDateString } from '@/schemas'
 * 
 * // Format date for API
 * const dateString = formatDateForAPI(new Date()) // "2025-12-08"
 * 
 * // Get today's date
 * const today = getTodayDateString() // "2025-12-08"
 * ```
 */
