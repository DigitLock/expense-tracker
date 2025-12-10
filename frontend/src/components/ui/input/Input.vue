<script setup lang="ts">
import { computed, type InputHTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  type?: InputHTMLAttributes['type']
  class?: InputHTMLAttributes['class']
  modelValue?: string | number
  disabled?: boolean
  placeholder?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const inputClass = computed(() =>
  cn(
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-base shadow-sm transition-colors',
    'file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground',
    'placeholder:text-muted-foreground',
    'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
    'disabled:cursor-not-allowed disabled:opacity-50',
    'md:text-sm',
    props.class
  )
)

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}
</script>

<template>
  <input
    :type="type"
    :class="inputClass"
    :value="modelValue"
    :disabled="disabled"
    :placeholder="placeholder"
    @input="handleInput"
  />
</template>
