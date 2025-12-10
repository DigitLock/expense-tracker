<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { reportsApi, accountsApi, categoriesApi } from '@/api'
import type { MonthlySummaryReport, Account, Category } from '@/types'
import {
  TrendingUp,
  TrendingDown,
  Wallet,
  PiggyBank,
  ArrowDownRight,
  Plus,
} from 'lucide-vue-next'
import { useModal } from '@/composables/useModal'
import CreateTransactionModal from '@/components/modals/CreateTransactionModal.vue'

const loading = ref(true)
const error = ref<string | null>(null)
const monthlyReport = ref<MonthlySummaryReport | null>(null)
const accounts = ref<Account[]>([])
const categories = ref<Category[]>([])
const createTransactionModal = useModal()

// Current month in YYYY-MM format
const currentMonth = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
})

async function loadData() {
  try {
    loading.value = true
    const [reportResponse, accountsResponse, categoriesResponse] = await Promise.all([
      reportsApi.monthlySummary(currentMonth.value),
      accountsApi.list(),
      categoriesApi.list(),
    ])

    if (reportResponse.success) {
      monthlyReport.value = reportResponse.data
    }
    if (accountsResponse.success) {
      accounts.value = accountsResponse.data.accounts.filter((a: Account) => a.is_active)
    }
    if (categoriesResponse.success) {
      categories.value = categoriesResponse.data.categories
    }
  } catch (err: unknown) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    error.value = axiosError.response?.data?.error?.message || 'Failed to load dashboard data'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadData()
})

function formatCurrency(amount: string | number, currency = 'RSD'): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount
  return new Intl.NumberFormat('en-US', {
    style: 'decimal',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(num) + ' ' + currency
}

async function handleTransactionCreated() {
  await loadData()
  createTransactionModal.close()
}
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p class="text-gray-500 mt-1">Overview of your finances</p>
      </div>

      <button
          @click="createTransactionModal.open"
          class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-5 w-5" />
        Add Transaction
      </button>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4">
      <p class="text-red-700">{{ error }}</p>
    </div>

    <!-- Dashboard content -->
    <div v-else>
      <!-- Summary cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <!-- Total Income -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500">Income</p>
              <p class="text-2xl font-bold text-gray-900 mt-1">
                {{ formatCurrency(monthlyReport?.summary.total_income || 0) }}
              </p>
            </div>
            <div class="p-3 bg-green-100 rounded-lg">
              <TrendingUp class="h-6 w-6 text-green-600" />
            </div>
          </div>
        </div>

        <!-- Total Expenses -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500">Expenses</p>
              <p class="text-2xl font-bold text-gray-900 mt-1">
                {{ formatCurrency(monthlyReport?.summary.total_expenses || 0) }}
              </p>
            </div>
            <div class="p-3 bg-red-100 rounded-lg">
              <TrendingDown class="h-6 w-6 text-red-600" />
            </div>
          </div>
        </div>

        <!-- Net Savings -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500">Net Savings</p>
              <p
                  class="text-2xl font-bold mt-1"
                  :class="parseFloat(monthlyReport?.summary.net_savings || '0') >= 0 ? 'text-green-600' : 'text-red-600'"
              >
                {{ formatCurrency(monthlyReport?.summary.net_savings || 0) }}
              </p>
            </div>
            <div class="p-3 bg-blue-100 rounded-lg">
              <PiggyBank class="h-6 w-6 text-blue-600" />
            </div>
          </div>
        </div>

        <!-- Total Balance -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-500">Total Balance</p>
              <p class="text-2xl font-bold text-gray-900 mt-1">
                {{ formatCurrency(monthlyReport?.account_balances.total || 0) }}
              </p>
            </div>
            <div class="p-3 bg-purple-100 rounded-lg">
              <Wallet class="h-6 w-6 text-purple-600" />
            </div>
          </div>
        </div>
      </div>

      <!-- Two columns layout -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Accounts -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">Accounts</h2>
          <div class="space-y-3">
            <div
                v-for="account in accounts"
                :key="account.id"
                class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div>
                <p class="font-medium text-gray-900">{{ account.name }}</p>
                <p class="text-sm text-gray-500 capitalize">{{ account.type }}</p>
              </div>
              <p
                  class="font-semibold"
                  :class="parseFloat(account.current_balance) >= 0 ? 'text-gray-900' : 'text-red-600'"
              >
                {{ formatCurrency(account.current_balance, account.currency) }}
              </p>
            </div>
            <div v-if="accounts.length === 0" class="text-center text-gray-500 py-4">
              No accounts yet
            </div>
          </div>
        </div>

        <!-- Expense breakdown -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <h2 class="text-lg font-semibold text-gray-900 mb-4">Expenses by Category</h2>
          <div class="space-y-3">
            <div
                v-for="(amount, category) in monthlyReport?.expense_breakdown"
                :key="category"
                class="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div class="flex items-center gap-3">
                <ArrowDownRight class="h-4 w-4 text-red-500" />
                <p class="font-medium text-gray-900">{{ category }}</p>
              </div>
              <p class="font-semibold text-gray-900">
                {{ formatCurrency(amount) }}
              </p>
            </div>
            <div
                v-if="!monthlyReport?.expense_breakdown || Object.keys(monthlyReport.expense_breakdown).length === 0"
                class="text-center text-gray-500 py-4"
            >
              No expenses this month
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Transaction Modal -->
    <CreateTransactionModal
        v-model:open="createTransactionModal.isOpen.value"
        :accounts="accounts"
        :categories="categories"
        @created="handleTransactionCreated"
    />
  </div>
</template>