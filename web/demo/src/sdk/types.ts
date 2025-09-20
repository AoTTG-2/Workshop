export type RateType = 'upvoted' | 'downvoted' | 'voted' | 'none';
export type ModerationStatus = 'approved' | 'declined' | 'pending';

export const PostTypes: Record<string, string> = {
    map_suite: 'Map Suite',
    game_mode: 'Game Mode',
    skin_set: 'Skin Set',
    custom_assets: 'Custom Assets'
}

export const ContentTypes: Record<string, string> = {
    custom_map: 'Custom Map',
    custom_logic: 'Custom Logic',
    custom_asset: 'Custom Assets',
    custom_skin: 'Skin Set'
}

export interface PostContent {
    id?: number;
    content_type: string;
    content_data: string;
    is_link: boolean;
}

export interface PostInteractionData {
    is_favorite: boolean;
    vote: RateType;
}

export interface PostModerationData {
    status: ModerationStatus;
    note: string;
}

export interface Post {
    id: number;
    author_id: string;
    title: string;
    description: string;
    preview_url: string;
    post_type: string;
    tags: string[];
    contents: PostContent[];
    created_at: string;
    updated_at: string;
    moderation_data: PostModerationData;
    interaction_data: PostInteractionData;
    rating: number;
    comments_count: number;
    favorites_count: number;
}

export interface Comment {
    id: number;
    post_id: number;
    author_id: string;
    content: string;
    created_at: string;
    updated_at: string;
}

export interface CreatePostRequest {
    title: string;
    description: string;
    preview_url: string;
    type: string;
    tags: string[];
    contents: { data: string; type: string; is_link: boolean }[];
}

export interface UpdatePostContentRequest {
    id?: number
    content_type: string
    content_data: string
    is_link: boolean
}

export interface UpdatePostRequest {
    title: string
    description: string
    preview_url: string
    type: string
    tags: string[]
    contents: UpdatePostContentRequest[]
}

export interface AddCommentRequest {
    content: string;
    postID: number;
}

export interface RatePostRequest {
    rating: 'upvote' | 'downvote' | 'retract';
}

export interface ModeratePostRequest {
    action: 'approve' | 'decline';
    note?: string;
}
