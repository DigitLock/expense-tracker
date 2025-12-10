<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Toaster } from 'vue-sonner'
import DefaultLayout from '@/components/layout/DefaultLayout.vue'
import AuthLayout from '@/components/layout/AuthLayout.vue'

const route = useRoute()

const layout = computed(() => {
  return route.meta.layout === 'auth' ? AuthLayout : DefaultLayout
})
</script>

<template>
  <div id="app">
    <Toaster
        position="top-right"
        :duration="5000"
        :close-button="true"
        :rich-colors="true"
    />

    <component :is="layout">
      <router-view />
    </component>
  </div>
</template>