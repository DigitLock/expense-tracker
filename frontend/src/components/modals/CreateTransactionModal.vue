<script setup lang="ts">
import { ref, computed } from 'vue'
import { transactionsApi } from '@/api'
import type { Account, Category } from '@/types'
import { Form } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { createTransactionSchema, getTodayDateString } from '@/schemas/transaction'
import {
  DialogRoot,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from '@/components/ui'
import { FormInput, FormSelect, FormTextarea } from '@/components/forms'
import { useToast } from '@/composables/useToast'

interface Props {
  open: boolean
  accounts: Account[]
  categories: Category[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  'created': []
}>()

const toast = useToast()
const submitting = ref(false)
const selectedType = ref<'income' | 'expense'>('expense')

// Options
const typeOptions = [
  { value: 'expense', label: 'Expense' },
  { value: 'income', label: 'Income' },
]

const currencyOptions = [
  { value: 'RSD', label: 'RSD (Dinar)' },
  { value: 'EUR', label: 'EUR (Euro)' },
  { value: 'USD', label: 'USD (Dollar)' },
]

const accountOptions = computed(() => {
  return props.accounts
      .filter(a => a.is_active)
      .map(a => ({
        value: a.id,
        label: `${a.name} (${a.currency})`,
      }))
})

const categoryOptions = computed(() => {
  return props.categories
      .filter(c => c.is_active && c.type === selectedType.value)
      .map(c => ({
        value: c.id,
        label: c.name,
      }))
})

const handleOpenChange = (open: boolean) => {
  emit('update:open', open)
  if (!open) {
    selectedType.value = 'expense'
  }
}

const onSubmit = async (values: any) => {
  try {
    submitting.value = true

    const response = await transactionsApi.create({
      type: values.type,
      amount: parseFloat(values.amount),
      currency: values.currency,
      account_id: values.account_id,
      category_id: values.category_id,
      description: values.description || undefined,
      date: values.date,
    })

    if (response.success) {
      toast.success('Transaction created successfully!')
      emit('created')
      handleOpenChange(false)
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string, details?: any[] } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to create transaction'
    const errorDetails = axiosError.response?.data?.error?.details

    if (errorDetails && errorDetails.length > 0) {
      // Show validation errors
      errorDetails.forEach((detail: any) => {
        toast.error(`${detail.field}: ${detail.message}`)
      })
    } else {
      toast.error(errorMessage)
    }
  } finally {
    submitting.value = false
  }
}

// Watch type change to reset category
const handleTypeChange = (newType: string) => {
  selectedType.value = newType as 'income' | 'expense'
}
</script>

<template>
  <DialogRoot :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Create Transaction</DialogTitle>
        <DialogDescription>
          Add a new income or expense transaction
        </DialogDescription>
      </DialogHeader>

      <Form
          @submit="onSubmit"
          :validation-schema="toTypedSchema(createTransactionSchema)"
          :initial-values="{
          type: 'expense',
          amount: '',
          currency: 'RSD',
          account_id: '',
          category_id: '',
          date: getTodayDateString(),
          description: '',
        }"
          class="space-y-4 py-4"
          v-slot="{ values }"
      >
        <!-- Type -->
        <FormSelect
            name="type"
            label="Type"
            :options="typeOptions"
            placeholder="Select type"
            required
            @update:model-value="handleTypeChange"
        />

        <!-- Amount & Currency (side by side) -->
        <div class="grid grid-cols-2 gap-4">
          <FormInput
              name="amount"
              label="Amount"
              type="number"
              step="0.01"
              placeholder="0.00"
              required
          />

          <FormSelect
              name="currency"
              label="Currency"
              :options="currencyOptions"
              placeholder="Select currency"
              required
          />
        </div>

        <!-- Account -->
        <FormSelect
            name="account_id"
            label="Account"
            :options="accountOptions"
            placeholder="Select account"
            required
        />

        <!-- Category (filtered by type) -->
        <FormSelect
            name="category_id"
            label="Category"
            :options="categoryOptions"
            placeholder="Select category"
            required
            :key="selectedType"
        />

        <!-- Date -->
        <FormInput
            name="date"
            label="Date"
            type="date"
            required
            :max="getTodayDateString()"
        />

        <!-- Description -->
        <FormTextarea
            name="description"
            label="Description (Optional)"
            placeholder="Add a note about this transaction..."
            rows="3"
        />

        <DialogFooter class="pt-4">
          <Button
              type="button"
              variant="outline"
              @click="handleOpenChange(false)"
              :disabled="submitting"
          >
            Cancel
          </Button>
          <Button
              type="submit"
              :disabled="submitting"
          >
            <span v-if="submitting" class="flex items-center gap-2">
              <span class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
              Creating...
            </span>
            <span v-else>Create Transaction</span>
          </Button>
        </DialogFooter>
      </Form>
    </DialogContent>
  </DialogRoot>
</template>