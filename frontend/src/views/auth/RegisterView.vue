<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { UserPlus, Mail, Lock, User, Users, AlertCircle } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const name = ref('')
const familyName = ref('')

// Whether the user has attempted to submit; client errors are shown afterwards.
const submitted = ref(false)
// Server-provided field errors (keyed by json field name), cleared on edit/submit.
const serverErrors = reactive<Record<string, string>>({})

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// Live client-side validation used both to disable submit and to show messages.
const clientErrors = computed<Record<string, string>>(() => {
  const errs: Record<string, string> = {}
  if (!email.value.trim()) {
    errs.email = 'Email is required'
  } else if (!emailRegex.test(email.value.trim())) {
    errs.email = 'Invalid email format'
  }
  if (!password.value) {
    errs.password = 'Password is required'
  } else if (password.value.length < 8) {
    errs.password = 'Password must be at least 8 characters'
  }
  if (confirmPassword.value !== password.value) {
    errs.confirm_password = 'Passwords do not match'
  }
  if (!name.value.trim()) {
    errs.name = 'Name is required'
  }
  return errs
})

const isFormValid = computed(() => Object.keys(clientErrors.value).length === 0)

// Merged view: server errors take precedence, then client errors (once submitted).
function fieldError(field: string): string | undefined {
  if (serverErrors[field]) return serverErrors[field]
  if (submitted.value) return clientErrors.value[field]
  return undefined
}

const familyPlaceholder = computed(() => `${name.value.trim() || 'Your name'}'s family`)

function clearServerError(field: string) {
  if (serverErrors[field]) delete serverErrors[field]
}

async function handleSubmit() {
  submitted.value = true
  // Drop any stale server errors before re-validating.
  Object.keys(serverErrors).forEach((k) => delete serverErrors[k])

  if (!isFormValid.value) return

  const result = await authStore.register({
    email: email.value.trim(),
    password: password.value,
    name: name.value.trim(),
    family_name: familyName.value.trim() || undefined,
  })

  if (result.ok) {
    await router.push('/')
    return
  }

  // 409 duplicate email -> show under the email field.
  if (result.code === 'EMAIL_ALREADY_EXISTS') {
    serverErrors.email = 'This email is already registered'
  }
  // 400 validation -> spread details onto their fields (json field names).
  for (const detail of result.fieldErrors) {
    serverErrors[detail.field] = detail.message
  }
}
</script>

<template>
  <div class="bg-white rounded-xl shadow-lg p-8">
    <!-- Header -->
    <div class="text-center mb-8">
      <h1 class="text-2xl font-bold text-gray-900">Create your account</h1>
      <p class="text-gray-500 mt-2">Start tracking your family finances</p>
    </div>

    <!-- General error message -->
    <div
      v-if="authStore.error"
      class="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-start gap-3"
    >
      <AlertCircle class="h-5 w-5 text-red-500 flex-shrink-0 mt-0.5" />
      <p class="text-sm text-red-700">{{ authStore.error }}</p>
    </div>

    <!-- Form -->
    <form @submit.prevent="handleSubmit" class="space-y-5" novalidate>
      <!-- Email -->
      <div>
        <label for="email" class="block text-sm font-medium text-gray-700 mb-1">Email</label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Mail class="h-5 w-5 text-gray-400" />
          </div>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="you@example.com"
            autocomplete="email"
            @input="clearServerError('email')"
            class="block w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
            :class="fieldError('email') ? 'border-red-300' : 'border-gray-300'"
          />
        </div>
        <p v-if="fieldError('email')" class="mt-1 text-sm text-red-600">{{ fieldError('email') }}</p>
      </div>

      <!-- Name -->
      <div>
        <label for="name" class="block text-sm font-medium text-gray-700 mb-1">Name</label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <User class="h-5 w-5 text-gray-400" />
          </div>
          <input
            id="name"
            v-model="name"
            type="text"
            placeholder="Igor Kudinov"
            autocomplete="name"
            @input="clearServerError('name')"
            class="block w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
            :class="fieldError('name') ? 'border-red-300' : 'border-gray-300'"
          />
        </div>
        <p v-if="fieldError('name')" class="mt-1 text-sm text-red-600">{{ fieldError('name') }}</p>
      </div>

      <!-- Family name (optional) -->
      <div>
        <label for="family_name" class="block text-sm font-medium text-gray-700 mb-1">
          Family name <span class="text-gray-400 font-normal">(optional)</span>
        </label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Users class="h-5 w-5 text-gray-400" />
          </div>
          <input
            id="family_name"
            v-model="familyName"
            type="text"
            :placeholder="familyPlaceholder"
            @input="clearServerError('family_name')"
            class="block w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
            :class="fieldError('family_name') ? 'border-red-300' : 'border-gray-300'"
          />
        </div>
        <p v-if="fieldError('family_name')" class="mt-1 text-sm text-red-600">
          {{ fieldError('family_name') }}
        </p>
      </div>

      <!-- Password -->
      <div>
        <label for="password" class="block text-sm font-medium text-gray-700 mb-1">Password</label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Lock class="h-5 w-5 text-gray-400" />
          </div>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="At least 8 characters"
            autocomplete="new-password"
            @input="clearServerError('password')"
            class="block w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
            :class="fieldError('password') ? 'border-red-300' : 'border-gray-300'"
          />
        </div>
        <p v-if="fieldError('password')" class="mt-1 text-sm text-red-600">
          {{ fieldError('password') }}
        </p>
      </div>

      <!-- Confirm password -->
      <div>
        <label for="confirm_password" class="block text-sm font-medium text-gray-700 mb-1">
          Confirm password
        </label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Lock class="h-5 w-5 text-gray-400" />
          </div>
          <input
            id="confirm_password"
            v-model="confirmPassword"
            type="password"
            placeholder="Re-enter your password"
            autocomplete="new-password"
            class="block w-full pl-10 pr-3 py-2.5 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors"
            :class="fieldError('confirm_password') ? 'border-red-300' : 'border-gray-300'"
          />
        </div>
        <p v-if="fieldError('confirm_password')" class="mt-1 text-sm text-red-600">
          {{ fieldError('confirm_password') }}
        </p>
      </div>

      <!-- Submit button -->
      <button
        type="submit"
        :disabled="authStore.loading || !isFormValid"
        class="w-full flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 focus:ring-4 focus:ring-blue-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        <UserPlus v-if="!authStore.loading" class="h-5 w-5" />
        <svg
          v-else
          class="animate-spin h-5 w-5 fill-none"
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
        >
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        {{ authStore.loading ? 'Creating account...' : 'Create account' }}
      </button>
    </form>

    <!-- Link to login -->
    <p class="mt-6 text-center text-sm text-gray-600">
      Already have an account?
      <RouterLink to="/login" class="font-medium text-blue-600 hover:text-blue-700">
        Sign in
      </RouterLink>
    </p>
  </div>
</template>
