<template>
  <div class="posts-view">
    <FilterMenu @filter="onFilter" :initialFilters="filters"/>
    <div class="post-list">
      <PostCard
          v-for="post in posts"
          :key="post.id"
          :post="post"
          :can-interact="!!auth.userId"
          @favorite="handleFavorite"
          @rateUpvote="handleRateUpvote"
          @rateDownvote="handleRateDownvote"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref, onMounted, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import FilterMenu from '../components/FilterMenu.vue'
import PostCard from '../components/PostCard.vue'
import {getPosts, favoritePost, unfavoritePost, ratePost} from '../sdk'
import {PostsQueryBuilder} from '../sdk'
import type {Post} from '../sdk'
import {auth} from '../store/auth'

const route = useRoute()
const router = useRouter()

const defaultFilters = {page: 1, limit: 10}
const filters = ref<Record<string, any>>({...defaultFilters, ...route.query})

const posts = ref<Post[]>([])

const fetchPosts = async () => {
  const builder = new PostsQueryBuilder()
      .page(filters.value.page)
      .limit(filters.value.limit)
  const query = {...builder.build(), ...filters.value}
  posts.value = await getPosts(query)
}

watch(filters, (newFilters) => {
  router.replace({query: {...newFilters}})
})

onMounted(() => {
  if (Object.keys(route.query).length) {
    filters.value = {...defaultFilters, ...route.query}
  }
  fetchPosts()
})

const onFilter = (newFilters: Record<string, any>) => {
  filters.value = {...filters.value, ...newFilters}
  fetchPosts()
}

const updatePost = (updatedPost: Post) => {
  posts.value = posts.value.map((p) => (p.id === updatedPost.id ? updatedPost : p))
}

const handleFavorite = async (post: Post) => {
  try {
    if (post.interaction_data.is_favorite) {
      await unfavoritePost(post.id)
      post.interaction_data.is_favorite = false
      post.favorites_count = Math.max(post.favorites_count - 1, 0)
    } else {
      await favoritePost(post.id)
      post.interaction_data.is_favorite = true
      post.favorites_count += 1
    }
    updatePost(post)
  } catch (error) {
    console.error(error)
  }
}

const handleRateUpvote = async (post: Post) => {
  try {
    if (post.interaction_data.vote === 'upvoted') {
      await ratePost(post.id, 'retract')
      post.rating -= 1
      post.interaction_data.vote = 'none'
    } else if (post.interaction_data.vote === 'downvoted') {
      await ratePost(post.id, 'upvote')
      post.rating += 2
      post.interaction_data.vote = 'upvoted'
    } else {
      await ratePost(post.id, 'upvote')
      post.rating += 1
      post.interaction_data.vote = 'upvoted'
    }
    updatePost(post)
  } catch (error) {
    console.error(error)
  }
}

const handleRateDownvote = async (post: Post) => {
  try {
    if (post.interaction_data.vote === 'downvoted') {
      await ratePost(post.id, 'retract')
      post.rating += 1
      post.interaction_data.vote = 'none'
    } else if (post.interaction_data.vote === 'upvoted') {
      await ratePost(post.id, 'downvote')
      post.rating -= 2
      post.interaction_data.vote = 'downvoted'
    } else {
      await ratePost(post.id, 'downvote')
      post.rating -= 1
      post.interaction_data.vote = 'downvoted'
    }
    updatePost(post)
  } catch (error) {
    console.error(error)
  }
}
</script>

<style lang="scss" scoped>
.posts-view {
  .post-list {
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin-top: 20px;
  }
}
</style>
