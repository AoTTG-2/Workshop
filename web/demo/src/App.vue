<template>
  <AppHeader
      @login="openLogin"
      @logout="handleLogout"
  />

  <main class="main">
    <router-view />
  </main>

  <DebugLoginModal
      v-if="showLogin"
      @close="closeLogin"
      @login="handleLogin"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppHeader from './components/AppHeader.vue'
import DebugLoginModal from './components/DebugLoginModal.vue'
import { setAuth, logout } from './store/auth'
import { setDebugAuth } from './sdk'

const showLogin = ref(false)
const router = useRouter()

function openLogin() {
  showLogin.value = true
}

function closeLogin() {
  showLogin.value = false
}

function handleLogin(payload: { userId: string; userRoles: string[] }) {
  setAuth(payload.userId, payload.userRoles)
  setDebugAuth(payload.userId, payload.userRoles)
  showLogin.value = false
  router.go(0)
}

function handleLogout() {
  logout()
  router.go(0)
}
</script>

<style scoped lang="scss">
.main {
  padding: 20px;
}
</style>
