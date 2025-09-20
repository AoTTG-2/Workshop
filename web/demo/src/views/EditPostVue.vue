<template>
  <div class="edit-post">
    <h1 class="edit-post__title">Edit Post</h1>

    <form @submit.prevent="onSubmit" class="edit-post__form">
      <label class="edit-post__label">
        Title:
        <input
            type="text"
            v-model="editData.title"
            class="edit-post__input"
            required
        />
      </label>

      <label class="edit-post__label">
        Description:
        <textarea
            v-model="editData.description"
            class="edit-post__textarea"
            rows="3"
        ></textarea>
      </label>

      <label class="edit-post__label">
        Preview URL:
        <input
            type="text"
            v-model="editData.preview_url"
            class="edit-post__input"
        />
      </label>

      <label class="edit-post__label">
        Type:
        <select
            v-model="editData.type"
            class="edit-post__select"
            required
        >
          <option disabled value="">Select Type</option>
          <option
              v-for="(text, key) in PostTypes"
              :key="key"
              :value="key"
          >
            {{ text }}
          </option>
        </select>
      </label>

      <label class="edit-post__label">
        Tags (max 5):
        <input
            type="text"
            v-model="tagsInput"
            class="edit-post__input"
        />
        <small>(comma separated)</small>
      </label>

      <div class="edit-post__contents">
        <h2>Contents</h2>
        <div
            v-for="(content, index) in editData.contents"
            :key="index"
            class="edit-post__content-item"
        >
          <label class="edit-post__label-inline">
            Type:
            <select
                v-model="content.content_type"
                class="edit-post__select"
                required
            >
              <option disabled value="">Select Type</option>
              <option
                  v-for="(text, key) in ContentTypes"
                  :key="key"
                  :value="key"
              >
                {{ text }}
              </option>
            </select>
          </label>
          <label class="edit-post__label-inline">
            Data:
            <input
                type="text"
                v-model="content.content_data"
                class="edit-post__input"
                placeholder="URL or text"
            />
          </label>
          <label class="edit-post__checkbox-label">
            Is Link?
            <input
                type="checkbox"
                v-model="content.is_link"
                class="edit-post__checkbox"
            />
          </label>

          <button
              v-if="editData.contents.length > 1"
              type="button"
              class="edit-post__remove-btn"
              @click="removeContent(index)"
          >
            Remove
          </button>
        </div>

        <button
            type="button"
            class="edit-post__add-content-btn"
            @click="addContent"
        >
          + Add Content
        </button>
      </div>

      <button type="submit" class="edit-post__submit-btn">
        Save Changes
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {ContentTypes, getPost, updatePost} from '../sdk'
import {PostTypes} from '../sdk'
import type {UpdatePostRequest, Post} from '../sdk'

const route = useRoute()
const router = useRouter()
const postId = route.params.postId

const editData = ref<UpdatePostRequest>({
  title: '',
  description: '',
  preview_url: '',
  type: '',
  tags: [],
  contents: []
})

const tagsInput = ref('')

onMounted(async () => {
  const existing: Post = await getPost(Number(postId))

  editData.value = {
    title: existing.title,
    description: existing.description,
    preview_url: existing.preview_url,
    type: existing.post_type,
    tags: existing.tags,
    contents: existing.contents.map((c) => ({
      id: c.id,
      content_type: c.content_type,
      content_data: c.content_data,
      is_link: c.is_link
    }))
  }

  tagsInput.value = existing.tags.join(', ')
})

watch(tagsInput, (val) => {
  const rawTags = val.split(',')
      .map((t) => t.trim())
      .filter(Boolean)
  if (rawTags.length > 5) rawTags.splice(5)
  editData.value.tags = rawTags
})

function addContent() {
  editData.value.contents.push({
    content_type: '',
    content_data: '',
    is_link: true
  })
}

function removeContent(index: number) {
  if (editData.value.contents.length > 1) {
    editData.value.contents.splice(index, 1)
  }
}

async function onSubmit() {
  if (!editData.value.type) {
    alert('Please select a post type.')
    return
  }
  if (editData.value.contents.length === 0) {
    alert('At least one content is required.')
    return
  }

  try {
    const updated = await updatePost(Number(postId), editData.value)
    await router.push(`/posts/${updated.id}`)
  } catch (error) {
    console.error(error)
    alert('Failed to update post.')
  }
}
</script>

<style scoped lang="scss">
.edit-post {
  max-width: 600px;
  margin: 0 auto;

  &__title {
    margin-bottom: 20px;
    font-size: 1.4rem;
    color: #fff;
  }

  &__form {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  &__label {
    display: flex;
    flex-direction: column;
    color: #fff;
    font-size: 0.9rem;
  }

  &__label-inline {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    color: #fff;
    font-size: 0.9rem;
    margin-right: 10px;
  }

  &__checkbox-label {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    margin-top: 5px;
    color: #fff;
  }

  &__input,
  &__select,
  &__textarea {
    margin-top: 5px;
    padding: 8px;
    border: 1px solid #444;
    border-radius: 4px;
    background-color: #3b3b3b;
    color: #fff;
    font-size: 0.9rem;
  }

  &__contents {
    background: #2b2b2b;
    padding: 10px;
    border-radius: 6px;

    h2 {
      color: #fff;
      margin-top: 0;
      margin-bottom: 10px;
    }
  }

  &__content-item {
    margin-bottom: 10px;
  }

  &__remove-btn,
  &__add-content-btn,
  &__submit-btn {
    padding: 8px 12px;
    background: #5a9;
    border: none;
    border-radius: 4px;
    color: #fff;
    cursor: pointer;
    transition: background 0.2s;
    margin-top: 5px;

    &:hover {
      background: lighten(#5a9, 10%);
    }
  }

  &__add-content-btn {
    display: block;
    margin-top: 10px;
  }

  &__submit-btn {
    align-self: flex-start;
    margin-top: 10px;
  }
}
</style>
