<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { categoriesApi } from '@/api'
import type { Category } from '@/types'
import { Tags, Plus, ArrowUpRight, ArrowDownRight } from 'lucide-vue-next'

const loading = ref(true)
const categories = ref<Category[]>([])
const activeTab = ref<'expense' | 'income'>('expense')

onMounted(async () => {
  try {
    const response = await categoriesApi.list()
    if (response.success) {
      categories.value = response.data.categories
    }
  } finally {
    loading.value = false
  }
})

const filteredCategories = computed(() => {
  return categories.value.filter((c: Category) => c.type === activeTab.value && c.is_active)
})

const parentCategories = computed(() => {
  return filteredCategories.value.filter((c: Category) => !c.parent_id)
})

function getChildren(parentId: string): Category[] {
  return filteredCategories.value.filter((c: Category) => c.parent_id === parentId)
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">Categories</h1>
        <p class="text-gray-500 mt-1">Organize your income and expenses</p>
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-5 w-5" />
        Add Category
      </button>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 mb-6">
      <button
        @click="activeTab = 'expense'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
        :class="activeTab === 'expense' 
          ? 'bg-red-100 text-red-700' 
          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      >
        <ArrowDownRight class="h-4 w-4 inline mr-2" />
        Expenses
      </button>
      <button
        @click="activeTab = 'income'"
        class="px-4 py-2 rounded-lg font-medium transition-colors"
        :class="activeTab === 'income' 
          ? 'bg-green-100 text-green-700' 
          : 'bg-gray-100 text-gray-600 hover:bg-gray-200'"
      >
        <ArrowUpRight class="h-4 w-4 inline mr-2" />
        Income
      </button>
    </div>

    <div v-if="loading" class="flex items-center justify-center h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <div v-else class="bg-white rounded-xl shadow-sm border border-gray-200">
      <div v-for="parent in parentCategories" :key="parent.id" class="border-b border-gray-200 last:border-b-0">
        <div class="flex items-center justify-between p-4">
          <div class="flex items-center gap-3">
            <div
              class="p-2 rounded-lg"
              :class="activeTab === 'expense' ? 'bg-red-100' : 'bg-green-100'"
            >
              <Tags
                class="h-4 w-4"
                :class="activeTab === 'expense' ? 'text-red-600' : 'text-green-600'"
              />
            </div>
            <span class="font-medium text-gray-900">{{ parent.name }}</span>
          </div>
        </div>
        
        <!-- Children -->
        <div v-if="getChildren(parent.id).length > 0" class="bg-gray-50 px-4 py-2">
          <div
            v-for="child in getChildren(parent.id)"
            :key="child.id"
            class="flex items-center justify-between py-2 pl-8"
          >
            <span class="text-gray-700">{{ child.name }}</span>
          </div>
        </div>
      </div>

      <div v-if="parentCategories.length === 0" class="p-8 text-center text-gray-500">
        No {{ activeTab }} categories yet
      </div>
    </div>
  </div>
</template>
