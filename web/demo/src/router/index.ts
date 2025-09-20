import {createRouter, createWebHistory} from 'vue-router'
import PostListView from '../views/PostListView.vue'
import PostView from '../views/PostView.vue'
import NotFound from '../views/NotFound.vue'
import CreatePostView from "../views/CreatePostView.vue";
import EditPostVue from "../views/EditPostVue.vue";

const routes = [
    {path: '/', redirect: '/posts'},
    {path: '/posts', name: 'Posts', component: PostListView},
    {path: '/posts/:postId', name: 'PostDetail', component: PostView, props: true},
    {path: '/posts/:postId/edit', name: 'EditPost', component: EditPostVue, props: true},
    {path: '/create', name: 'CreatePost', component: CreatePostView},
    {path: '/:pathMatch(.*)*', name: 'NotFound', component: NotFound},
]

export default createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes
})
