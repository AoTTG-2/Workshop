<template>
  <div class="modal-backdrop" @click.self="close">
    <div class="modal">
      <div class="modal__box">
        <textarea ref="textRef" class="modal__text" readonly>{{ content }}</textarea>
      </div>
      <div class="modal__actions">
        <AppButton variant="primary" @click="copyContent">
          Copy
        </AppButton>
        <AppButton variant="secondary" @click="close">
          Close
        </AppButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits, ref, onMounted } from 'vue';
import AppButton from './AppButton.vue';

const props = defineProps<{
  content: string
}>();

const emits = defineEmits<{
  (e: 'close'): void;
}>();

const close = () => {
  emits('close');
};

const copyContent = async () => {
  try {
    await navigator.clipboard.writeText(props.content);
  } catch (error) {
    console.error('Copy failed', error);
  }
};

const textRef = ref<HTMLTextAreaElement | null>(null);

onMounted(() => {
  textRef.value?.focus();
});
</script>

<style lang="scss" scoped>
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: #171717;
  border-radius: 6px;
  padding: 20px;
  width: 400px;
  max-width: 90%;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  outline: none;

  &__box {
    width: 100%;
    height: 300px;
    margin-bottom: 15px;
  }

  &__text {
    width: 100%;
    height: 100%;
    resize: none;
    border: 1px solid #77bbad;
    border-radius: 4px;
    background: #464646;
    color: #fff;
    font-family: inherit;
    font-size: 0.9rem;
    padding: 5px;
    box-sizing: border-box;
    overflow: auto;
  }

  &__actions {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  }
}
</style>
