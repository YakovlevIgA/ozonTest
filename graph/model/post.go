package model

type Post struct {
	ID               string                `json:"id"`
	Title            string                `json:"title"`
	Content          string                `json:"content"`
	AuthorID         string                `json:"authorID"`
	CreatedAt        string                `json:"createdAt"`
	CommentsDisabled bool                  `json:"commentsDisabled"`
	Comments         []*CommentWithReplies `json:"comments,omitempty"`
}
