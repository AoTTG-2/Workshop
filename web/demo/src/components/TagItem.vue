<template>
  <span class="tag-item">#{{ value }}</span>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'

defineProps<{ value: string }>()
</script>

<style lang="scss" scoped>
.tag-item {
  background: lighten(#3b3b3b, 10%);
  padding: 4px 8px;
  border-radius: 4px;
  color: lighten(#ccc, 10%);
  font-size: 0.85rem;
}
</style>
