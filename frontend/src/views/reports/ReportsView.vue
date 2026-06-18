<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { reportsApi } from '@/api'
import type { MonthlySummaryReport, SpendingByCategoryReport, Currency } from '@/types'
import { useToast } from '@/composables/useToast'
import { BarChart3, PieChart, TrendingUp, TrendingDown } from 'lucide-vue-next'

const toast = useToast()

const loading = ref(true)
const monthlyReport = ref<MonthlySummaryReport | null>(null)
const spendingReport = ref<SpendingByCategoryReport | null>(null)

// Controls
function defaultMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}
const selectedMonth = ref(defaultMonth()) // YYYY-MM (input type="month")
const selectedType = ref<'expense' | 'income'>('expense')
const selectedCurrency = ref<Currency>('RSD')

const currencyOptions: Currency[] = ['RSD', 'EUR', 'USD']

// Derive the spending-by-category date range from the selected month.
const monthStart = computed(() => `${selectedMonth.value}-01`)
const monthEnd = computed(() => {
  const parts = selectedMonth.value.split('-')
  const year = Number(parts[0])
  const month = Number(parts[1])
  const lastDay = new Date(year, month, 0).getDate()
  return `${selectedMonth.value}-${String(lastDay).padStart(2, '0')}`
})

async function loadReports() {
  loading.value = true
  try {
    const [monthlyResponse, spendingResponse] = await Promise.all([
      reportsApi.monthlySummary(selectedMonth.value, selectedCurrency.value),
      reportsApi.spendingByCategory(
        monthStart.value,
        monthEnd.value,
        selectedType.value,
        selectedCurrency.value
      ),
    ])

    if (monthlyResponse.success) monthlyReport.value = monthlyResponse.data
    if (spendingResponse.success) spendingReport.value = spendingResponse.data

    // Surface a currency fallback note (no rate available) once per load.
    const note = monthlyResponse.data?.currency_note || spendingResponse.data?.currency_note
    if (note) toast.info(note)
  } finally {
    loading.value = false
  }
}

onMounted(loadReports)

// Reload both reports when month or currency changes.
watch([selectedMonth, selectedCurrency], loadReports)

// Type only affects spending-by-category; reload just that report.
watch(selectedType, async () => {
  const response = await reportsApi.spendingByCategory(
    monthStart.value,
    monthEnd.value,
    selectedType.value,
    selectedCurrency.value
  )
  if (response.success) spendingReport.value = response.data
})

// Currency formatting driven by the response currency, not hardcoded RSD.
function symbolFor(currency?: string): { prefix: string; suffix: string } {
  switch (currency) {
    case 'EUR':
      return { prefix: '€', suffix: '' }
    case 'USD':
      return { prefix: '$', suffix: '' }
    default:
      return { prefix: '', suffix: ' RSD' }
  }
}

function formatCurrency(amount: string | number, currency?: string): string {
  const num = typeof amount === 'string' ? parseFloat(amount) : amount
  const { prefix, suffix } = symbolFor(currency)
  const formatted = new Intl.NumberFormat('en-US', {
    style: 'decimal',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(num)
  return `${prefix}${formatted}${suffix}`
}

// Currency actually used in each report (may differ from selected on fallback).
const monthlyCurrency = computed(() => monthlyReport.value?.currency)
const spendingCurrency = computed(() => spendingReport.value?.currency)

function formatPercent(value: string): string {
  return parseFloat(value).toFixed(1) + '%'
}
</script>

<template>
  <div>
    <div class="mb-8 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Reports</h1>
        <p class="text-gray-500 mt-1">Analyze your financial data</p>
      </div>

      <!-- Controls -->
      <div class="flex flex-wrap items-center gap-3">
        <!-- Month selector -->
        <div>
          <label for="report-month" class="sr-only">Month</label>
          <input
            id="report-month"
            v-model="selectedMonth"
            type="month"
            class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <!-- Currency switcher -->
        <div>
          <label for="report-currency" class="sr-only">Currency</label>
          <select
            id="report-currency"
            v-model="selectedCurrency"
            class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option v-for="c in currencyOptions" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>
      </div>
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
          <span class="text-sm text-gray-500">({{ selectedMonth }})</span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="p-4 bg-green-50 rounded-lg">
            <div class="flex items-center gap-2 mb-2">
              <TrendingUp class="h-4 w-4 text-green-600" />
              <span class="text-sm text-green-700">Income</span>
            </div>
            <p class="text-xl font-bold text-green-700">
              {{ formatCurrency(monthlyReport?.summary.total_income || 0, monthlyCurrency) }}
            </p>
          </div>

          <div class="p-4 bg-red-50 rounded-lg">
            <div class="flex items-center gap-2 mb-2">
              <TrendingDown class="h-4 w-4 text-red-600" />
              <span class="text-sm text-red-700">Expenses</span>
            </div>
            <p class="text-xl font-bold text-red-700">
              {{ formatCurrency(monthlyReport?.summary.total_expenses || 0, monthlyCurrency) }}
            </p>
          </div>

          <div class="p-4 bg-blue-50 rounded-lg">
            <span class="text-sm text-blue-700">Net Savings</span>
            <p
              class="text-xl font-bold"
              :class="parseFloat(monthlyReport?.summary.net_savings || '0') >= 0 ? 'text-blue-700' : 'text-red-700'"
            >
              {{ formatCurrency(monthlyReport?.summary.net_savings || 0, monthlyCurrency) }}
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
        <div class="flex items-center justify-between gap-3 mb-6">
          <div class="flex items-center gap-3">
            <div class="p-2 bg-orange-100 rounded-lg">
              <PieChart class="h-5 w-5 text-orange-600" />
            </div>
            <h2 class="text-lg font-semibold text-gray-900">
              {{ selectedType === 'expense' ? 'Spending' : 'Income' }} by Category
            </h2>
          </div>

          <!-- Transaction type toggle -->
          <div class="inline-flex rounded-lg border border-gray-300 p-0.5 bg-gray-50">
            <button
              type="button"
              @click="selectedType = 'expense'"
              :class="[
                'px-3 py-1.5 text-sm font-medium rounded-md transition-colors',
                selectedType === 'expense' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700',
              ]"
            >
              Expense
            </button>
            <button
              type="button"
              @click="selectedType = 'income'"
              :class="[
                'px-3 py-1.5 text-sm font-medium rounded-md transition-colors',
                selectedType === 'income' ? 'bg-white shadow-sm text-gray-900' : 'text-gray-500 hover:text-gray-700',
              ]"
            >
              Income
            </button>
          </div>
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
              {{ formatCurrency(category.total_amount, spendingCurrency) }}
            </span>
          </div>

          <div
            v-if="!spendingReport?.spending_by_category?.length"
            class="text-center text-gray-500 py-4"
          >
            No {{ selectedType }} data for this period
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
