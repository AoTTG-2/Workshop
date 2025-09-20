<template>
  <div class="post-interactions__buttons">
    <button
        :disabled="!isLoggedIn"
        @click="onFavorite"
        :class="['post-interactions__btn', { active: post.interaction_data.is_favorite }]"
    >
      <FavoriteIcon :active="post.interaction_data.is_favorite"/>
    </button>
    <button
        :disabled="!isLoggedIn"
        @click="onRateUpvote"
        :class="['post-interactions__btn', { active: post.interaction_data.vote === 'upvoted' }]"
    >
      <UpvoteIcon :active="post.interaction_data.vote === 'upvoted'"/>
    </button>
    <button
        :disabled="!isLoggedIn"
        @click="onRateDownvote"
        :class="['post-interactions__btn', { active: post.interaction_data.vote === 'downvoted' }]"
    >
      <DownvoteIcon :active="post.interaction_data.vote === 'downvoted'"/>
    </button>
  </div>
</template>

<script setup lang="ts">
import {defineProps, defineEmits} from 'vue'
import type {Post as PostType} from '../sdk/types'
import FavoriteIcon from './icons/FavoriteIcon.vue'
import UpvoteIcon from './icons/UpvoteIcon.vue'
import DownvoteIcon from './icons/DownvoteIcon.vue'

const props = defineProps<{
  post: PostType
  isLoggedIn: boolean
}>()

const emit = defineEmits<{
  (e: 'favorite', post: PostType): void
  (e: 'rateUpvote', post: PostType): void
  (e: 'rateDownvote', post: PostType): void
}>()

function onFavorite() {
  emit('favorite', props.post)
}

function onRateUpvote() {
  emit('rateUpvote', props.post)
}

function onRateDownvote() {
  emit('rateDownvote', props.post)
}
</script>

<style scoped lang="scss">
$post-interactions-bg: #3b3b3b;
$active-color: #ffa500;

.post-interactions {
  display: flex;
  flex-direction: column;
  gap: 10px;

  &__buttons {
    display: flex;
    gap: 10px;

    .post-interactions__btn {
      background: none;
      border: none;
      cursor: pointer;
      padding: 0;
      display: flex;
      align-items: center;
      transition: transform 0.2s ease, opacity 0.2s ease;
      position: relative;
      color: #ccc;

      &:hover {
        transform: scale(1.1);
      }

      &:disabled {
        cursor: not-allowed;
        color: #7a7a7a;

        &::after {
          content: "Login required";
          position: absolute;
          top: -25px;
          left: 50%;
          transform: translateX(-50%);
          background-color: rgba(0, 0, 0, 0.8);
          color: #fff;
          padding: 3px 6px;
          border-radius: 4px;
          font-size: 0.75rem;
          opacity: 0;
          pointer-events: none;
          white-space: nowrap;
          transition: opacity 0.2s ease;
        }

        &:hover::after {
          opacity: 1;
        }
      }

      &.active {
        transform: scale(1.15);
        color: $active-color;
      }

      svg {
        width: 24px;
        height: 24px;
      }
    }
  }
}
</style>
