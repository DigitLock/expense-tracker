<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { categoriesApi } from '@/api'
import type { Category } from '@/types'
import { Tags, Plus, ArrowUpRight, ArrowDownRight, Trash2 } from 'lucide-vue-next'
import { useModal } from '@/composables/useModal'
import { useToast } from '@/composables/useToast'
import { Form } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import {
  createCategorySchema,
  updateCategorySchema,
  categoryTypes,
  categoryTypeLabels,
} from '@/schemas'
import {
  DialogRoot,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from '@/components/ui'
import { FormInput, FormSelect } from '@/components/forms'
import DeleteCategoryModal from '@/components/modals/DeleteCategoryModal.vue'

const loading = ref(true)
const categories = ref<Category[]>([])
const activeTab = ref<'expense' | 'income'>('expense')
const createModal = useModal()
const editModal = useModal()
const deleteModal = ref(false)
const toast = useToast()
const selectedCategory = ref<Category | null>(null)
const deleteLoading = ref(false)

// Load categories
async function loadCategories() {
  try {
    loading.value = true
    const response = await categoriesApi.list()
    if (response.success) {
      categories.value = response.data.categories
    }
  } catch (err) {
    console.error('Failed to load categories:', err)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadCategories()
})

// Computed
const filteredCategories = computed(() => {
  return categories.value.filter((c: Category) => c.type === activeTab.value && c.is_active)
})

const parentCategories = computed(() => {
  return filteredCategories.value.filter((c: Category) => !c.parent_id)
})

const parentOptions = computed(() => {
  // Only root categories of the current type (for create modal)
  return categories.value
      .filter((c: Category) => c.type === activeTab.value && !c.parent_id && c.is_active)
      .map((c: Category) => ({
        value: c.id,
        label: c.name,
      }))
})

const editParentOptions = computed(() => {
  if (!selectedCategory.value) return []

  // Only root categories of the same type, excluding self
  return categories.value
      .filter((c: Category) =>
          c.type === selectedCategory.value!.type &&
          !c.parent_id &&
          c.id !== selectedCategory.value!.id &&
          c.is_active
      )
      .map((c: Category) => ({
        value: c.id,
        label: c.name,
      }))
})

const categoryTypeOptions = categoryTypes.map(t => ({
  value: t,
  label: categoryTypeLabels[t],
}))

function getChildren(parentId: string): Category[] {
  return filteredCategories.value.filter((c: Category) => c.parent_id === parentId)
}

function hasChildren(categoryId: string): boolean {
  return categories.value.some((c: Category) => c.parent_id === categoryId)
}

// Get parent category name by ID
function getParentName(parentId: string | null): string {
  if (!parentId) return ''
  const parent = categories.value.find((c: Category) => c.id === parentId)
  return parent?.name || 'Unknown'
}

// CREATE
const onCreateSubmit = async (values: any) => {
  try {
    const response = await categoriesApi.create({
      name: values.name,
      type: values.type,
      parent_id: values.parent_id || undefined,
    })

    if (response.success) {
      toast.success('Category created successfully!')
      createModal.close()
      await loadCategories()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to create category'
    toast.error(errorMessage)
  }
}

// EDIT
function openEditModal(category: Category) {
  selectedCategory.value = category
  editModal.open()
}

const onEditSubmit = async (values: any) => {
  if (!selectedCategory.value) return

  try {
    const response = await categoriesApi.update(selectedCategory.value.id, {
      name: values.name,
      parent_id: values.parent_id,
      is_active: values.is_active,
    })

    if (response.success) {
      toast.success('Category updated successfully!')
      editModal.close()
      await loadCategories()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to update category'
    toast.error(errorMessage)
  }
}

// DELETE
function openDeleteModal() {
  editModal.close()
  deleteModal.value = true
}

async function handleDeleteConfirm() {
  if (!selectedCategory.value) return

  try {
    deleteLoading.value = true
    const response = await categoriesApi.delete(selectedCategory.value.id)

    if (response.success) {
      toast.success('Category deleted successfully!')
      deleteModal.value = false
      selectedCategory.value = null
      await loadCategories()
    }
  } catch (err) {
    const axiosError = err as { response?: { data?: { error?: { message?: string } } } }
    const errorMessage = axiosError.response?.data?.error?.message || 'Failed to delete category'
    toast.error(errorMessage)
  } finally {
    deleteLoading.value = false
  }
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
          @click="createModal.open"
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
        <div
            @click="openEditModal(parent)"
            class="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50 transition-colors"
        >
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
        <div v-if="getChildren(parent.id).length > 0" class="bg-gray-50">
          <div
              v-for="child in getChildren(parent.id)"
              :key="child.id"
              @click="openEditModal(child)"
              class="flex items-center justify-between py-2 pl-12 pr-4 cursor-pointer hover:bg-gray-100 transition-colors border-t border-gray-100"
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

  <!-- Create Category Modal -->
  <DialogRoot v-model:open="createModal.isOpen.value">
    <DialogContent class="sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle>Create New Category</DialogTitle>
        <DialogDescription>
          Add a new category to organize your transactions
        </DialogDescription>
      </DialogHeader>

      <Form
          @submit="onCreateSubmit"
          :validation-schema="toTypedSchema(createCategorySchema)"
          :initial-values="{
          name: '',
          type: activeTab,
          parent_id: null,
        }"
          class="space-y-4 py-4"
      >
        <FormInput
            name="name"
            label="Category Name"
            placeholder="e.g., Groceries"
            required
        />

        <FormSelect
            name="type"
            label="Type"
            :options="categoryTypeOptions"
            placeholder="Select type"
            required
        />

        <FormSelect
            name="parent_id"
            label="Parent Category (Optional)"
            :options="parentOptions"
            placeholder="None - Top level category"
        />

        <DialogFooter class="pt-4">
          <Button type="button" variant="outline" @click="createModal.close">
            Cancel
          </Button>
          <Button type="submit">
            Create Category
          </Button>
        </DialogFooter>
      </Form>
    </DialogContent>
  </DialogRoot>

  <!-- Edit Category Modal -->
  <DialogRoot v-model:open="editModal.isOpen.value">
    <DialogContent class="sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle>Edit Category</DialogTitle>
        <DialogDescription>
          Update category information
        </DialogDescription>
      </DialogHeader>

      <Form
          v-if="selectedCategory"
          @submit="onEditSubmit"
          :validation-schema="toTypedSchema(updateCategorySchema)"
          :initial-values="{
          name: selectedCategory.name,
          parent_id: selectedCategory.parent_id,
          is_active: selectedCategory.is_active,
        }"
          class="space-y-4 py-4"
      >
        <FormInput
            name="name"
            label="Category Name"
            placeholder="e.g., Groceries"
            required
        />

        <!-- READ-ONLY TYPE -->
        <div class="space-y-2 p-4 bg-gray-50 rounded-lg border border-gray-200">
          <p class="text-sm text-gray-600">
            <span class="font-medium">Type:</span>
            <span
                class="ml-2 px-2 py-1 rounded text-xs font-medium"
                :class="selectedCategory.type === 'income'
                ? 'bg-green-100 text-green-700'
                : 'bg-red-100 text-red-700'"
            >
              {{ categoryTypeLabels[selectedCategory.type] }}
            </span>
          </p>
          <p class="text-xs text-gray-500 italic">
            💡 Category type cannot be changed
          </p>
        </div>

        <FormSelect
            v-if="!selectedCategory.parent_id"
            name="parent_id"
            label="Parent Category (Optional)"
            :options="editParentOptions"
            placeholder="None - Keep as top level"
        />

        <div v-else class="space-y-2 p-4 bg-gray-50 rounded-lg border border-gray-200">
          <p class="text-sm text-gray-600">
            <span class="font-medium">Parent Category:</span>
            <span class="ml-2">{{ getParentName(selectedCategory.parent_id) }}</span>
          </p>
          <p class="text-xs text-gray-500 italic">
            💡 Parent category cannot be changed for child categories
          </p>
        </div>

        <DialogFooter class="pt-4 flex justify-between">
          <!-- DELETE BUTTON (LEFT) -->
          <Button
              type="button"
              variant="destructive"
              @click="openDeleteModal"
              class="mr-auto"
              :disabled="hasChildren(selectedCategory.id)"
          >
            <Trash2 class="h-4 w-4 mr-2" />
            Delete
          </Button>

          <!-- CANCEL & SAVE (RIGHT) -->
          <div class="flex gap-2">
            <Button type="button" variant="outline" @click="editModal.close">
              Cancel
            </Button>
            <Button type="submit">
              Save Changes
            </Button>
          </div>
        </DialogFooter>

        <div v-if="hasChildren(selectedCategory.id)" class="text-xs text-amber-600 bg-amber-50 p-3 rounded-lg">
          ⚠️ Cannot delete this category because it has child categories. Delete the children first.
        </div>
      </Form>
    </DialogContent>
  </DialogRoot>

  <!-- Delete Confirmation Modal -->
  <DeleteCategoryModal
      v-if="selectedCategory"
      v-model:open="deleteModal"
      :category-name="selectedCategory.name"
      :loading="deleteLoading"
      @confirm="handleDeleteConfirm"
  />
</template>