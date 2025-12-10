<script setup lang="ts">
import { ref, computed } from 'vue'
import { transactionsApi } from '@/api'
import type { Transaction, Account, Category } from '@/types'
import { Form } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { updateTransactionSchema } from '@/schemas/transaction'
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
import { Trash2 } from 'lucide-vue-next'

interface Props {
  open: boolean
  transaction: Transaction
  accounts: Account[]
  categories: Category[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  'updated': []
  'delete': []
}>()

const toast = useToast()
const submitting = ref(false)

// Options
const currencyOptions = [
  { value: 'RSD', label: 'RSD (Dinar)' },
  { value: 'EUR', label: 'EUR (Euro)' },
  { value: 'USD', label: 'USD (Dollar)' },
]

const categoryOptions = computed(() => {
  return props.categories
      .filter(c => c.is_active && c.type === props.transaction.type)
      .map(c => ({
        value: c.id,
        label: c.name,
      }))
})

const handleOpenChange = (open: boolean) => {
  emit('update:open', open)
}

const onSubmit = async (values: any) => {
  try {
    submitting.value = true

    const updateData: any = {}

    // Only send changed fields
    if (values.amount && parseFloat(values.amount) !== parseFloat(props.transaction.amount)) {
      updateData.amount = parseFloat(values.amount)
    }

    if (values.currency && values.currency !== props.transaction.currency) {
      updateData.currency = values.currency
    }

    if (values.category_id && values.category_id !== props.transaction.category.id) {
      updateData.category_id = values.category_id
    }

    if (values.date && values.date !== props.transaction.date) {
      updateData.date = values.date
    }

    if (values.description !== (props.transaction.description || '')) {
      updateData.description = values.description || undefined
    }

    // If nothing changed, just close
    if (Object.keys(updateData).length === 0) {
      toast.info('No changes to save')
      handleOpenChange(false)
      return
    }

    const response = await transactionsApi.update(props.transaction.id, updateData)

    if (response.success) {
      toast.success('Transaction updated successfully!')
      emit('updated')
      handleOpenChange(false)
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string, details?: any[] } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to update transaction'
    const errorDetails = axiosError.response?.data?.error?.details

    if (errorDetails && errorDetails.length > 0) {
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

const handleDelete = () => {
  emit('delete')
}
</script>

<template>
  <DialogRoot :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>Edit Transaction</DialogTitle>
        <DialogDescription>
          Update transaction information
        </DialogDescription>
      </DialogHeader>

      <Form
          @submit="onSubmit"
          :validation-schema="toTypedSchema(updateTransactionSchema)"
          :initial-values="{
          amount: transaction.amount,
          currency: transaction.currency,
          category_id: transaction.category.id,
          date: transaction.date,
          description: transaction.description || '',
        }"
          class="space-y-4 py-4"
      >
        <!-- READ-ONLY: Type & Account -->
        <div class="space-y-3 p-4 bg-gray-50 rounded-lg border border-gray-200">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm font-medium text-gray-700 mb-1">Type</p>
              <div class="flex items-center gap-2">
                <span
                    class="px-3 py-1.5 rounded text-sm font-medium"
                    :class="transaction.type === 'income'
                    ? 'bg-green-100 text-green-700'
                    : 'bg-red-100 text-red-700'"
                >
                  {{ transaction.type === 'income' ? 'Income' : 'Expense' }}
                </span>
              </div>
            </div>

            <div>
              <p class="text-sm font-medium text-gray-700 mb-1">Account</p>
              <p class="text-sm text-gray-900 py-1.5">{{ transaction.account.name }}</p>
            </div>
          </div>

          <p class="text-xs text-gray-500 italic">
            💡 Type and Account cannot be changed. Delete and create a new transaction instead.
          </p>
        </div>

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

        <!-- Category (filtered by type) -->
        <FormSelect
            name="category_id"
            label="Category"
            :options="categoryOptions"
            placeholder="Select category"
            required
        />

        <!-- Date -->
        <FormInput
            name="date"
            label="Date"
            type="date"
            required
        />

        <!-- Description -->
        <FormTextarea
            name="description"
            label="Description (Optional)"
            placeholder="Add a note about this transaction..."
            rows="3"
        />

        <DialogFooter class="pt-4 flex justify-between">
          <!-- DELETE BUTTON (LEFT) -->
          <Button
              type="button"
              variant="destructive"
              @click="handleDelete"
              :disabled="submitting"
              class="mr-auto"
          >
            <Trash2 class="h-4 w-4 mr-2" />
            Delete
          </Button>

          <!-- CANCEL & SAVE (RIGHT) -->
          <div class="flex gap-2">
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
                Saving...
              </span>
              <span v-else>Save Changes</span>
            </Button>
          </div>
        </DialogFooter>
      </Form>
    </DialogContent>
  </DialogRoot>
</template>