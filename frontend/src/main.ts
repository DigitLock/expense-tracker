import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import { useAuthStore } from './stores/auth'

// vue-sonner base styles (provides the toaster z-index: 999999999 so toasts
// render above shadcn dialogs at z-50). Imported BEFORE ./style.css so our
// custom positioning overrides win without touching z-index.
import 'vue-sonner/style.css'
import './style.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

// Initialize auth store before mounting
const authStore = useAuthStore()
authStore.initialize()

app.mount('#app')
