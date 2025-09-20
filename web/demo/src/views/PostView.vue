<template>
  <div class="post-detail">
    <div class="post-detail__wrapper">
      <div class="post-detail__left">
        <h1 class="post-detail__title">{{ post.title }}</h1>

        <div class="post-detail__tags">
          <TagItem
              class="tag"
              v-for="tag in post.tags"
              :key="tag"
              :value="tag"
          />
        </div>

        <p class="post-detail__type">Author: {{ post.author_id }}</p>
        <p class="post-detail__type">Type: {{ post.post_type }}</p>

        <div class="post-detail__description">
          <h2>Description</h2>
          <p>{{ post.description }}</p>
        </div>

        <div class="post-detail__contents">
          <h2>Contents</h2>
          <div
              v-for="content in post.contents"
              :key="content.content_data"
              class="post-detail__content-item"
          >
            <p>{{ ContentTypes[content.content_type] }}</p>
            <div v-if="content.is_link">
              <AppButton
                  class="post-detail__content-btn"
                  @click="openLink(content.content_data)"
              >
                Link
              </AppButton>
            </div>
            <div v-else>
              <AppButton variant="primary" @click="openModal(content.content_data)">
                Show
              </AppButton>
            </div>
          </div>
        </div>
      </div>

      <div class="post-detail__right">
        <img :src="post.preview_url" alt="preview" class="post-detail__image" />

        <PostInteractions
            :post="post"
            :isLoggedIn="isLoggedIn"
            @favorite="handleFavorite"
            @rateUpvote="handleRateUpvote"
            @rateDownvote="handleRateDownvote"
        />

        <AppButton
            v-if="post.author_id === auth.userId"
            class="post-detail__edit-btn"
            @click="goToEdit"
        >
          Edit Post
        </AppButton>

        <p class="post-detail__stats">
          Rating: {{ post.rating }} |
          Favorites: {{ post.favorites_count }} |
          Comments: {{ post.comments_count }}
        </p>

        <div class="post-detail__dates">
          <span>Created: {{ post.created_at }}</span>
          <span>Updated: {{ post.updated_at }}</span>
        </div>
      </div>
    </div>
  </div>
  <CommentsSection v-if="post.id" :post-id="post.id"/>
  <!-- Модальное окно для отображения контента (если открыто) -->
  <ContentModal
      v-if="modalVisible"
      :content="modalContent"
      @close="modalVisible = false"
  />
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { auth } from '../store/auth';
import {getPost, favoritePost, unfavoritePost, ratePost, ContentTypes} from '../sdk';
import type { Post as PostType } from '../sdk/types';
import TagItem from '../components/TagItem.vue';
import PostInteractions from '../components/PostInteractions.vue';
import router from '../router';
import AppButton from '../components/AppButton.vue';
import CommentsSection from './CommentsSection.vue';
import ContentModal from "../components/ContentModal.vue";

const props = defineProps<{ postId: string }>();

const post = ref<PostType>({
  id: 0,
  author_id: '',
  title: '',
  description: '',
  preview_url: '',
  post_type: '',
  tags: [],
  contents: [],
  created_at: '',
  updated_at: '',
  moderation_data: { status: 'pending', note: '' },
  interaction_data: { is_favorite: false, vote: 'none' },
  rating: 0,
  comments_count: 0,
  favorites_count: 0,
});

onMounted(async () => {
  post.value = await getPost(props.postId);
});

const goToEdit = () => {
  router.push(`/posts/${post.value.id}/edit`);
};

const isLoggedIn = computed(() => !!auth.userId);

const openLink = (url: string) => {
  window.open(url, '_blank');
};

const handleFavorite = async () => {
  if (post.value.interaction_data.is_favorite) {
    await unfavoritePost(post.value.id);
    post.value.interaction_data.is_favorite = false;
    post.value.favorites_count = Math.max(post.value.favorites_count - 1, 0);
  } else {
    await favoritePost(post.value.id);
    post.value.interaction_data.is_favorite = true;
    post.value.favorites_count += 1;
  }
};

