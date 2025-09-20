<template>
  <button
      :type=type
      :class="['app-button', 'app-button--' + variant]"
  >
    <slot/>
  </button>
</template>

<script setup lang="ts">
import {defineProps} from 'vue'
import type {PropType} from 'vue'

defineProps({
  variant: {
    type: String as PropType<'primary' | 'secondary'>,
    default: 'primary'
  },
  type: {
    type: String as PropType<"button" | "submit" | "reset">,
    default: 'button'
  }
})
</script>

<style scoped lang="scss">
$color-primary: #5a9;
$color-secondary: #999;
$text-color: #fff;

.app-button {
  padding: 8px 12px;
  border: none;
  border-radius: 4px;
  color: $text-color;
  cursor: pointer;
  transition: background 0.2s;

  &--primary {
    background: $color-primary;

    &:hover {
      background: lighten($color-primary, 10%);
    }
  }

  &--secondary {
    background: $color-secondary;

    &:hover {
      background: lighten($color-secondary, 10%);
    }
  }
}
</style>
