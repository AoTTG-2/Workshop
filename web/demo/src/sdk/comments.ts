import { callApi } from './http';
import type { Comment, AddCommentRequest } from './types';
import { CommentsQueryBuilder } from './queryBuilder';

export async function getComments(postId: number, page: number, limit: number, extraParams?: Record<string, any>): Promise<Comment[]> {
    const builder = new CommentsQueryBuilder().postId(postId).page(page).limit(limit);
    if (extraParams) {
        Object.assign(builder, extraParams);
    }
    const query = builder.build();
    const response = await callApi('/comments', { params: query });
    if (!Array.isArray(response)) {
        throw new Error('Invalid response for getComments');
    }
    return response;
}

export async function addComment(content: string, postId: number): Promise<Comment> {
    const payload: AddCommentRequest = { content, postID: postId };
    const response = await callApi('/comments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
    });
    if (!response || typeof response !== 'object') {
        throw new Error('Invalid response for addComment');
    }
    return response;
}

export async function updateComment(commentID: number, content: string): Promise<Comment> {
    const response = await callApi(`/comments/${commentID}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content })
    });
    if (!response || typeof response !== 'object') {
        throw new Error('Invalid response for updateComment');
    }
    return response;
}

export async function deleteComment(commentID: number): Promise<void> {
    const response = await callApi(`/comments/${commentID}`, { method: 'DELETE' });
    if (response !== null) {
        throw new Error('Invalid response for deleteComment');
    }
}
