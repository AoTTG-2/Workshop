<template>
  <svg viewBox="0 0 24 24" class="icon">
    <path
        v-if="active"
        fill="currentColor"
        d="M12 17.27L18.18 21 16.54 13.97 22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z"
    />
    <path
        v-else
        fill="none"
        stroke="currentColor"
        d="M12 2l2.81 6.17L22 9.24l-5 4.87L18.18 21 12 17.27 5.82 21 7 14.11l-5-4.87 7.19-.61z"
    />
  </svg>
</template>

<script setup lang="ts">
defineProps<{ active: boolean }>();
</script>

<style scoped lang="scss">
.icon {
  width: 24px;
  height: 24px;
}
</style>
