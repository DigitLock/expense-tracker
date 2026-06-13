<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { reportsApi } from '@/api'
import type { MonthlySummaryReport, SpendingByCategoryReport } from '@/types'
import { BarChart3, PieChart, TrendingUp, TrendingDown } from 'lucide-vue-next'

const loading = ref(true)
const monthlyReport = ref<MonthlySummaryReport | null>(null)
const spendingReport = ref<SpendingByCategoryReport | null>(null)

const currentMonth = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
})

const monthStart = computed(() => {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`
})

const monthEnd = computed(() => {
  const now = new Date()
  const lastDay = new Date(now.getFullYear(), now.getMonth() + 1, 0).getDate()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${lastDay}`
})

onMounted(async () => {
  try {
    const [monthlyResponse, spendingResponse] = await Promise.all([
      reportsApi.monthlySummary(currentMonth.value),
      reportsApi.spendingByCategory(monthStart.value, monthEnd.value),
    ])

    if (monthlyResponse.success) {
      monthlyReport.value = monthlyResponse.data
    }
    if (spendingResponse.success) {
      spendingReport.value = spendingResponse.data
    }
  } finally {
    loading.value = false
  }
})

function formatCurrency(amount: string | number): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount
  return new Intl.NumberFormat('en-US', {
    style: 'decimal',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(num) + ' RSD'
}

function formatPercent(value: string): string {
  return parseFloat(value).toFixed(1) + '%'
}
</script>

<template>
  <div>
    <div class="mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Reports</h1>
      <p class="text-gray-500 mt-1">Analyze your financial data</p>
    </div>

    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else class="space-y-6">
      <!-- Monthly Summary -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div class="flex items-center gap-3 mb-6">
          <div class="p-2 bg-blue-100 rounded-lg">
            <BarChart3 class="h-5 w-5 text-blue-600" />
          </div>
          <h2 class="text-lg font-semibold text-gray-900">Monthly Summary</h2>
          <span class="text-sm text-gray-500">({{ currentMonth }})</span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="p-4 bg-green-50 rounded-lg">
            <div class="flex items-center gap-2 mb-2">
              <TrendingUp class="h-4 w-4 text-green-600" />
              <span class="text-sm text-green-700">Income</span>
            </div>
            <p class="text-xl font-bold text-green-700">
              {{ formatCurrency(monthlyReport?.summary.total_income || 0) }}
            </p>
          </div>

          <div class="p-4 bg-red-50 rounded-lg">
            <div class="flex items-center gap-2 mb-2">
              <TrendingDown class="h-4 w-4 text-red-600" />
              <span class="text-sm text-red-700">Expenses</span>
            </div>
            <p class="text-xl font-bold text-red-700">
              {{ formatCurrency(monthlyReport?.summary.total_expenses || 0) }}
            </p>
          </div>

          <div class="p-4 bg-blue-50 rounded-lg">
            <span class="text-sm text-blue-700">Net Savings</span>
            <p
              class="text-xl font-bold"
              :class="parseFloat(monthlyReport?.summary.net_savings || '0') >= 0 ? 'text-blue-700' : 'text-red-700'"
            >
              {{ formatCurrency(monthlyReport?.summary.net_savings || 0) }}
            </p>
          </div>

          <div class="p-4 bg-purple-50 rounded-lg">
            <span class="text-sm text-purple-700">Savings Rate</span>
            <p
              class="text-xl font-bold"
              :class="parseFloat(monthlyReport?.summary.savings_rate || '0') >= 0 ? 'text-purple-700' : 'text-red-700'"
            >
              {{ formatPercent(monthlyReport?.summary.savings_rate || '0') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Spending by Category -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
        <div class="flex items-center gap-3 mb-6">
          <div class="p-2 bg-orange-100 rounded-lg">
            <PieChart class="h-5 w-5 text-orange-600" />
          </div>
          <h2 class="text-lg font-semibold text-gray-900">Spending by Category</h2>
        </div>

        <div class="space-y-3">
          <div
            v-for="category in spendingReport?.spending_by_category"
            :key="category.category_id"
            class="flex items-center gap-4"
          >
            <div class="flex-1">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm font-medium text-gray-900">{{ category.category_name }}</span>
                <span class="text-sm text-gray-500">{{ formatPercent(category.percentage) }}</span>
              </div>
              <div class="h-2 bg-gray-100 rounded-full overflow-hidden">
                <div
                  class="h-full bg-blue-500 rounded-full"
                  :style="{ width: `${Math.min(parseFloat(category.percentage), 100)}%` }"
                ></div>
              </div>
            </div>
            <span class="text-sm font-medium text-gray-900 w-28 text-right">
              {{ formatCurrency(category.total_amount) }}
            </span>
          </div>

          <div
            v-if="!spendingReport?.spending_by_category?.length"
            class="text-center text-gray-500 py-4"
          >
            No spending data for this period
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