const handleRateUpvote = async () => {
  await ratePost(
      post.value.id,
      post.value.interaction_data.vote === 'upvoted' ? 'retract' : 'upvote'
  );
  if (post.value.interaction_data.vote === 'upvoted') {
    post.value.interaction_data.vote = 'none';
    post.value.rating -= 1;
  } else if (post.value.interaction_data.vote === 'none') {
    post.value.interaction_data.vote = 'upvoted';
    post.value.rating += 1;
  } else if (post.value.interaction_data.vote === 'downvoted') {
    post.value.interaction_data.vote = 'upvoted';
    post.value.rating += 2;
  }
};

const handleRateDownvote = async () => {
  await ratePost(
      post.value.id,
      post.value.interaction_data.vote === 'downvoted' ? 'retract' : 'downvote'
  );
  if (post.value.interaction_data.vote === 'downvoted') {
    post.value.interaction_data.vote = 'none';
    post.value.rating += 1;
  } else if (post.value.interaction_data.vote === 'none') {
    post.value.interaction_data.vote = 'downvoted';
    post.value.rating -= 1;
  } else if (post.value.interaction_data.vote === 'upvoted') {
    post.value.interaction_data.vote = 'downvoted';
    post.value.rating -= 2;
  }
};

const modalVisible = ref(false);
const modalContent = ref('');

const openModal = (content: string) => {
  modalContent.value = content;
  modalVisible.value = true;
};
</script>

<style lang="scss" scoped>
$primary-color: #5a9;
$border-color: #444;
$background-color: #3b3b3b;
$text-color: #ccc;
$title-color: #fff;
$active-color: #ffa500;
$disabled-opacity: 1;

.post-detail {
  padding: 20px;
  color: $text-color;

  &__wrapper {
    display: flex;
    gap: 30px;
  }

  &__left {
    flex: 0.6;
    display: flex;
    flex-direction: column;

    .post-detail__title {
      font-size: 1.8rem;
      color: $title-color;
      margin: 0 0 10px;
    }

    .post-detail__tags {
      margin: 0 0 15px;
      display: flex;
      gap: 5px;
    }

    .post-detail__type {
      color: lighten($text-color, 10%);
      font-style: italic;
      margin: 0 0 20px;
    }

    .post-detail__description {
      margin-bottom: 20px;

      h2 {
        margin: 0 0 10px;
        color: $title-color;
      }

      p {
        margin: 0;
        line-height: 1.4;
      }
    }

    .post-detail__contents {
      background: lighten($background-color, 3%);
      padding: 10px;
      border-radius: 4px;

      h2 {
        margin-top: 0;
        margin-bottom: 10px;
        color: $title-color;
      }

      &__content-item {
        border-top: 1px solid $border-color;
        padding-top: 10px;
        margin-top: 10px;

        &:first-of-type {
          border-top: none;
          margin-top: 0;
          padding-top: 0;
        }

        .post-detail__content-btn {
          background: $primary-color;
          color: #fff;
          border: none;
          padding: 4px 8px;
          border-radius: 4px;
          cursor: pointer;
          transition: background 0.2s;

          &:hover {
            background: lighten($primary-color, 10%);
          }
        }
      }
    }
  }

  &__right {
    flex: 0.4;
    display: flex;
    flex-direction: column;
    align-items: flex-start;

    .post-detail {
      &__image {
        width: 100%;
        max-width: 400px;
        border-radius: 4px;
        object-fit: cover;
        margin-bottom: 15px;
      }

      &__edit-btn {
        margin-top: 15px;
      }

      &__stats {
        font-size: 0.9rem;
        color: lighten($text-color, 15%);
        margin-bottom: 15px;
      }

      &__dates {
        display: flex;
        flex-direction: column;
        gap: 3px;
        font-size: 0.8rem;
        color: lighten($text-color, 10%);
      }
    }
  }
}
</style>
