import { apiClient } from './client'
import type {
  ApiResponse,
  LoginRequest,
  LoginResponse,
  Account,
  CreateAccountRequest,
  UpdateAccountRequest,
  AccountBalance,
  Category,
  CreateCategoryRequest,
  UpdateCategoryRequest,
  Transaction,
  CreateTransactionRequest,
  UpdateTransactionRequest,
  TransactionFilters,
  SpendingByCategoryReport,
  MonthlySummaryReport,
  ExchangeRate,
  CurrencyConversion,
  PaginationMeta,
} from '@/types'

// ============================================
// Auth API
// ============================================

export const authApi = {
  login: async (data: LoginRequest): Promise<ApiResponse<LoginResponse>> => {
    const response = await apiClient.post<ApiResponse<LoginResponse>>('/auth/login', data)
    return response.data
  },
}

// ============================================
// Accounts API
// ============================================

interface AccountsListResponse {
  accounts: Account[]
}

export const accountsApi = {
  list: async (): Promise<ApiResponse<AccountsListResponse>> => {
    const response = await apiClient.get<ApiResponse<AccountsListResponse>>('/accounts')
    return response.data
  },

  get: async (id: string): Promise<ApiResponse<Account>> => {
    const response = await apiClient.get<ApiResponse<Account>>(`/accounts/${id}`)
    return response.data
  },

  create: async (data: CreateAccountRequest): Promise<ApiResponse<Account>> => {
    const response = await apiClient.post<ApiResponse<Account>>('/accounts', data)
    return response.data
  },

  update: async (id: string, data: UpdateAccountRequest): Promise<ApiResponse<Account>> => {
    const response = await apiClient.patch<ApiResponse<Account>>(`/accounts/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<ApiResponse<{ message: string }>> => {
    const response = await apiClient.delete<ApiResponse<{ message: string }>>(`/accounts/${id}`)
    return response.data
  },

  getBalance: async (id: string): Promise<ApiResponse<AccountBalance>> => {
    const response = await apiClient.get<ApiResponse<AccountBalance>>(`/accounts/${id}/balance`)
    return response.data
  },
}

// ============================================
// Categories API
// ============================================

interface CategoriesListResponse {
  categories: Category[]
}

export const categoriesApi = {
  list: async (type?: 'income' | 'expense'): Promise<ApiResponse<CategoriesListResponse>> => {
    const params = type ? { type } : {}
    const response = await apiClient.get<ApiResponse<CategoriesListResponse>>('/categories', { params })
    return response.data
  },

  get: async (id: string): Promise<ApiResponse<Category>> => {
    const response = await apiClient.get<ApiResponse<Category>>(`/categories/${id}`)
    return response.data
  },

  create: async (data: CreateCategoryRequest): Promise<ApiResponse<Category>> => {
    const response = await apiClient.post<ApiResponse<Category>>('/categories', data)
    return response.data
  },

  update: async (id: string, data: UpdateCategoryRequest): Promise<ApiResponse<Category>> => {
    const response = await apiClient.patch<ApiResponse<Category>>(`/categories/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<ApiResponse<{ message: string }>> => {
    const response = await apiClient.delete<ApiResponse<{ message: string }>>(`/categories/${id}`)
    return response.data
  },
}

// ============================================
// Transactions API
// ============================================

interface TransactionsListResponse {
  transactions: Transaction[]
  pagination: PaginationMeta
}

export const transactionsApi = {
  list: async (filters?: TransactionFilters): Promise<ApiResponse<TransactionsListResponse>> => {
    const response = await apiClient.get<ApiResponse<TransactionsListResponse>>('/transactions', {
      params: filters,
    })
    return response.data
  },

  get: async (id: string): Promise<ApiResponse<Transaction>> => {
    const response = await apiClient.get<ApiResponse<Transaction>>(`/transactions/${id}`)
    return response.data
  },

  create: async (data: CreateTransactionRequest): Promise<ApiResponse<Transaction>> => {
    const response = await apiClient.post<ApiResponse<Transaction>>('/transactions', data)
    return response.data
  },

  update: async (id: string, data: UpdateTransactionRequest): Promise<ApiResponse<Transaction>> => {
    const response = await apiClient.patch<ApiResponse<Transaction>>(`/transactions/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<ApiResponse<{ message: string }>> => {
    const response = await apiClient.delete<ApiResponse<{ message: string }>>(`/transactions/${id}`)
    return response.data
  },
}

// ============================================
// Reports API
// ============================================

export const reportsApi = {
  spendingByCategory: async (
    startDate: string,
    endDate: string
  ): Promise<ApiResponse<SpendingByCategoryReport>> => {
    const response = await apiClient.get<ApiResponse<SpendingByCategoryReport>>(
      '/reports/spending-by-category',
      { params: { start_date: startDate, end_date: endDate } }
    )
    return response.data
  },

  monthlySummary: async (month?: string): Promise<ApiResponse<MonthlySummaryReport>> => {
    const params = month ? { month } : {}
    const response = await apiClient.get<ApiResponse<MonthlySummaryReport>>(
      '/reports/monthly-summary',
      { params }
    )
    return response.data
  },
}

// ============================================
// Currencies API
// ============================================

export const currenciesApi = {
  getRates: async (): Promise<ApiResponse<ExchangeRate[]>> => {
    const response = await apiClient.get<ApiResponse<ExchangeRate[]>>('/currencies/rates')
    return response.data
  },

  convert: async (
    amount: number,
    from: string,
    to: string
  ): Promise<ApiResponse<CurrencyConversion>> => {
    const response = await apiClient.get<ApiResponse<CurrencyConversion>>('/currencies/convert', {
      params: { amount, from, to },
    })
    return response.data
  },
}
