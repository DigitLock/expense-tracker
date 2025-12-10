<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { transactionsApi, accountsApi, categoriesApi } from '@/api'
import type { Transaction, Account, Category, TransactionFilters } from '@/types'
import {
  ArrowUpRight,
  ArrowDownRight,
  Plus,
  Filter,
  ChevronLeft,
  ChevronRight,
} from 'lucide-vue-next'
import { useModal } from '@/composables/useModal'
import { useToast } from '@/composables/useToast'
import { formatDateForDisplay, formatAmount } from '@/schemas/transaction'
import CreateTransactionModal from '@/components/modals/CreateTransactionModal.vue'
import EditTransactionModal from '@/components/modals/EditTransactionModal.vue'
import DeleteTransactionModal from '@/components/modals/DeleteTransactionModal.vue'

const loading = ref(true)
const transactions = ref<Transaction[]>([])
const accounts = ref<Account[]>([])
const categories = ref<Category[]>([])
const pagination = ref({
  page: 1,
  per_page: 20,
  total: 0,
  total_pages: 0,
})

const createModal = useModal()
const editModal = useModal()
const deleteModal = ref(false)
const toast = useToast()
const selectedTransaction = ref<Transaction | null>(null)
const deleteLoading = ref(false)

// Filters
const activeTab = ref<'all' | 'expense' | 'income'>('all')
const selectedAccountId = ref<string>('')
const selectedMonth = ref<string>('')
const showFilters = ref(false)

// Load data
async function loadTransactions() {
  try {
    loading.value = true

    const filters: TransactionFilters = {
      page: pagination.value.page,
      per_page: pagination.value.per_page,
    }

    if (activeTab.value !== 'all') {
      filters.type = activeTab.value
    }

    if (selectedAccountId.value) {
      filters.account_id = selectedAccountId.value
    }

    if (selectedMonth.value) {
      filters.month = selectedMonth.value
    }

    const response = await transactionsApi.list(filters)
    if (response.success) {
      transactions.value = response.data.transactions
      pagination.value = response.data.pagination
    }
  } catch (err) {
    console.error('Failed to load transactions:', err)
    toast.error('Failed to load transactions')
  } finally {
    loading.value = false
  }
}

async function loadAccounts() {
  try {
    const response = await accountsApi.list()
    if (response.success) {
      accounts.value = response.data.accounts
    }
  } catch (err) {
    console.error('Failed to load accounts:', err)
  }
}

async function loadCategories() {
  try {
    const response = await categoriesApi.list()
    if (response.success) {
      categories.value = response.data.categories
    }
  } catch (err) {
    console.error('Failed to load categories:', err)
  }
}

onMounted(async () => {
  await Promise.all([
    loadTransactions(),
    loadAccounts(),
    loadCategories(),
  ])
})

// Watch filters
watch([activeTab, selectedAccountId, selectedMonth], () => {
  pagination.value.page = 1 // Reset to first page
  loadTransactions()
})

// Computed
const accountOptions = computed(() => {
  return accounts.value
      .filter(a => a.is_active)
      .map(a => ({
        value: a.id,
        label: `${a.name} (${a.currency})`,
      }))
})

const monthOptions = computed(() => {
  const months = []
  const currentDate = new Date()

  for (let i = 0; i < 12; i++) {
    const date = new Date(currentDate.getFullYear(), currentDate.getMonth() - i, 1)
    const value = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
    const label = new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'long',
    }).format(date)

    months.push({ value, label })
  }

  return months
})

const hasActiveFilters = computed(() => {
  return selectedAccountId.value !== '' || selectedMonth.value !== ''
})

// CRUD handlers
function openCreateModal() {
  createModal.open()
}

function openEditModal(transaction: Transaction) {
  selectedTransaction.value = transaction
  editModal.open()
}

function openDeleteModal() {
  editModal.close()
  deleteModal.value = true
}

