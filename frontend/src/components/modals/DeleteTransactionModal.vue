<script setup lang="ts">
import {
  DialogRoot,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from '@/components/ui'
import { AlertTriangle } from 'lucide-vue-next'

interface Props {
  open: boolean
  transactionDescription: string
  loading?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{
  'update:open': [value: boolean]
  'confirm': []
}>()

const handleOpenChange = (open: boolean) => {
  emit('update:open', open)
}

const handleConfirm = () => {
  emit('confirm')
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
            <DialogTitle>Delete Transaction</DialogTitle>
            <DialogDescription>
              This action cannot be undone
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <div class="py-4">
        <div class="bg-red-50 border border-red-200 rounded-lg p-4">
          <p class="text-sm text-red-800">
            You are about to delete the transaction:<br>
            <strong class="font-semibold">{{ transactionDescription }}</strong>
          </p>
          <p class="text-sm text-red-700 mt-2">
            The account balance will be automatically recalculated.
          </p>
        </div>
      </div>

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
            type="button"
            variant="destructive"
            @click="handleConfirm"
            :disabled="loading"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <span class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
            Deleting...
          </span>
          <span v-else>Delete Transaction</span>
        </Button>
      </DialogFooter>
    </DialogContent>
  </DialogRoot>
</template>