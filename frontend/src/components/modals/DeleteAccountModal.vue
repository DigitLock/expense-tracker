<script setup lang="ts">
import { ref, computed } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { z } from 'zod'
import {
  DialogRoot,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from '@/components/ui'
import { FormInput } from '@/components/forms'
import { AlertTriangle } from 'lucide-vue-next'

interface Props {
  open: boolean
  accountName: string
  loading?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  'confirm': []
}>()

// Validation: user must type exact account name
const deleteSchema = z.object({
  confirmation: z.string().refine(
      (val) => val === props.accountName,
      { message: `Please type "${props.accountName}" to confirm` }
  ),
})

const { handleSubmit, resetForm } = useForm({
  validationSchema: toTypedSchema(deleteSchema),
})

const onSubmit = handleSubmit(() => {
  emit('confirm')
})

// Reset form when modal closes
const handleOpenChange = (open: boolean) => {
  emit('update:open', open)
  if (!open) {
    resetForm()
  }
}
</script>

<template>
  <DialogRoot :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[425px]">
      <DialogHeader>
        <div class="flex items-center gap-3">
          <div class="p-2 bg-red-100 rounded-lg">
            <AlertTriangle class="h-5 w-5 text-red-600" />
          </div>
          <div>
            <DialogTitle>Delete Account</DialogTitle>
            <DialogDescription>
              This action cannot be undone
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <form @submit="onSubmit" class="space-y-4 py-4">
        <div class="bg-red-50 border border-red-200 rounded-lg p-4">
          <p class="text-sm text-red-800">
            You are about to delete <strong>{{ accountName }}</strong>.
            All related transactions will be affected.
          </p>
        </div>

        <FormInput
            name="confirmation"
            label="Type account name to confirm"
            :placeholder="accountName"
            autocomplete="off"
        />

        <DialogFooter class="pt-4">
          <Button
              type="button"
              variant="outline"
              @click="handleOpenChange(false)"
              :disabled="loading"
          >
            Cancel
          </Button>
          <Button
              type="submit"
              variant="destructive"
              :disabled="loading"
          >
            <span v-if="loading" class="flex items-center gap-2">
              <span class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
              Deleting...
            </span>
            <span v-else>Delete Account</span>
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </DialogRoot>
</template>
