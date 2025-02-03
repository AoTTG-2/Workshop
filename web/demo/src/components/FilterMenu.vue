<template>
  <form class="filter-menu" @submit.prevent="applyFilters">
    <label class="filter-menu__label">
      Search Query
      <input
          type="text"
          v-model="filters.search_query"
          class="filter-menu__input"
      />
    </label>
    <label class="filter-menu__label">
      Author ID
      <input
          type="text"
          v-model="filters.author_id"
          class="filter-menu__input"
      />
    </label>
    <label class="filter-menu__label filter-menu__checkbox-label">
      Only Approved
      <input
          type="checkbox"
          v-model="filters.only_approved"
          class="filter-menu__checkbox"
      />
    </label>
    <label class="filter-menu__label filter-menu__checkbox-label">
      Show Declined
      <input
          type="checkbox"
          v-model="filters.show_declined"
          class="filter-menu__checkbox"
      />
    </label>
    <label class="filter-menu__label">
      Post Type
      <select v-model="filters.type" class="filter-menu__select">
        <option value="">None</option>
        <option
            v-for="(text, key) in PostTypes"
            :key="key"
            :value="key"
        >
          {{ text }}
        </option>
      </select>
    </label>
    <label class="filter-menu__label">
      Tags (comma separated)
      <input
          type="text"
          v-model="tagsInput"
          class="filter-menu__input"
      />
    </label>
    <label class="filter-menu__label">
      For User ID
      <input
          type="text"
          v-model="filters.for_user_id"
          class="filter-menu__input"
      />
    </label>
    <label class="filter-menu__label filter-menu__checkbox-label">
      Only Favorites
      <input
          type="checkbox"
          v-model="filters.only_favorites"
          class="filter-menu__checkbox"
      />
    </label>
    <label class="filter-menu__label">
      Rating Filter
      <select v-model="filters.rating_filter" class="filter-menu__select">
        <option value="">None</option>
        <option value="upvoted">Upvoted</option>
        <option value="downvoted">Downvoted</option>
        <option value="voted">Voted</option>
      </select>
    </label>
    <label class="filter-menu__label">
      Sort Type
      <select v-model="filters.sort_type" class="filter-menu__select">
        <option value="">None</option>
        <option value="popularity">Popularity</option>
        <option value="best_rated">Best Rated</option>
        <option value="newest">Newest</option>
        <option value="recently_updated">Recently Updated</option>
        <option value="most_discussed">Most Discussed</option>
      </select>
    </label>
    <label class="filter-menu__label">
      Sort Order
      <select v-model="filters.sort_order" class="filter-menu__select">
        <option value="desc">desc</option>
        <option value="asc">asc</option>
      </select>
    </label>
    <label class="filter-menu__label">
      Page
      <input
          type="number"
          v-model.number="filters.page"
          min="1"
          class="filter-menu__input"
          required
      />
    </label>
    <label class="filter-menu__label">
      Limit
      <input
          type="number"
          v-model.number="filters.limit"
          min="1"
          max="100"
          class="filter-menu__input"
          required
      />
    </label>

    <div class="filter-menu__actions">
      <AppButton type="submit">Apply Filters</AppButton>
    </div>
  </form>
</template>

<script setup lang="ts">
import {reactive, ref, watch, defineProps, defineEmits} from 'vue'
import AppButton from './AppButton.vue'
import {PostTypes} from "../sdk";

const props = defineProps<{ initialFilters?: Record<string, any> }>()

const emit = defineEmits<{
  (e: 'filter', filters: Record<string, any>): void
}>()

const defaultFilters = {
  search_query: '',
  author_id: '',
  only_approved: false,
  show_declined: false,
  type: '',
  tags: [] as string[],
  for_user_id: '',
  only_favorites: false,
  rating_filter: '',
  sort_type: '',
  sort_order: 'desc',
  page: 1,
  limit: 10
}

const filters = reactive({
  ...defaultFilters,
  ...(props.initialFilters || {})
})

const tagsInput = ref('')

watch(tagsInput, (newVal) => {
  filters.tags = newVal
      .split(',')
      .map((t) => t.trim())
      .filter((t) => t)
})

const applyFilters = () => {
  emit('filter', {...filters})
}
</script>

<style lang="scss" scoped>
$border-color: #444;

.filter-menu {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 15px;
  background-color: #2b2b2b;
  border: 1px solid #444;
  border-radius: 6px;

  &__label {
    display: flex;
    flex-direction: column;
    color: #fff;
    font-size: 0.9rem;
    min-width: 140px;
    margin-bottom: 0;
  }

  &__checkbox-label {
    flex-direction: row;
    align-items: center;

    input[type='checkbox'] {
      margin-left: 5px;
      margin-top: 0;
    }
  }

  &__input,
  &__select {
    padding: 5px;
    border: 1px solid $border-color;
    border-radius: 4px;
    background-color: #3b3b3b;
    color: #fff;
    font-size: 0.9rem;
  }

  &__checkbox {
    margin-top: 5px;
  }

  &__actions {
    width: 100%;
    display: flex;
    justify-content: flex-end;
    margin-top: 10px;
  }
}
</style>
