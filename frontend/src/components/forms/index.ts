/**
 * Form Components - Reusable form fields with VeeValidate integration
 * 
 * These components automatically handle:
 * - Label rendering
 * - Error message display
 * - Validation state
 * - Consistent styling
 * 
 * @module components/forms
 */

export { default as FormField } from './FormField.vue'
export { default as FormInput } from './FormInput.vue'
export { default as FormSelect } from './FormSelect.vue'
export { default as FormTextarea } from './FormTextarea.vue'

/**
 * Usage examples:
 * 
 * @example Basic text input
 * ```vue
 * <FormInput 
 *   name="name" 
 *   label="Account Name" 
 *   placeholder="e.g., My Wallet"
 *   required
 * />
 * ```
 * 
 * @example Number input
 * ```vue
 * <FormInput 
 *   name="amount" 
 *   label="Amount" 
 *   type="number"
 *   step="0.01"
 *   min="0"
 * />
 * ```
 * 
 * @example Select dropdown
 * ```vue
 * <FormSelect 
 *   name="type" 
 *   label="Account Type"
 *   :options="accountTypeOptions"
 *   placeholder="Select type"
 *   required
 * />
 * ```
 * 
 * @example Textarea
 * ```vue
 * <FormTextarea 
 *   name="description" 
 *   label="Description"
 *   placeholder="Add notes..."
 *   :rows="4"
 * />
 * ```
 * 
 * @example Custom field with FormField
 * ```vue
 * <FormField name="custom" label="Custom Field" required>
 *   <template #default="{ hasError }">
 *     <CustomInput :error="hasError" />
 *   </template>
 * </FormField>
 * ```
 */
