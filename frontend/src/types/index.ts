// ============================================
// API Response Types
// ============================================

export interface ApiResponse<T> {
  success: boolean
  data: T
}

export interface ApiError {
  success: false
  error: {
    code: string
    message: string
    details?: ValidationError[]
  }
}

export interface ValidationError {
  field: string
  message: string
}

export interface PaginationMeta {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface PaginatedResponse<T> {
  success: boolean
  data: {
    transactions: T[]
    pagination: PaginationMeta
  }
}

// ============================================
// Auth Types
// ============================================

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
  family_name?: string
}

export interface LoginResponse {
  token: string
  user: User
  expires_in: number
}

export interface User {
  id: string
  email: string
  name: string
  family_id: string
  role?: string
}

// ============================================
// Account Types
// ============================================

export type AccountType = 'cash' | 'bank' | 'credit' | 'savings' | 'investment'
export type Currency = 'RSD' | 'EUR' | 'USD'

export interface Account {
  id: string
  name: string
  type: AccountType
  currency: Currency
  initial_balance: string
  current_balance: string
  is_active: boolean
  created_at: string
  updated_at: string
  description?: string
}

export interface CreateAccountRequest {
  name: string
  type: AccountType
  currency: Currency
  initial_balance: number
  description?: string
}

export interface UpdateAccountRequest {
  name?: string
  is_active?: boolean
  description?: string
}

export interface AccountBalance {
  account_id: string
  account_name: string
  current_balance: string
  currency: Currency
}

// ============================================
// Category Types
// ============================================

export type CategoryType = 'income' | 'expense'

export interface Category {
  id: string
  name: string
  type: CategoryType
  parent_id: string | null
  parent_name: string | null
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateCategoryRequest {
  name: string
  type: CategoryType
  parent_id?: string
}

export interface UpdateCategoryRequest {
  name?: string
  parent_id?: string | null
  is_active?: boolean
}

// ============================================
// Transaction Types
// ============================================

export type TransactionType = 'income' | 'expense'

export interface Transaction {
  id: string
  type: TransactionType
  amount: string
  currency: Currency
  amount_base: string
  base_currency: Currency
  category: {
    id: string
    name: string
    type: CategoryType
  }
  account: {
    id: string
    name: string
    type: AccountType
  }
  description: string | null
  date: string
  created_at: string
  created_by: string
}

export interface CreateTransactionRequest {
  type: TransactionType
  amount: number
  currency: Currency
  account_id: string
  category_id: string
  description?: string
  date: string
}

export interface UpdateTransactionRequest {
  type?: TransactionType
  amount?: number
  currency?: Currency
  account_id?: string
  category_id?: string
  description?: string
  date?: string
}

export interface TransactionFilters {
  type?: TransactionType
  account_id?: string
  month?: string
  page?: number
  per_page?: number
}

// ============================================
// Report Types
// ============================================

export interface SpendingByCategory {
  category_id: string
  category_name: string
  total_amount: string
  transaction_count: number
  percentage: string
  average_per_transaction: string
}

export interface SpendingByCategoryReport {
  report_type: 'spending_by_category'
  period: {
    start_date: string
    end_date: string
  }
  currency: Currency
  transaction_type: string
  spending_by_category: SpendingByCategory[]
  total_amount: string
  total_transactions: number
  generated_at: string
  currency_note?: string
}

export interface MonthlySummary {
  total_income: string
  total_expenses: string
  net_savings: string
  savings_rate: string
}

export interface MonthlySummaryReport {
  report_type: 'monthly_summary'
  month: string
  currency: Currency
  summary: MonthlySummary
  income_breakdown: Record<string, string>
  expense_breakdown: Record<string, string>
  account_balances: {
    accounts: Record<string, string>
    total: string
  }
  currency_note?: string
}

// ============================================
// Currency Types
// ============================================

export interface ExchangeRate {
  from_currency: Currency
  to_currency: Currency
  rate: string
  date: string
}

export interface CurrencyConversion {
  from_currency: Currency
  to_currency: Currency
  from_amount: string
  to_amount: string
  rate: string
  rate_date: string
}
