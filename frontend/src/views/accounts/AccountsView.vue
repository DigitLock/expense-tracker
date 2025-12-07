<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { accountsApi } from '@/api'
import type { Account } from '@/types'
import { Wallet, Plus } from 'lucide-vue-next'

const loading = ref(true)
const accounts = ref<Account[]>([])

onMounted(async () => {
  try {
    const response = await accountsApi.list()
    if (response.success) {
      accounts.value = response.data.accounts
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
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Accounts</h1>
        <p class="text-gray-500 mt-1">Manage your financial accounts</p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-5 w-5" />
        Add Account
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="account in accounts"
        :key="account.id"
        class="bg-white rounded-xl shadow-sm border border-gray-200 p-6"
        :class="{ 'opacity-50': !account.is_active }"
      >
        <div class="flex items-start justify-between mb-4">
          <div class="p-2 bg-blue-100 rounded-lg">
            <Wallet class="h-5 w-5 text-blue-600" />
          </div>
          <span
            class="text-xs font-medium px-2 py-1 rounded-full capitalize"
            :class="{
              'bg-green-100 text-green-700': account.type === 'checking',
              'bg-yellow-100 text-yellow-700': account.type === 'cash',
              'bg-purple-100 text-purple-700': account.type === 'savings',
            }"
          >
            {{ account.type }}
          </span>
        </div>
        <h3 class="font-semibold text-gray-900">{{ account.name }}</h3>
        <p
          class="text-2xl font-bold mt-2"
          :class="parseFloat(account.current_balance) >= 0 ? 'text-gray-900' : 'text-red-600'"
        >
          {{ formatCurrency(account.current_balance, account.currency) }}
        </p>
        <p class="text-sm text-gray-500 mt-1">
          Initial: {{ formatCurrency(account.initial_balance, account.currency) }}
        </p>
      </div>
    </div>
  </div>
</template>
