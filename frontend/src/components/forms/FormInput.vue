<script setup lang="ts">
import { useField } from 'vee-validate'
import { Input } from '@/components/ui'
import FormField from './FormField.vue'

/**
 * FormInput - Text/Number/Email input with automatic validation
 * 
 * Integrates VeeValidate with shadcn Input component.
 * Automatically shows validation errors.
 * 
 * @example
 * ```vue
 * <FormInput 
 *   name="email" 
 *   label="Email Address" 
 *   type="email"
 *   placeholder="user@example.com"
 *   required
 * />
 * ```
 */

interface FormInputProps {
  /** Field name (must match schema) */
  name: string
  /** Label text */
  label?: string
  /** Input type */
  type?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url' | 'date'
  /** Placeholder text */
  placeholder?: string
  /** Show required asterisk */
  required?: boolean
  /** Description/hint text */
  description?: string
  /** Disabled state */
  disabled?: boolean
  /** Step for number inputs */
  step?: string | number
  /** Min value for number/date inputs */
  min?: string | number
  /** Max value for number/date inputs */
  max?: string | number
}

const props = withDefaults(defineProps<FormInputProps>(), {
  type: 'text',
  required: false,
  disabled: false,
})

// VeeValidate integration
const { value, errorMessage: _errorMessage } = useField<string | number>(() => props.name)

// For number inputs, convert string to number
const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (props.type === 'number') {
    value.value = target.value === '' ? 0 : Number(target.value)
  } else {
    value.value = target.value
  }
}
</script>

<template>
  <FormField
    :name="name"
    :label="label"
    :required="required"
    :description="description"
  >
    <template #default="{ hasError }">
      <Input
        :id="name"
        :model-value="value"
        :type="type"
        :placeholder="placeholder"
        :disabled="disabled"
        :step="step"
        :min="min"
        :max="max"
        :class="{ 'border-red-500': hasError }"
        @input="handleInput"
      />
    </template>
  </FormField>
</template>
