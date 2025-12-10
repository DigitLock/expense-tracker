<script setup lang="ts">
import { computed, type HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface Props {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()

const footerClass = computed(() =>
  cn('flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2', props.class)
)
</script>

<template>
  <div :class="footerClass">
    <slot />
  </div>
</template>
