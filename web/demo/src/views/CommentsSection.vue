<template>
  <div class="comments-section">
    <h2 class="comments-section__title">Comments</h2>

    <div v-if="isUserLoggedIn" class="comments-section__form-wrapper">
      <form @submit.prevent="submitComment" class="comments-section__form">
        <textarea
            v-model="commentText"
            class="comments-section__textarea"
            placeholder="Write a comment..."
            required
            @keydown="handleTextareaKeydown"
        />
        <AppButton variant="primary" type="submit">
          Post Comment
        </AppButton>
      </form>
    </div>
    <div v-else class="comments-section__login-prompt">
      <p>Please log in to comment.</p>
    </div>

    <div v-if="commentsList.length" class="comments-section__list">
      <article
          v-for="comment in commentsList"
          :key="comment.id"
          class="comments-section__item"
          @mouseenter="showActions(comment.id)"
          @mouseleave="hideActions(comment.id)"
      >
        <header class="comments-section__item-header">
          <span class="comments-section__item-author">{{ comment.author_id }}</span>
          <span class="comments-section__item-date">{{ formatDate(comment.created_at) }}</span>
        </header>
        <div class="comments-section__item-content">
          <p>{{ comment.content }}</p>
        </div>
        <div
            v-if="comment.author_id === auth.userId && activeActionCommentId === comment.id"
            class="comments-section__item-actions"
        >
          <CommentActionsIcons
              @edit="startEditingComment(comment)"
              @delete="removeComment(comment.id)"
          />
        </div>

        <div v-if="editingCommentId === comment.id" class="comments-section__item-edit">
          <textarea
              v-model="editedCommentText"
              class="comments-section__item-edit-textarea"
          ></textarea>
          <div class="comments-section__item-edit-buttons">
            <AppButton variant="primary" @click="submitCommentEdit(comment.id)">
              Save
            </AppButton>
            <AppButton variant="secondary" @click="cancelEditing">
              Cancel
            </AppButton>
          </div>
        </div>
      </article>
    </div>

    <div class="comments-section__load-more">
      <AppButton
          variant="primary"
          @click="fetchComments"
          v-if="hasMoreComments"
      >
        Load More
      </AppButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, defineProps} from 'vue';
import {auth} from '../store/auth';
import {getComments, addComment, updateComment, deleteComment, type Comment} from '../sdk';
import AppButton from '../components/AppButton.vue';
import CommentActionsIcons from '../components/CommentActionsIcons.vue';

const props = defineProps<{ postId: number }>();

const commentsList = ref<Comment[]>([]);
const commentText = ref('');
const editingCommentId = ref<number | null>(null);
const editedCommentText = ref('');
const page = ref(1);
const commentsPerPage = ref(10);
const hasMoreComments = ref(true);
const activeActionCommentId = ref<number | null>(null);

const isUserLoggedIn = computed(() => !!auth.userId);

const fetchComments = async () => {
  if (!hasMoreComments.value) return;
  const newComments = await getComments(props.postId, page.value, commentsPerPage.value);
  if (newComments.length < commentsPerPage.value) {
    hasMoreComments.value = false;
  }
  commentsList.value.push(...newComments);
  page.value++;
};

const submitComment = async () => {
  const newComment = await addComment(commentText.value, props.postId);
  commentsList.value.unshift(newComment);
  commentText.value = '';
};

const startEditingComment = (comment: Comment) => {
  editingCommentId.value = comment.id;
  editedCommentText.value = comment.content;
};

const submitCommentEdit = async (commentId: number) => {
  const updatedComment = await updateComment(commentId, editedCommentText.value);
  const index = commentsList.value.findIndex(c => c.id === commentId);
  if (index !== -1) {
    commentsList.value[index] = updatedComment;
  }
  editingCommentId.value = null;
};

const removeComment = async (commentId: number) => {
  await deleteComment(commentId);
  commentsList.value = commentsList.value.filter(c => c.id !== commentId);
};

const cancelEditing = () => {
  editingCommentId.value = null;
};

const showActions = (commentId: number) => {
  activeActionCommentId.value = commentId;
};

const hideActions = (commentId: number) => {
  if (activeActionCommentId.value === commentId) {
    activeActionCommentId.value = null;
  }
};

const formatDate = (dateString: string): string => {
  const date = new Date(dateString);
  const dd = String(date.getDate()).padStart(2, '0');
  const mm = String(date.getMonth() + 1).padStart(2, '0');
  const yyyy = date.getFullYear();
  const hh = String(date.getHours()).padStart(2, '0');
  const min = String(date.getMinutes()).padStart(2, '0');
  return `${dd}/${mm}/${yyyy} ${hh}:${min}`;
};

const handleTextareaKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Enter' && !event.shiftKey) {
    event.preventDefault();
    submitComment();
  }
};

onMounted(() => {
  fetchComments();
});
</script>

<style lang="scss" scoped>
.comments-section {
  &__title {
    font-size: 1.5rem;
    margin-bottom: 15px;
  }

  &__form-wrapper {
    margin-bottom: 20px;
  }

  &__form {
    display: flex;
    flex-direction: column;

    & .comments-section__textarea {
      padding: 8px;
      border: 1px solid #555;
      border-radius: 4px;
      background: #444;
      color: #ddd;
      margin-bottom: 10px;
      resize: vertical;
    }
  }

  &__login-prompt {
    margin-bottom: 20px;
    color: #bbb;
  }

  &__list {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  &__item {
    position: relative;
    background: #333;
    padding: 15px;
    border-radius: 4px;
    transition: background 0.2s;

    &:hover {
      background: #444;
    }

    &-header {
      display: flex;
      align-items: center;
      margin-bottom: 8px;

      & > .comments-section__item-author {
        font-weight: bold;
        margin-right: 10px;
      }

      & > .comments-section__item-date {
        font-size: 0.85rem;
        color: #aaa;
      }
    }

    &-content {
      p {
        margin: 0;
      }
    }

    &-actions {
      position: absolute;
      top: 10px;
      right: 10px;
    }

    &-edit {
      margin-top: 10px;

      &-textarea {
        width: 100%;
        padding: 8px;
        border: 1px solid #555;
        border-radius: 4px;
        background: #444;
        color: #ddd;
        resize: vertical;
      }

      &-buttons {
        margin-top: 8px;
        display: flex;
        gap: 10px;
      }
    }
  }

  &__load-more {
    margin-top: 20px;
    text-align: center;
  }
}
</style>
