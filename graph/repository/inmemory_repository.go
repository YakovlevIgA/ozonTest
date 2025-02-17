package repository

import (
	"context"
	"fmt"
	"github.com/YakovlevIgA/forozon/graph/model"
	"log"
	"sort"
	"time"
)

// InMemoryRepository репозиторий на основе in memory
type InMemoryRepository struct {
	posts    map[string]*model.Post
	comments map[string]*model.Comment
}

// NewInMemoryRepository создает новый экземпляр InMemoryRepository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		posts:    make(map[string]*model.Post),
		comments: make(map[string]*model.Comment),
	}
}

// CreatePost создание поста
func (s *InMemoryRepository) CreatePost(_ context.Context, title, content, authorID string, commentsDisabled bool) (*model.Post, error) {
	// Валидация

	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if authorID == "" {
		return nil, fmt.Errorf("authorID is required")
	}

	// Исполнение

	id := generateID()
	createdAt := time.Now().UTC()

	post := &model.Post{
		ID:               id,
		Title:            title,
		Content:          content,
		AuthorID:         authorID,
		CreatedAt:        createdAt.String(),
		CommentsDisabled: commentsDisabled,
	}

	s.posts[id] = post
	log.Printf("Post created: %+v", post)

	return post, nil
}

// AddComment добавление комментария
func (s *InMemoryRepository) AddComment(ctx context.Context, postID string, parentID *string, authorID string, content string) (*model.Comment, error) {
	// Валидация

	if authorID == "" {
		return nil, fmt.Errorf("authorID is required")
	}

	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	post, exists := s.posts[postID]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	if post.CommentsDisabled {
		return nil, fmt.Errorf("comments are disabled for this post")
	}

	if parentID != nil && s.comments[*parentID] == nil {
		return nil, fmt.Errorf("comments parenID is not found")
	}

	// Исполнение

	id := generateID()
	createdAt := time.Now().UTC()

	comment := &model.Comment{
		ID:        id,
		PostID:    postID,
		ParentID:  parentID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: createdAt.String(),
	}

	s.comments[id] = comment
	log.Printf("Comment added: %+v", comment)

	return comment, nil
}

// GetCommentsForPost получает комментарии с пагинацией для поста
func (s *InMemoryRepository) GetCommentsForPost(ctx context.Context, postID string, limit int, cursor *string) ([]*model.CommentWithReplies, error) {
	var comments []*model.CommentWithReplies
	for _, c := range s.comments {
		if c.PostID == postID {
			comments = append(comments, &model.CommentWithReplies{
				ID:        c.ID,
				PostID:    c.PostID,
				ParentID:  c.ParentID,
				AuthorID:  c.AuthorID,
				Content:   c.Content,
				CreatedAt: c.CreatedAt,
			})
		}
	}

	sort.Slice(comments, func(i, j int) bool { return comments[i].CreatedAt < comments[j].CreatedAt })

	comments = paginateComments(comments, limit, cursor)
	log.Printf("Comments fetched for post %s: %d comments", postID, len(comments))
	return buildCommentTree(comments), nil
}

// paginateComments выполняет пагинацию комментариев
func paginateComments(comments []*model.CommentWithReplies, limit int, cursor *string) []*model.CommentWithReplies {
	if cursor != nil && *cursor != "" {
		// Если курсор существует, начинаем с комментария, следующего за ним
		var filteredComments []*model.CommentWithReplies
		found := false
		for _, c := range comments {
			if found {
				filteredComments = append(filteredComments, c)
			}
			if c.ID == *cursor {
				found = true
			}
		}
		comments = filteredComments
	}

	// Ограничиваем результат лимитом
	if len(comments) > limit {
		comments = comments[:limit]
	}

	return comments
}

// GetPosts получает все посты из памяти
func (s *InMemoryRepository) GetPosts(ctx context.Context) ([]*model.Post, error) {
	var posts []*model.Post
	for _, p := range s.posts {
		posts = append(posts, p)
	}
	return posts, nil
}

// GetPostByID получает пост по ID из памяти
func (s *InMemoryRepository) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	post, exists := s.posts[id]
	if !exists {
		return nil, fmt.Errorf("post not found")
	}

	// Загружаем комментарии и их ответы для поста
	comments := make([]*model.CommentWithReplies, 0)
	for _, comment := range s.comments {
		if comment.PostID == id {
			// Добавляем комментарий с ответами
			commentWithReplies := &model.CommentWithReplies{
				ID:        comment.ID,
				PostID:    comment.PostID,
				ParentID:  comment.ParentID,
				AuthorID:  comment.AuthorID,
				Content:   comment.Content,
				CreatedAt: comment.CreatedAt,
				Replies:   s.getRepliesForComment(comment.ID), // Загружаем ответы для комментария
			}
			comments = append(comments, commentWithReplies)
		}
	}

	// Добавляем комментарии в пост
	post.Comments = comments

	return post, nil
}

// Функция для получения ответов на комментарий
func (s *InMemoryRepository) getRepliesForComment(commentID string) []*model.CommentWithReplies {
	replies := make([]*model.CommentWithReplies, 0)
	for _, comment := range s.comments {
		// Разыменовываем ParentID, если он не nil, и сравниваем
		if comment.ParentID != nil && *comment.ParentID == commentID {
			replies = append(replies, &model.CommentWithReplies{
				ID:        comment.ID,
				PostID:    comment.PostID,
				ParentID:  comment.ParentID,
				AuthorID:  comment.AuthorID,
				Content:   comment.Content,
				CreatedAt: comment.CreatedAt,
				Replies:   []*model.CommentWithReplies{}, // Ответы для этого комментария
			})
		}
	}
	return replies
}

// buildCommentTree строит иерархию комментариев
func buildCommentTree(comments []*model.CommentWithReplies) []*model.CommentWithReplies {
	commentMap := make(map[string]*model.CommentWithReplies)
	for _, c := range comments {
		commentMap[c.ID] = c
	}
	var roots []*model.CommentWithReplies
	for _, c := range comments {
		if c.ParentID == nil || *c.ParentID == "" { // Корневой комментарий
			roots = append(roots, c)
		} else {
			parent := commentMap[*c.ParentID]
			if parent != nil {
				parent.Replies = append(parent.Replies, c)
			}
		}
	}
	return roots
}
