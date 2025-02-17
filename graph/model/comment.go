package model

type Comment struct {
	ID        string  `json:"id"`
	PostID    string  `json:"postID"`
	ParentID  *string `json:"parentID,omitempty"`
	AuthorID  string  `json:"authorID"`
	Content   string  `json:"content"`
	CreatedAt string  `json:"createdAt"`
}

type CommentConnection struct {
	Edges    []*CommentWithReplies `json:"edges"`
	PageInfo *PageInfo             `json:"pageInfo"`
}

type CommentWithReplies struct {
	ID        string                `json:"id"`
	PostID    string                `json:"postID"`
	ParentID  *string               `json:"parentID,omitempty"`
	AuthorID  string                `json:"authorID"`
	Content   string                `json:"content"`
	CreatedAt string                `json:"createdAt"`
	Replies   []*CommentWithReplies `json:"replies"`
}

type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor,omitempty"`
}
