<template>
  <div class="login-modal">
    <div class="login-modal__content">
      <h2 class="login-modal__title">Debug Login</h2>
      <form @submit.prevent="submitLogin" class="login-modal__form">
        <label class="login-modal__label">
          User ID:
          <input type="text" v-model="userId" class="login-modal__input" required/>
        </label>
        <label class="login-modal__label">
          User Roles:
          <input type="text" v-model="userRolesInput" class="login-modal__input" required/>
        </label>
        <div class="login-modal__buttons">
          <AppButton type="submit" class="login-modal__btn">Login</AppButton>
          <AppButton type="button" class="login-modal__btn" @click="close" variant="secondary">Cancel</AppButton>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import AppButton from './AppButton.vue'

const emit = defineEmits<{
  close: [],
  login: [{ userId: string, userRoles: string[] }]
}>()

const userId = ref('')
const userRolesInput = ref('')

const submitLogin = () => {
  const roles = userRolesInput.value
      .split(',')
      .map(r => r.trim())
      .filter(r => r)
  emit('login', {userId: userId.value, userRoles: roles})
}

const close = () => {
  emit('close')
}
</script>

<style lang="scss" scoped>
$overlay-bg: rgba(0, 0, 0, 0.7);
$modal-bg: #2b2b2b;
$text-color: #f0f0f0;
$label-color: #ccc;
$field-bg: #3b3b3b;
$border-color: #444;
$box-shadow-color: rgba(0, 0, 0, 0.6);

.login-modal {
  position: fixed;
  inset: 0;
  background: $overlay-bg;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;

  &__content {
    background: $modal-bg;
    padding: 24px;
    border-radius: 8px;
    width: 320px;
    box-shadow: 0 8px 20px $box-shadow-color;
  }

  &__title {
    margin: 0 0 20px;
    color: $text-color;
    text-align: center;
    font-size: 1.4rem;
  }

  &__form {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  &__label {
    display: flex;
    flex-direction: column;
    color: $label-color;
    font-size: 0.9rem;
  }

  &__input {
    padding: 8px;
    margin-top: 5px;
    border: 1px solid $border-color;
    border-radius: 4px;
    background: $field-bg;
    color: $text-color;
    font-size: 0.9rem;
    transition: border-color 0.2s;

    &:focus {
      outline: none;
      border-color: lighten($border-color, 15%);
    }
  }

  &__buttons {
    display: flex;
    gap: 10px;
    margin-top: 10px;
  }

  &__btn {
    flex: 1;
    padding: 10px 15px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: background 0.2s;
  }
}
</style>
