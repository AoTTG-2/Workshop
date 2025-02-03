<template>
  <div class="create-post">
    <h1 class="create-post__title">Create a New Post</h1>

    <form @submit.prevent="onSubmit" class="create-post__form">
      <label class="create-post__label">
        Title:
        <input
            type="text"
            v-model="postData.title"
            class="create-post__input"
            required
            maxlength="120"
        />
      </label>

      <label class="create-post__label">
        Description:
        <textarea
            v-model="postData.description"
            class="create-post__textarea"
            rows="3"
            required
        ></textarea>
      </label>

      <label class="create-post__label">
        Preview URL:
        <input
            type="text"
            v-model="postData.preview_url"
            class="create-post__input"
        />
      </label>

      <label class="create-post__label">
        Type:
        <select
            v-model="postData.type"
            class="create-post__select"
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

      <label class="create-post__label">
        Tags (max 5):
        <input
            type="text"
            v-model="tagsInput"
            class="create-post__input"
        />
        <small>(comma separated, up to 5)</small>
      </label>

      <div class="create-post__contents">
        <h2>Contents (min 1)</h2>
        <div
            v-for="(content, index) in postData.contents"
            :key="index"
            class="create-post__content-item"
        >
          <label class="create-post__label-inline">
            Type:
            <select
                v-model="content.type"
                class="create-post__select"
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
          <label class="create-post__label-inline">
            Data:
            <input
                type="text"
                v-model="content.data"
                class="create-post__input"
                placeholder="URL or text"
                required
            />
          </label>
          <label class="create-post__checkbox-label">
            Is Link?
            <input
                type="checkbox"
                v-model="content.is_link"
                class="create-post__checkbox"
            />
          </label>

          <AppButton
              type="button"
              class="create-post__remove-btn"
              v-if="postData.contents.length > 1"
              @click="removeContent(index)"
          >
            Remove
          </AppButton>
        </div>
        <AppButton
            type="button"
            @click="addContent"
        >
          + Add Content
        </AppButton>
      </div>

      <AppButton type="submit">
        Create Post
      </AppButton>
    </form>
  </div>
</template>

<script setup lang="ts">
import {ref, watch} from 'vue'
import {type CreatePostRequest, PostTypes} from '../sdk'
import {createPost} from '../sdk'
import {useRouter} from 'vue-router'
import AppButton from "../components/AppButton.vue";

const postData = ref<CreatePostRequest>({
  title: '',
  description: '',
  preview_url: '',
  type: '',
  tags: [],
  contents: [
    {
      data: '',
      type: '',
      is_link: true
    }
  ]
})

const tagsInput = ref('')

watch(tagsInput, (val) => {
  const rawTags = val.split(',').map(t => t.trim()).filter(Boolean)
  if (rawTags.length > 5) {
    // keep only first 5
    // TODO: Make it in another way
    rawTags.splice(5)
  }
  postData.value.tags = rawTags
})

function addContent() {
  postData.value.contents.push({
    data: '',
    type: '',
    is_link: true
  })
}

function removeContent(index: number) {
  if (postData.value.contents.length > 1) {
    postData.value.contents.splice(index, 1)
  }
}

const router = useRouter()

async function onSubmit() {
  if (!postData.value.type) {
    alert('Please select a post type.')
    return
  }
  if (postData.value.contents.length === 0) {
    alert('At least one content is required.')
    return
  }
  try {
    const created = await createPost({...postData.value})
    await router.push(`/posts/${created.id}`)
  } catch (error) {
    console.error(error)
    alert('Failed to create post.')
  }
}
</script>

<style scoped lang="scss">
.create-post {
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

  &__textarea {
    resize: vertical;
  }

  &__contents {
    background: #2b2b2b;
    padding: 10px;
    border-radius: 6px;

    & > h2 {
      color: #fff;
      margin-top: 0;
      margin-bottom: 10px;
    }
  }

  &__content-item {
    margin-bottom: 10px;
  }

  &__remove-btn {
    margin-top: 5px;
  }
}
</style>
