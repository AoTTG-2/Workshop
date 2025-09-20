<template>
  <div class="post-card">
    <div class="post-card__content">
      <router-link :to="`/posts/${post.id}`" class="post-card__link">
        <h2 class="post-card__title">{{ post.title }}</h2>
        <div class="post-card__tags">
          <TagItem
              v-for="tag in post.tags"
              :key="tag"
              :value="tag"
          />
        </div>
        <p class="post-card__type">{{ PostTypes[post.post_type] }}</p>
        <p class="post-card__description">{{ truncatedDescription }}</p>
        <p class="post-card__moderation">
          Moderation: {{ post.moderation_data.status }}
          <span v-if="post.moderation_data.note">
            ({{ post.moderation_data.note }})
          </span>
        </p>
        <p class="post-card__interaction">
          Rating: {{ post.rating }} | Comments: {{ post.comments_count }} | Favorites: {{ post.favorites_count }}
        </p>
        <p class="post-card__dates">
          Created: {{ post.created_at }} | Updated: {{ post.updated_at }}
        </p>
      </router-link>
    </div>
    <div class="post-card__aside">
      <img :src="post.preview_url" alt="preview" class="post-card__image"/>
      <PostInteractions
          :post="post"
          :isLoggedIn="isLoggedIn"
          @favorite="handleFavorite"
          @rateUpvote="handleRateUpvote"
          @rateDownvote="handleRateDownvote"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import TagItem from './TagItem.vue'
import PostInteractions from './PostInteractions.vue'
import {type Post, PostTypes} from '../sdk'

const props = defineProps<{
  post: Post
  canInteract: boolean
}>()

const emit = defineEmits<{
  favorite: [post: Post]
  rateUpvote: [post: Post]
  rateDownvote: [post: Post]
}>()

const truncatedDescription = computed(() => {
  const maxLength = 64
  if (props.post.description.length > maxLength) {
    return props.post.description.slice(0, maxLength) + '...'
  }
  return props.post.description
})

const isLoggedIn = computed(() => {
  return props.post && props.canInteract
})

const handleFavorite = (post: Post) => {
  emit('favorite', post)
}

const handleRateUpvote = (post: Post) => {
  emit('rateUpvote', post)
}

const handleRateDownvote = (post: Post) => {
  emit('rateDownvote', post)
}
</script>

<style lang="scss" scoped>
$post-border-color: #444;
$post-bg: #3b3b3b;
$post-text: #ccc;
$post-title: #fff;
$post-active: #ffa500;

.post-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border: 1px solid $post-border-color;
  border-radius: 6px;
  background-color: $post-bg;
  padding: 15px;

  &__content {
    flex: 1;
    margin-right: 20px;

    .post-card__link {
      text-decoration: none;
      color: inherit;

      .post-card__title {
        font-size: 1.5rem;
        margin: 10px 0;
        color: $post-title;
      }

      .post-card__tags {
        display: flex;
        gap: 5px;
      }

      p {
        margin: 5px 0;
        color: $post-text;
      }
    }
  }

  &__aside {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  &__image {
    width: 300px;
    border-radius: 4px;
    object-fit: cover;
    margin-bottom: 10px;
  }
}
</style>
