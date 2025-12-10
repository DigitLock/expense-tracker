<script setup lang="ts">
import { useField } from 'vee-validate'
import FormField from './FormField.vue'

/**
 * FormTextarea - Multi-line text input with automatic validation
 * 
 * Integrates VeeValidate with native textarea element.
 * Automatically shows validation errors.
 * 
 * @example
 * ```vue
 * <FormTextarea 
 *   name="description" 
 *   label="Description"
 *   placeholder="Add notes..."
 *   :rows="4"
 * />
 * ```
 */

interface FormTextareaProps {
  /** Field name (must match schema) */
  name: string
  /** Label text */
  label?: string
  /** Placeholder text */
  placeholder?: string
  /** Show required asterisk */
  required?: boolean
  /** Description/hint text */
  description?: string
  /** Disabled state */
  disabled?: boolean
  /** Number of visible rows */
  rows?: number
  /** Max length */
  maxlength?: number
}

const props = withDefaults(defineProps<FormTextareaProps>(), {
  required: false,
  disabled: false,
  rows: 3,
})

// VeeValidate integration
const { value, errorMessage } = useField<string>(() => props.name)
</script>

<template>
  <FormField
    :name="name"
    :label="label"
    :required="required"
    :description="description"
  >
    <template #default="{ hasError }">
      <textarea
        :id="name"
        v-model="value"
        :placeholder="placeholder"
        :disabled="disabled"
        :rows="rows"
        :maxlength="maxlength"
        :class="[
          'flex min-h-20 w-full rounded-md border border-input bg-background px-3 py-2 text-sm',
          'placeholder:text-muted-foreground',
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
          'disabled:cursor-not-allowed disabled:opacity-50',
          { 'border-red-500': hasError }
        ]"
      />
    </template>
  </FormField>
</template>
