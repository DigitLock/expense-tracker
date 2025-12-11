<script setup lang="ts">
import { computed, watch } from 'vue'
import { useField } from 'vee-validate'
import {
  SelectRoot,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui'
import FormField from './FormField.vue'

/**
 * FormSelect - Select dropdown with automatic validation
 *
 * Integrates VeeValidate with shadcn Select component.
 * Supports both simple arrays and option objects.
 *
 * @example
 * ```vue
 * <FormSelect
 *   name="type"
 *   label="Account Type"
 *   :options="accountTypeOptions"
 *   placeholder="Select type"
 *   required
 * />
 * ```
 */

interface SelectOption {
  value: string
  label: string
  disabled?: boolean
}

interface FormSelectProps {
  /** Field name (must match schema) */
  name: string
  /** Label text */
  label?: string
  /** Options array */
  options: SelectOption[] | string[]
  /** Placeholder text */
  placeholder?: string
  /** Show required asterisk */
  required?: boolean
  /** Description/hint text */
  description?: string
  /** Disabled state */
  disabled?: boolean
}

const props = withDefaults(defineProps<FormSelectProps>(), {
  required: false,
  disabled: false,
  placeholder: 'Select an option',
})

// VeeValidate integration
const { value, errorMessage: _errorMessage, setValue } = useField<string>(() => props.name)

// Normalize options to object format
const normalizedOptions = computed(() => {
  return props.options.map(opt => {
    if (typeof opt === 'string') {
      return { value: opt, label: opt }
    }
    return opt
  })
})

// Watch for options changes and reset value if current value is not in new options
watch(normalizedOptions, (newOptions) => {
  if (value.value) {
    const isValueInOptions = newOptions.some(opt => opt.value === value.value)
    if (!isValueInOptions) {
      setValue('')
    }
  }
}, { deep: true })
</script>

<template>
  <FormField
      :name="name"
      :label="label"
      :required="required"
      :description="description"
  >
    <template #default="{ hasError }">
      <SelectRoot
          v-model="value"
          :disabled="disabled"
      >
        <SelectTrigger
            :id="name"
            :class="{ 'border-red-500': hasError }"
        >
          <SelectValue :placeholder="placeholder" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem
              v-for="option in normalizedOptions"
              :key="option.value"
              :value="option.value"
              :disabled="option.disabled"
          >
            {{ option.label }}
          </SelectItem>
        </SelectContent>
      </SelectRoot>
    </template>
  </FormField>
</template>