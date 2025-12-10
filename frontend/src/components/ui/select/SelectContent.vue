<script setup lang="ts">
import {
  SelectContent,
  SelectPortal,
  SelectScrollDownButton,
  SelectScrollUpButton,
  SelectViewport,
  type SelectContentEmits,
  type SelectContentProps,
} from 'radix-vue'
import { ChevronDown, ChevronUp } from 'lucide-vue-next'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface ContentProps extends SelectContentProps {
  class?: HTMLAttributes['class']
}

const props = withDefaults(defineProps<ContentProps>(), {
  position: 'popper',
  sideOffset: 4,
})

const emits = defineEmits<SelectContentEmits>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})

const contentClass = computed(() =>
  cn(
    'relative z-50 max-h-96 min-w-32 overflow-hidden rounded-md border bg-popover text-popover-foreground shadow-md',
    'bg-blue-50 border border-gray-300 rounded-md shadow-lg',
    'text-gray-900',
    'data-[state=open]:animate-in data-[state=closed]:animate-out',
    'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
    'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
    'data-[side=bottom]:slide-in-from-top-2',
    'data-[side=left]:slide-in-from-right-2',
    'data-[side=right]:slide-in-from-left-2',
    'data-[side=top]:slide-in-from-bottom-2',
    props.position === 'popper' &&
      'data-[side=bottom]:translate-y-1 data-[side=left]:-translate-x-1 data-[side=right]:translate-x-1 data-[side=top]:-translate-y-1',
    props.class
  )
)

const viewportClass = computed(() =>
  cn(
    'p-1',
    props.position === 'popper' &&
      'h-[var(--radix-select-trigger-height)] w-full min-w-[var(--radix-select-trigger-width)]'
  )
)
</script>

<template>
  <SelectPortal>
    <SelectContent
      v-bind="delegatedProps"
      :class="contentClass"
      @close-auto-focus="emits('closeAutoFocus', $event)"
      @escape-key-down="emits('escapeKeyDown', $event)"
      @pointer-down-outside="emits('pointerDownOutside', $event)"
    >
      <SelectScrollUpButton
        class="flex cursor-default items-center justify-center py-1"
      >
        <ChevronUp class="h-4 w-4" />
      </SelectScrollUpButton>

      <SelectViewport :class="viewportClass">
        <slot />
      </SelectViewport>

      <SelectScrollDownButton
        class="flex cursor-default items-center justify-center py-1"
      >
        <ChevronDown class="h-4 w-4" />
      </SelectScrollDownButton>
    </SelectContent>
  </SelectPortal>
</template>
