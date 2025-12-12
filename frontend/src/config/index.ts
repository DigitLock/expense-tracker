export const config = {
  api: {
    baseUrl: import.meta.env.VITE_API_BASE_URL || '/api/v1',
    timeout: 10000,
  },
  auth: {
    tokenKey: 'expense_tracker_token',
    userKey: 'expense_tracker_user',
  },
  app: {
    name: 'Expense Tracker',
    version: '1.0.0',
    defaultCurrency: 'RSD' as const,
    supportedCurrencies: ['RSD', 'EUR', 'USD'] as const,
  },
  pagination: {
    defaultPageSize: 20,
    maxPageSize: 100,
  },
}

export type Config = typeof config
