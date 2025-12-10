<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { accountsApi } from '@/api'
import type { Account } from '@/types'
import { Wallet, Plus, Trash2 } from 'lucide-vue-next'
import { useModal } from '@/composables/useModal'
import { useToast } from '@/composables/useToast'
import { Form } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import {
  createAccountSchema,
  updateAccountSchema,
  accountTypes,
  accountTypeLabels,
  currencyCodes,
} from '@/schemas'
import {
  DialogRoot,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from '@/components/ui'
import { FormInput, FormSelect } from '@/components/forms'
import DeleteAccountModal from '@/components/modals/DeleteAccountModal.vue'

const loading = ref(true)
const accounts = ref<Account[]>([])
const createModal = useModal()
const editModal = useModal()
const deleteModal = ref(false)
const toast = useToast()

// Currently selected account for edit/delete
const selectedAccount = ref<Account | null>(null)
const deleteLoading = ref(false)

// Options for selects
const accountTypeOptions = accountTypes.map(t => ({
  value: t,
  label: accountTypeLabels[t],
}))

const currencyOptions = currencyCodes.map(c => ({
  value: c,
  label: c,
}))

// Load accounts
async function loadAccounts() {
  try {
    loading.value = true
    const response = await accountsApi.list()
    if (response.success) {
      accounts.value = response.data.accounts
    }
  } catch (err) {
    console.error('Failed to load accounts:', err)
  } finally {
    loading.value = false
  }
}

// CREATE handler
const onCreateSubmit = async (values: any) => {
  try {
    const response = await accountsApi.create({
      name: values.name,
      type: values.type,
      currency: values.currency,
      initial_balance: values.initial_balance,
      description: values.description || undefined,
    })

    if (response.success) {
      toast.success('Account created successfully!')
      createModal.close()
      await loadAccounts()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to create account'
    toast.error(errorMessage)
  }
}

// EDIT handlers
function openEditModal(account: Account) {
  selectedAccount.value = account
  editModal.open()
}

const onEditSubmit = async (values: any) => {
  if (!selectedAccount.value) return

  try {
    const response = await accountsApi.update(selectedAccount.value.id, {
      name: values.name,
      description: values.description,
      is_active: values.is_active,
    })

    if (response.success) {
      toast.success('Account updated successfully!')
      editModal.close()
      await loadAccounts()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to update account'
    toast.error(errorMessage)
  }
}

// DELETE handlers
function openDeleteModal() {
  editModal.close()
  deleteModal.value = true
}

async function handleDeleteConfirm() {
  if (!selectedAccount.value) return

  try {
    deleteLoading.value = true
    const response = await accountsApi.delete(selectedAccount.value.id)

    if (response.success) {
      toast.success('Account deleted successfully!')
      deleteModal.value = false
      selectedAccount.value = null
      await loadAccounts()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to delete account'
    toast.error(errorMessage)
  } finally {
    deleteLoading.value = false
  }
}

onMounted(async () => {
  await loadAccounts()
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
          @click="createModal.open"
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
          @click="openEditModal(account)"
          class="bg-white rounded-xl shadow-sm border border-gray-200 p-6 cursor-pointer hover:shadow-md hover:border-blue-300 transition-all"
          :class="{ 'opacity-50': !account.is_active }"
      >
        <div class="flex items-start justify-between mb-4">
          <div class="p-2 bg-blue-100 rounded-lg">
            <Wallet class="h-5 w-5 text-blue-600" />
          </div>
          <span
              class="text-xs font-medium px-2 py-1 rounded-full capitalize"
              :class="{
              'bg-green-100 text-green-700': account.type === 'bank',
              'bg-yellow-100 text-yellow-700': account.type === 'cash',
              'bg-purple-100 text-purple-700': account.type === 'savings',
              'bg-blue-100 text-blue-700': account.type === 'credit',
              'bg-indigo-100 text-indigo-700': account.type === 'investment',
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
        <p v-if="account.description" class="text-xs text-gray-400 mt-2 line-clamp-1">
          {{ account.description }}
        </p>
      </div>
    </div>
  </div>

  <!-- Create Account Modal -->
  <DialogRoot v-model:open="createModal.isOpen.value">
    <DialogContent class="sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle>Create New Account</DialogTitle>
        <DialogDescription>
          Add a new account to track your finances
        </DialogDescription>
      </DialogHeader>

      <Form
          @submit="onCreateSubmit"
          :validation-schema="toTypedSchema(createAccountSchema)"
          :initial-values="{
          name: '',
          type: 'cash',
          currency: 'RSD',
          initial_balance: 0,
          description: '',
        }"
          class="space-y-4 py-4"
      >
        <FormInput
            name="name"
            label="Account Name"
            placeholder="e.g., My Wallet"
            required
        />

        <FormSelect
            name="type"
            label="Account Type"
            :options="accountTypeOptions"
            placeholder="Select type"
            required
        />

        <FormSelect
            name="currency"
            label="Currency"
            :options="currencyOptions"
            placeholder="Select currency"
            required
        />

        <FormInput
            name="initial_balance"
            label="Initial Balance"
            type="number"
            step="0.01"
            placeholder="0.00"
        />

        <FormInput
            name="description"
            label="Description"
            placeholder="Add notes about this account"
        />

        <DialogFooter class="pt-4">
          <Button type="button" variant="outline" @click="createModal.close">
            Cancel
          </Button>
          <Button type="submit">
            Create Account
          </Button>
        </DialogFooter>
      </Form>
    </DialogContent>
  </DialogRoot>

  <!-- Edit Account Modal -->
  <DialogRoot v-model:open="editModal.isOpen.value">
    <DialogContent class="sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle>Edit Account</DialogTitle>
        <DialogDescription>
          Update account information
        </DialogDescription>
      </DialogHeader>

      <Form
          v-if="selectedAccount"
          @submit="onEditSubmit"
          :validation-schema="toTypedSchema(updateAccountSchema)"
          :initial-values="{
          name: selectedAccount.name,
          description: selectedAccount.description || '',
          is_active: selectedAccount.is_active,
        }"
          class="space-y-4 py-4"
      >
        <FormInput
            name="name"
            label="Account Name"
            placeholder="e.g., My Wallet"
            required
        />

        <!-- READ-ONLY FIELDS -->
        <div class="space-y-2 p-4 bg-gray-50 rounded-lg border border-gray-200">
          <p class="text-sm text-gray-600">
            <span class="font-medium">Type:</span>
            <span class="ml-2 capitalize">{{ accountTypeLabels[selectedAccount.type] }}</span>
          </p>
          <p class="text-sm text-gray-600">
            <span class="font-medium">Currency:</span>
            <span class="ml-2">{{ selectedAccount.currency }}</span>
          </p>
          <p class="text-sm text-gray-600">
            <span class="font-medium">Initial Balance:</span>
            {{ formatCurrency(selectedAccount.initial_balance, selectedAccount.currency) }}
          </p>
          <p class="text-sm text-gray-600">
            <span class="font-medium">Current Balance:</span>
            <span :class="parseFloat(selectedAccount.current_balance) >= 0 ? 'text-gray-900' : 'text-red-600'">
              {{ formatCurrency(selectedAccount.current_balance, selectedAccount.currency) }}
            </span>
          </p>
          <p class="text-xs text-gray-500 italic mt-2">
            💡 Type, currency and balances cannot be changed
          </p>
        </div>

        <FormInput
            name="description"
            label="Description"
            placeholder="Add notes about this account"
        />

        <DialogFooter class="pt-4 flex justify-between">
          <!-- DELETE BUTTON (LEFT) -->
          <Button
              type="button"
              variant="destructive"
              @click="openDeleteModal"
              class="mr-auto"
          >
            <Trash2 class="h-4 w-4 mr-2" />
            Delete
          </Button>

          <!-- CANCEL & SAVE (RIGHT) -->
          <div class="flex gap-2">
            <Button type="button" variant="outline" @click="editModal.close">
              Cancel
            </Button>
            <Button type="submit">
              Save Changes
            </Button>
          </div>
        </DialogFooter>
      </Form>
    </DialogContent>
  </DialogRoot>

  <!-- Delete Confirmation Modal -->
  <DeleteAccountModal
      v-if="selectedAccount"
      v-model:open="deleteModal"
      :account-name="selectedAccount.name"
      :loading="deleteLoading"
      @confirm="handleDeleteConfirm"
  />
</template>