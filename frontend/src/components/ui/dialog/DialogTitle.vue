<script setup lang="ts">
import { DialogTitle, type DialogTitleProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface Props extends DialogTitleProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()

const titleClass = computed(() =>
  cn('text-lg font-semibold leading-none tracking-tight', props.class)
)
</script>

<template>
  <DialogTitle :class="titleClass">
    <slot />
  </DialogTitle>
</template>
