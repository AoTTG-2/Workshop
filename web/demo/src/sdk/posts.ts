import {callApi} from './http';
import type {Post, CreatePostRequest, UpdatePostRequest, ModeratePostRequest, RatePostRequest} from './types';
import {PostsQueryBuilder} from './queryBuilder';

export async function getPosts(filters?: Record<string, any>): Promise<Post[]> {
    const builder = new PostsQueryBuilder();

    if (filters) {
        if (filters.search_query) builder.search(filters.search_query);
        if (filters.author_id) builder.author(filters.author_id);
        if (filters.only_approved !== undefined) builder.onlyApproved(filters.only_approved);
        if (filters.show_declined !== undefined) builder.showDeclined(filters.show_declined);
        if (filters.type) builder.postType(filters.type);
        if (filters.tags) builder.tags(filters.tags);
        if (filters.for_user_id) builder.forUser(filters.for_user_id);
        if (filters.only_favorites !== undefined) builder.onlyFavorites(filters.only_favorites);
        if (filters.rating_filter) builder.ratingFilter(filters.rating_filter);
        if (filters.sort_type) builder.sortType(filters.sort_type);
        if (filters.sort_order) builder.sortOrder(filters.sort_order);
        if (filters.page) builder.page(filters.page);
        if (filters.limit) builder.limit(filters.limit);
    }

    const query = builder.build();

    Object.keys(query).forEach((key) => {
        if (query[key] === undefined || query[key] === null) {
            delete query[key];
        }
    });

    const response = await callApi('/posts', {params: query});

    if (!Array.isArray(response)) {
        throw new Error('Invalid response for getPosts');
    }

    return response;
}


export async function getPost(postId: string | number): Promise<Post> {
    const response = await callApi(`/posts/${postId}`);
    if (!response || typeof response !== 'object') {
        throw new Error('Invalid response for getPost');
    }
    return response;
}

export async function favoritePost(postId: number): Promise<void> {
    const response = await callApi(`/posts/${postId}/favorite`, {method: 'POST'});
    if (response !== null) {
        throw new Error('Invalid response for favoritePost');
    }
}

export async function unfavoritePost(postId: number): Promise<void> {
    const response = await callApi(`/posts/${postId}/favorite`, {method: 'DELETE'});
    if (response !== null) {
        throw new Error('Invalid response for unfavoritePost');
    }
}

export async function ratePost(postId: number, rating: 'upvote' | 'downvote' | 'retract'): Promise<void> {
    const payload: RatePostRequest = {rating};
    const response = await callApi(`/posts/${postId}/rate`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(payload)
    });
    if (response !== null) {
        throw new Error('Invalid response for ratePost');
    }
}

export async function moderatePost(postId: number, action: 'approve' | 'decline', note?: string): Promise<void> {
    const payload: ModeratePostRequest = {action, note};
    const response = await callApi(`/posts/${postId}/moderate`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(payload)
    });
    if (response !== null) {
        throw new Error('Invalid response for moderatePost');
    }
}

export async function createPost(request: CreatePostRequest): Promise<Post> {
    const response = await callApi('/posts', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(request)
    });
    if (!response || typeof response !== 'object') {
        throw new Error('Invalid response for createPost');
    }
    return response;
}

export async function updatePost(postId: number, request: UpdatePostRequest): Promise<Post> {
    const response = await callApi(`/posts/${postId}`, {
        method: 'PUT',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(request)
    })
    if (!response || typeof response !== 'object') {
        throw new Error('Invalid response for updatePost')
    }
    return response
}

export async function deletePost(postId: number): Promise<void> {
    const response = await callApi(`/posts/${postId}`, {
        method: 'DELETE',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({})
    });
    if (response !== null) {
        throw new Error('Invalid response for deletePost');
    }
}
