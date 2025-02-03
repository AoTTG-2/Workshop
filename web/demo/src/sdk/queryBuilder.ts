export class PostsQueryBuilder {
    private params: Record<string, any> = {};
    search(query: string) {
        this.params.search_query = query;
        return this;
    }
    author(authorId: string) {
        this.params.author_id = authorId;
        return this;
    }
    onlyApproved(flag: boolean = true) {
        this.params.only_approved = flag;
        return this;
    }
    showDeclined(flag: boolean = true) {
        this.params.show_declined = flag;
        return this;
    }
    postType(type: string) {
        this.params.type = type;
        return this;
    }
    tags(tags: string[]) {
        this.params.tags = tags;
        return this;
    }
    forUser(userId: string) {
        this.params.for_user_id = userId;
        return this;
    }
    onlyFavorites(flag: boolean = true) {
        this.params.only_favorites = flag;
        return this;
    }
    ratingFilter(filter: 'upvoted' | 'downvoted' | 'voted' | 'none') {
        this.params.rating_filter = filter;
        return this;
    }
    sortType(sort: 'popularity' | 'best_rated' | 'newest' | 'recently_updated' | 'most_discussed') {
        this.params.sort_type = sort;
        return this;
    }
    sortOrder(order: 'asc' | 'desc') {
        this.params.sort_order = order;
        return this;
    }
    page(page: number) {
        this.params.page = page;
        return this;
    }
    limit(limit: number) {
        this.params.limit = limit;
        return this;
    }
    build() {
        return this.params;
    }
}

export class CommentsQueryBuilder {
    private params: Record<string, any> = {};
    postId(postId: number) {
        this.params.postID = postId;
        return this;
    }
    author(authorId: string) {
        this.params.authorID = authorId;
        return this;
    }
    sortOrder(order: 'asc' | 'desc') {
        this.params.sort_order = order;
        return this;
    }
    page(page: number) {
        this.params.page = page;
        return this;
    }
    limit(limit: number) {
        this.params.limit = limit;
        return this;
    }
    build() {
        return this.params;
    }
}
