<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { transactionsApi } from '@/api'
import type { Transaction, PaginationMeta } from '@/types'
import { Plus, ArrowUpRight, ArrowDownRight } from 'lucide-vue-next'

const loading = ref(true)
const transactions = ref<Transaction[]>([])
const pagination = ref<PaginationMeta | null>(null)

onMounted(async () => {
  try {
    const response = await transactionsApi.list({ per_page: 20 })
    if (response.success) {
      transactions.value = response.data.transactions
      pagination.value = response.data.pagination
    }
  } finally {
    loading.value = false
  }
})

function formatCurrency(amount: string, currency: string): string {
  const num = parseFloat(amount)
  return new Intl.NumberFormat('en-US', {
    style: 'decimal',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(num) + ' ' + currency
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Transactions</h1>
        <p class="text-gray-500 mt-1">Track your income and expenses</p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-5 w-5" />
        Add Transaction
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      <table class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
            <th class="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Description</th>
            <th class="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Category</th>
            <th class="text-left px-6 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Account</th>
            <th class="text-right px-6 py-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="tx in transactions" :key="tx.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(tx.date) }}
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center gap-3">
                <div
                  class="p-1.5 rounded-lg"
                  :class="tx.type === 'income' ? 'bg-green-100' : 'bg-red-100'"
                >
                  <ArrowUpRight v-if="tx.type === 'income'" class="h-4 w-4 text-green-600" />
                  <ArrowDownRight v-else class="h-4 w-4 text-red-600" />
                </div>
                <span class="text-sm text-gray-900">{{ tx.description || '-' }}</span>
              </div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ tx.category?.name || '-' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ tx.account?.name || '-' }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right">
              <span
                class="text-sm font-medium"
                :class="tx.type === 'income' ? 'text-green-600' : 'text-red-600'"
              >
                {{ tx.type === 'income' ? '+' : '-' }}{{ formatCurrency(tx.amount, tx.currency) }}
              </span>
            </td>
          </tr>
          <tr v-if="transactions.length === 0">
            <td colspan="5" class="px-6 py-8 text-center text-gray-500">
              No transactions yet
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination info -->
      <div v-if="pagination" class="px-6 py-3 bg-gray-50 border-t border-gray-200 text-sm text-gray-500">
        Showing {{ transactions.length }} of {{ pagination.total }} transactions
      </div>
    </div>
  </div>
</template>
