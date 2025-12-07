<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  LayoutDashboard,
  Wallet,
  Tags,
  ArrowLeftRight,
  BarChart3,
  LogOut,
  Menu,
  X,
} from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const isMobileMenuOpen = ref(false)

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Accounts', href: '/accounts', icon: Wallet },
  { name: 'Categories', href: '/categories', icon: Tags },
  { name: 'Transactions', href: '/transactions', icon: ArrowLeftRight },
  { name: 'Reports', href: '/reports', icon: BarChart3 },
]

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function closeMobileMenu() {
  isMobileMenuOpen.value = false
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Mobile menu button -->
    <div class="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white border-b border-gray-200 px-4 py-3">
      <div class="flex items-center justify-between">
        <span class="text-lg font-semibold text-gray-900">Expense Tracker</span>
        <button
          @click="isMobileMenuOpen = !isMobileMenuOpen"
          class="p-2 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100"
        >
          <Menu v-if="!isMobileMenuOpen" class="h-6 w-6" />
          <X v-else class="h-6 w-6" />
        </button>
      </div>
    </div>

    <!-- Mobile menu overlay -->
    <div
      v-if="isMobileMenuOpen"
      class="lg:hidden fixed inset-0 z-30 bg-black/50"
      @click="closeMobileMenu"
    />

    <!-- Mobile sidebar -->
    <aside
      v-if="isMobileMenuOpen"
      class="lg:hidden fixed top-14 left-0 bottom-0 z-40 w-64 bg-white border-r border-gray-200 overflow-y-auto"
    >
      <nav class="p-4 space-y-1">
        <router-link
          v-for="item in navigation"
          :key="item.name"
          :to="item.href"
          @click="closeMobileMenu"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-700 hover:bg-gray-100 transition-colors"
          active-class="bg-blue-50 text-blue-700 hover:bg-blue-50"
        >
          <component :is="item.icon" class="h-5 w-5" />
          {{ item.name }}
        </router-link>
      </nav>
      
      <div class="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-200">
        <div class="mb-3 px-3">
          <p class="text-sm font-medium text-gray-900">{{ authStore.userName }}</p>
          <p class="text-xs text-gray-500">{{ authStore.userEmail }}</p>
        </div>
        <button
          @click="handleLogout"
          class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-red-600 hover:bg-red-50 transition-colors"
        >
          <LogOut class="h-5 w-5" />
          Sign out
        </button>
      </div>
    </aside>

    <!-- Desktop sidebar -->
    <aside class="hidden lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col">
      <div class="flex flex-col flex-grow bg-white border-r border-gray-200">
        <!-- Logo -->
        <div class="flex items-center h-16 px-6 border-b border-gray-200">
          <span class="text-xl font-bold text-gray-900">Expense Tracker</span>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 p-4 space-y-1 overflow-y-auto">
          <router-link
            v-for="item in navigation"
            :key="item.name"
            :to="item.href"
            class="flex items-center gap-3 px-3 py-2 rounded-lg text-gray-700 hover:bg-gray-100 transition-colors"
            active-class="bg-blue-50 text-blue-700 hover:bg-blue-50"
          >
            <component :is="item.icon" class="h-5 w-5" />
            {{ item.name }}
          </router-link>
        </nav>

        <!-- User section -->
        <div class="p-4 border-t border-gray-200">
          <div class="mb-3 px-3">
            <p class="text-sm font-medium text-gray-900">{{ authStore.userName }}</p>
            <p class="text-xs text-gray-500">{{ authStore.userEmail }}</p>
          </div>
          <button
            @click="handleLogout"
            class="flex items-center gap-3 w-full px-3 py-2 rounded-lg text-red-600 hover:bg-red-50 transition-colors"
          >
            <LogOut class="h-5 w-5" />
            Sign out
          </button>
        </div>
      </div>
    </aside>

    <!-- Main content -->
    <main class="lg:pl-64">
      <div class="pt-14 lg:pt-0">
        <div class="p-4 lg:p-8">
          <slot />
        </div>
      </div>
    </main>
  </div>
</template>
