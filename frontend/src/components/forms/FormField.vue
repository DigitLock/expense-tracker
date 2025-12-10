<script setup lang="ts">
import { computed } from 'vue'
import { useField } from 'vee-validate'
import { Label } from '@/components/ui'

/**
 * FormField - Wrapper component for form fields
 * 
 * Automatically handles:
 * - Label rendering
 * - Error message display
 * - Required field indicator
 * - Error state styling
 * 
 * @example
 * ```vue
 * <FormField name="email" label="Email">
 *   <Input v-model="email" type="email" />
 * </FormField>
 * ```
 */

interface FormFieldProps {
  /** Field name for validation (must match schema) */
  name: string
  /** Label text to display */
  label?: string
  /** Show required asterisk */
  required?: boolean
  /** Additional description/hint text */
  description?: string
}

const props = withDefaults(defineProps<FormFieldProps>(), {
  required: false,
})

// Get field error from VeeValidate
const { errorMessage } = useField(() => props.name)

const hasError = computed(() => !!errorMessage.value)
</script>

<template>
  <div class="space-y-2">
    <!-- Label -->
    <Label v-if="label" :for="name" class="text-sm font-medium text-gray-700">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </Label>

    <!-- Description -->
    <p v-if="description" class="text-sm text-gray-500">
      {{ description }}
    </p>

    <!-- Input slot -->
    <div>
      <slot :has-error="hasError" :error-message="errorMessage" />
    </div>

    <!-- Error message -->
    <p v-if="hasError" class="text-sm text-red-600">
      {{ errorMessage }}
    </p>
  </div>
</template>
