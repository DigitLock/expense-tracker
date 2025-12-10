import { toast } from 'vue-sonner'

/**
 * Toast composable for consistent notification usage across the app
 * 
 * @example
 * ```typescript
 * import { useToast } from '@/composables/useToast'
 * 
 * const toast = useToast()
 * 
 * toast.success('Account created successfully!')
 * toast.error('Failed to create account')
 * toast.info('Processing your request...')
 * toast.warning('This action cannot be undone')
 * ```
 */
export function useToast() {
  return {
    /**
     * Show success notification
     */
    success: (message: string, description?: string) => {
      toast.success(message, {
        description,
      })
    },

    /**
     * Show error notification
     */
    error: (message: string, description?: string) => {
      toast.error(message, {
        description,
      })
    },

    /**
     * Show info notification
     */
    info: (message: string, description?: string) => {
      toast.info(message, {
        description,
      })
    },

    /**
     * Show warning notification
     */
    warning: (message: string, description?: string) => {
      toast.warning(message, {
        description,
      })
    },

    /**
     * Show loading notification
     * Returns a function to dismiss the toast
     */
    loading: (message: string) => {
      const id = toast.loading(message)
      return () => toast.dismiss(id)
    },

    /**
     * Show promise-based notification
     * Automatically shows loading, success, or error based on promise result
     */
    promise: <T>(
      promise: Promise<T>,
      messages: {
        loading: string
        success: string | ((data: T) => string)
        error: string | ((error: Error) => string)
      }
    ) => {
      return toast.promise(promise, messages)
    },

    /**
     * Dismiss a specific toast or all toasts
     */
    dismiss: (toastId?: string | number) => {
      toast.dismiss(toastId)
    },
  }
}
