<template>
  <div class="comment-actions">
    <button class="comment-actions__button" @click="emit('edit')">
      <EditIcon class="comment-actions__icon"/>
    </button>
    <button class="comment-actions__button" @click="emit('delete')">
      <DeleteIcon class="comment-actions__icon"/>
    </button>
  </div>
</template>

<script setup lang="ts">
import EditIcon from './icons/EditIcon.vue';
import DeleteIcon from './icons/DeleteIcon.vue';

const emit = defineEmits<{
  edit: [];
  delete: [];
}>()
</script>

<style lang="scss" scoped>
.comment-actions {
  display: flex;
  gap: 5px;

  &__button {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 4px;
    display: flex;
    align-items: center;
    justify-content: center;

    &:hover {
      opacity: 0.8;
    }
  }

  &__icon {
    width: 20px;
    height: 20px;
    fill: #fff;
  }
}
</style>
