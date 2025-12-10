<script setup lang="ts">
import { DialogDescription, type DialogDescriptionProps } from 'radix-vue'
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface Props extends DialogDescriptionProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()

const descriptionClass = computed(() =>
  cn('text-sm text-muted-foreground', props.class)
)
</script>

<template>
  <DialogDescription :class="descriptionClass">
    <slot />
  </DialogDescription>
</template>
