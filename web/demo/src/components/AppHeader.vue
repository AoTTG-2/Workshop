<template>
  <header class="header">
    <nav class="header__nav">
      <router-link to="/posts" class="header__link">Posts</router-link>

      <AppButton
          class="header__create-btn"
          :class="{ disabled: !isLoggedIn }"
          @click="onCreateClick"
      >
        Create Post
      </AppButton>

      <AppButton
          v-if="!auth.userId"
          @click="emit('login')"
      >
        Login
      </AppButton>

      <span v-else class="header__user">User: {{ auth.userId }}</span>

      <AppButton
          v-if="auth.userId"
          @click="$emit('logout')"
      >
        Logout
      </AppButton>
    </nav>
  </header>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {auth} from '../store/auth'
import AppButton from './AppButton.vue'
import {useRouter} from 'vue-router'

const emit = defineEmits<{
  login: [],
  logout: []
}>()

const isLoggedIn = computed(() => !!auth.userId)

const router = useRouter()

function onCreateClick() {
  if (!isLoggedIn.value) {
    return
  }
  router.push('/create')
}
</script>

<style scoped lang="scss">
.header {
  background-color: #333;
  padding: 10px 20px;

  &__nav {
    display: flex;
    gap: 15px;
    align-items: center;
  }

  &__link {
    color: #fff;
    text-decoration: none;
  }

  &__user {
    color: #fff;
    font-size: 0.9rem;
  }

  .header__create-btn {
    &.disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  }
}
</style>
