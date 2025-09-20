<template>
  <svg viewBox="0 0 24 24" class="icon">
    <path
        v-if="active"
        fill="currentColor"
        d="M1 3h4v12H1V3zm22 11.5c0 .83-.67 1.5-1.5 1.5h-6.31l.95 4.57.03.32c0 .41-.17.79-.44 1.06L14 22l-7-6v-11c0-1.1.9-2 2-2h9c.83 0 1.54.5 1.84 1.22l3.02 7.05c.09.23.14.47.14.73v1.5z"
    />
    <path
        v-else
        fill="none"
        stroke="currentColor"
        d="M1 3h4v12H1V3zm22 11.5c0 .83-.67 1.5-1.5 1.5h-6.31l.95 4.57.03.32c0 .41-.17.79-.44 1.06L14 22l-7-6v-11c0-1.1.9-2 2-2h9c.83 0 1.54.5 1.84 1.22l3.02 7.05c.09.23.14-.47.14-.73v-1.5z"
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