async function handleTransactionCreated() {
  await loadTransactions()
  createModal.close()
}

async function handleTransactionUpdated() {
  await loadTransactions()
  editModal.close()
}

async function handleDeleteConfirm() {
  if (!selectedTransaction.value) return

  try {
    deleteLoading.value = true
    const response = await transactionsApi.delete(selectedTransaction.value.id)

    if (response.success) {
      toast.success('Transaction deleted successfully!')
      deleteModal.value = false
      selectedTransaction.value = null
      await loadTransactions()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to delete transaction'
    toast.error(errorMessage)
  } finally {
    deleteLoading.value = false
  }
}

// Pagination
function goToPage(page: number) {
  pagination.value.page = page
  loadTransactions()
}

function nextPage() {
  if (pagination.value.page < pagination.value.total_pages) {
    goToPage(pagination.value.page + 1)
  }
}

function prevPage() {
  if (pagination.value.page > 1) {
    goToPage(pagination.value.page - 1)
  }
}

// Filters
function clearFilters() {
  selectedAccountId.value = ''
  selectedMonth.value = ''
  pagination.value.page = 1
}
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Transactions</h1>
        <p class="text-gray-500 mt-1">Track your income and expenses</p>
      </div>
      <button
          @click="openCreateModal"
          class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-5 w-5" />
        Add Transaction
      </button>
    </div>

    <!-- Tabs & Filters -->
    <div class="flex flex-col sm:flex-row gap-4 mb-6">
      <!-- Type Tabs -->
      <div class="flex gap-2">
        <button
            @click="activeTab = 'all'"
            class="px-4 py-2 rounded-lg font-medium transition-colors"
            :class="activeTab === 'all'
            ? 'bg-blue-100 text-blue-700'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
        >
          All
        </button>
        <button
            @click="activeTab = 'expense'"
            class="px-4 py-2 rounded-lg font-medium transition-colors"
            :class="activeTab === 'expense'
            ? 'bg-red-100 text-red-700'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
        >
          <ArrowDownRight class="h-4 w-4 inline mr-2" />
          Expenses
        </button>
        <button
            @click="activeTab = 'income'"
            class="px-4 py-2 rounded-lg font-medium transition-colors"
            :class="activeTab === 'income'
            ? 'bg-green-100 text-green-700'
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
        >
          <ArrowUpRight class="h-4 w-4 inline mr-2" />
          Income
        </button>
      </div>

      <!-- Filter Toggle -->
      <button
          @click="showFilters = !showFilters"
          class="flex items-center gap-2 px-4 py-2 rounded-lg font-medium transition-colors ml-auto"
          :class="hasActiveFilters || showFilters
          ? 'bg-blue-100 text-blue-700'
          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      >
        <Filter class="h-4 w-4" />
        Filters
        <span v-if="hasActiveFilters" class="px-2 py-0.5 bg-blue-600 text-white text-xs rounded-full">
          {{ (selectedAccountId ? 1 : 0) + (selectedMonth ? 1 : 0) }}
        </span>
      </button>
    </div>

    <!-- Filter Panel -->
    <div v-if="showFilters" class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- Account Filter -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Account</label>
          <select
              v-model="selectedAccountId"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">All Accounts</option>
            <option v-for="option in accountOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>

        <!-- Month Filter -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">Month</label>
          <select
              v-model="selectedMonth"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">All Time</option>
            <option v-for="option in monthOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
      </div>

      <!-- Clear Filters -->
      <div v-if="hasActiveFilters" class="mt-4">
        <button
            @click="clearFilters"
            class="text-sm text-blue-600 hover:text-blue-700 font-medium"
        >
          Clear all filters
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Transactions List -->
    <div v-else-if="transactions.length > 0" class="space-y-4">
      <div
          v-for="transaction in transactions"
          :key="transaction.id"
          @click="openEditModal(transaction)"
          class="bg-white rounded-xl shadow-sm border border-gray-200 p-4 cursor-pointer hover:border-blue-300 transition-colors"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center gap-3">
              <div
                  class="p-2 rounded-lg"
                  :class="transaction.type === 'income' ? 'bg-green-100' : 'bg-red-100'"
              >
                <component
                    :is="transaction.type === 'income' ? ArrowUpRight : ArrowDownRight"
                    class="h-4 w-4"
                    :class="transaction.type === 'income' ? 'text-green-600' : 'text-red-600'"
                />
              </div>
              <div>
                <h3 class="font-medium text-gray-900">{{ transaction.category.name }}</h3>
                <p class="text-sm text-gray-500">{{ transaction.account.name }}</p>
              </div>
            </div>

            <p v-if="transaction.description" class="text-sm text-gray-600 mt-2 ml-11">
              {{ transaction.description }}
            </p>

            <div class="flex items-center gap-4 mt-2 ml-11 text-xs text-gray-400">
              <span>{{ formatDateForDisplay(transaction.date) }}</span>
              <span>by {{ transaction.created_by }}</span>
            </div>
          </div>

          <div class="text-right">
            <p
                class="text-lg font-semibold"
                :class="transaction.type === 'income' ? 'text-green-600' : 'text-red-600'"
            >
              {{ transaction.type === 'income' ? '+' : '-' }}{{ formatAmount(transaction.amount, transaction.currency) }}
            </p>
            <p class="text-xs text-gray-500 mt-1">
              {{ transaction.currency }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="bg-white rounded-xl shadow-sm border border-gray-200 p-8 text-center">
      <p class="text-gray-500">No transactions found</p>
      <button
          @click="clearFilters"
          v-if="hasActiveFilters"
          class="mt-4 text-blue-600 hover:text-blue-700 font-medium"
      >
        Clear filters
      </button>
    </div>

    <!-- Pagination -->
    <div v-if="pagination.total_pages > 1" class="flex items-center justify-between mt-6">
      <p class="text-sm text-gray-600">
        Showing {{ (pagination.page - 1) * pagination.per_page + 1 }} to
        {{ Math.min(pagination.page * pagination.per_page, pagination.total) }} of
        {{ pagination.total }} transactions
      </p>

      <div class="flex items-center gap-2">
        <button
            @click="prevPage"
            :disabled="pagination.page === 1"
            class="p-2 rounded-lg border border-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <ChevronLeft class="h-5 w-5" />
        </button>

        <div class="flex gap-1">
          <button
              v-for="page in Math.min(pagination.total_pages, 5)"
              :key="page"
              @click="goToPage(page)"
              class="px-3 py-1 rounded-lg transition-colors"
              :class="pagination.page === page
              ? 'bg-blue-600 text-white'
              : 'hover:bg-gray-100 text-gray-700'"
          >
            {{ page }}
          </button>
        </div>

        <button
            @click="nextPage"
            :disabled="pagination.page === pagination.total_pages"
            class="p-2 rounded-lg border border-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <ChevronRight class="h-5 w-5" />
        </button>
      </div>
    </div>
  </div>

  <!-- Modals -->
  <CreateTransactionModal
      v-model:open="createModal.isOpen.value"
      :accounts="accounts"
      :categories="categories"
      @created="handleTransactionCreated"
  />

  <EditTransactionModal
      v-if="selectedTransaction"
      v-model:open="editModal.isOpen.value"
      :transaction="selectedTransaction"
      :accounts="accounts"
      :categories="categories"
      @updated="handleTransactionUpdated"
      @delete="openDeleteModal"
  />

  <DeleteTransactionModal
      v-if="selectedTransaction"
      v-model:open="deleteModal"
      :transaction-description="`${selectedTransaction.category.name} - ${formatAmount(selectedTransaction.amount, selectedTransaction.currency)}`"
      :loading="deleteLoading"
      @confirm="handleDeleteConfirm"
  />
</template>