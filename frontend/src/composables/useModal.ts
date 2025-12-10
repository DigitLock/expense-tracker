import { ref } from 'vue'

/**
 * Composable for managing modal/dialog state
 * 
 * Provides a clean API for opening and closing modals with TypeScript support.
 * Works seamlessly with shadcn-vue Dialog component and v-model:open binding.
 * 
 * @example
 * ```typescript
 * import { useModal } from '@/composables/useModal'
 * 
 * const createModal = useModal()
 * const editModal = useModal()
 * 
 * // Open modal
 * createModal.open()
 * 
 * // Close modal
 * createModal.close()
 * 
 * // Toggle modal
 * createModal.toggle()
 * 
 * // Check if open
 * if (createModal.isOpen.value) {
 *   console.log('Modal is open')
 * }
 * ```
 * 
 * @example
 * ```vue
 * <script setup lang="ts">
 * import { useModal } from '@/composables/useModal'
 * import { DialogRoot, DialogContent, DialogTrigger } from '@/components/ui'
 * 
 * const modal = useModal()
 * </script>
 * 
 * <template>
 *   <DialogRoot v-model:open="modal.isOpen.value">
 *     <DialogTrigger as-child>
 *       <Button @click="modal.open">Open Modal</Button>
 *     </DialogTrigger>
 *     <DialogContent>
 *       <!-- Modal content -->
 *       <Button @click="modal.close">Close</Button>
 *     </DialogContent>
 *   </DialogRoot>
 * </template>
 * ```
 */
export function useModal(initialState = false) {
  /**
   * Reactive state for modal visibility
   */
  const isOpen = ref(initialState)

  /**
   * Open the modal
   */
  function open() {
    isOpen.value = true
  }

  /**
   * Close the modal
   */
  function close() {
    isOpen.value = false
  }

  /**
   * Toggle modal state
   */
  function toggle() {
    isOpen.value = !isOpen.value
  }

  /**
   * Set modal state directly
   * @param value - New state value
   */
  function setState(value: boolean) {
    isOpen.value = value
  }

  return {
    isOpen,
    open,
    close,
    toggle,
    setState,
  }
}

/**
 * Type for modal composable return value
 */
export type UseModalReturn = ReturnType<typeof useModal>
