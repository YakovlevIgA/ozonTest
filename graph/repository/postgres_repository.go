package repository

import (
	"context"
	"fmt"
	"github.com/YakovlevIgA/forozon/graph"
	"github.com/YakovlevIgA/forozon/graph/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"time"
)

// PostgresRepository репозиторий на основе postgres
type PostgresRepository struct {
	conn *pgx.Conn
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(conn *pgx.Conn) (graph.Storage, error) {
	return &PostgresRepository{conn: conn}, nil
}

// CreatePost создание поста
func (s *PostgresRepository) CreatePost(ctx context.Context, title, content, authorID string, commentsDisabled bool) (*model.Post, error) {
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

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	id := generateID()
	createdAt := time.Now().UTC()
	_, err = tx.Exec(
		ctx,
		"INSERT INTO posts (id, title, content, authorID, commentsDisabled, createdAt) VALUES ($1, $2, $3, $4, $5, $6)",
		id, title, content, authorID, commentsDisabled, createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert post: %v", err)
	}

	post := &model.Post{
		ID:               id,
		Title:            title,
		Content:          content,
		AuthorID:         authorID,
		CreatedAt:        createdAt.String(),
		CommentsDisabled: commentsDisabled,
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %v", err)
	}

	return post, nil
}

// AddComment добавление комментария
func (s *PostgresRepository) AddComment(ctx context.Context, postID string, parentID *string, authorID string, content string) (*model.Comment, error) {
	// Валидация

	if authorID == "" {
		return nil, fmt.Errorf("authorID is required")
	}

	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	post, err := s.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post for comment: %v", err)
	}

	if post == nil {
		return nil, fmt.Errorf("post for comment not found")
	}

	if post.CommentsDisabled {
		return nil, fmt.Errorf("comments are disabled for this post")
	}

	// Исполнение

	id := generateID()

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	createdAt := time.Now().UTC().String()
	_, err = tx.Exec(
		ctx,
		"INSERT INTO comments (id, postID, parentID, authorID, content, createdAt) VALUES ($1, $2, $3, $4, $5, $6)",
		id, postID, parentID, authorID, content, createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert comment: %v", err)
	}

	comment := &model.Comment{
		ID:        id,
		PostID:    postID,
		ParentID:  parentID,
		AuthorID:  authorID,
		Content:   content,
		CreatedAt: createdAt,
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %v", err)
	}

	return comment, nil
}

func (s *PostgresRepository) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := s.conn.QueryRow(ctx, "SELECT id, title, content, authorID, createdAt, commentsDisabled FROM posts WHERE id=$1", id).Scan(
		&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.CommentsDisabled,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve post: %v", err)
	}

	comments, err := s.GetCommentsByPostID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comments: %v", err)
	}
	post.Comments = comments

	return &post, nil
}

func (s *PostgresRepository) GetCommentsByPostID(ctx context.Context, postID string) ([]*model.CommentWithReplies, error) {
	var comments []*model.CommentWithReplies
	rows, err := s.conn.Query(ctx, "SELECT id, postID, parentID, authorID, content, createdAt FROM comments WHERE postID=$1", postID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve comments: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comment model.CommentWithReplies
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.ParentID, &comment.AuthorID, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}
		comments = append(comments, &comment)
	}

	return buildCommentHierarchy(comments), nil
}

func buildCommentHierarchy(comments []*model.CommentWithReplies) []*model.CommentWithReplies {
	m := make(map[string]*model.CommentWithReplies)
	var roots []*model.CommentWithReplies

	for _, c := range comments {
		m[c.ID] = c
	}

	for _, c := range comments {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			if parent, ok := m[*c.ParentID]; ok {
				parent.Replies = append(parent.Replies, c)
			}
		}
	}

	return roots
}

// GetPosts Получение всех постов с комментариями
func (s *PostgresRepository) GetPosts(ctx context.Context) ([]*model.Post, error) {
	// Шаг 1: Получаем все посты
	rows, err := s.conn.Query(ctx, "SELECT id, title, content, authorID, createdAt, commentsDisabled FROM posts")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %v", err)
	}
	defer rows.Close()

	var posts []*model.Post
	postIDs := []string{} // Сохраняем ID постов для дальнейшего запроса комментариев
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.CommentsDisabled); err != nil {
			return nil, err
		}
		postIDs = append(postIDs, post.ID)
		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Шаг 2: Получаем все комментарии для этих постов
	query := `SELECT id, postID, parentID, authorID, content, createdAt FROM comments WHERE postID = ANY($1)`
	rows, err = s.conn.Query(ctx, query, postIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var comments []*model.CommentWithReplies
	commentMap := make(map[string]*model.CommentWithReplies)
	for rows.Next() {
		var c model.CommentWithReplies
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentID, &c.AuthorID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
		commentMap[c.ID] = &c
	}

	// Шаг 3: Строим иерархию комментариев
	commentTree := buildCommentTree(comments)

	// Привязываем комментарии к постам в ходе формирования ответа
	for _, post := range posts {
		var postComments []*model.CommentWithReplies
		for _, c := range commentTree {
			if c.PostID == post.ID {
				postComments = append(postComments, c)
			}
		}
		// Вставляем комментарии в ответ
		post.Comments = postComments
	}

	return posts, nil
}

func (s *PostgresRepository) GetCommentsForPost(ctx context.Context, postID string, limit int, cursor *string) ([]*model.CommentWithReplies, error) {
	var query string
	var args []interface{}

	if cursor != nil && *cursor != "" {
		query = `SELECT id, postID, parentID, authorID, content, createdAt FROM comments WHERE postID=$1 AND id > $2 ORDER BY id LIMIT $3`
		args = append(args, postID, *cursor, limit)
	} else {
		query = `SELECT id, postID, parentID, authorID, content, createdAt FROM comments WHERE postID=$1 ORDER BY id LIMIT $2`
		args = append(args, postID, limit)
	}

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*model.CommentWithReplies
	for rows.Next() {
		var c model.CommentWithReplies
		if err := rows.Scan(&c.ID, &c.PostID, &c.ParentID, &c.AuthorID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}

	return buildCommentTree(comments), nil
}

func generateID() string {
	return uuid.New().String() // Используем UUID для генерации уникальных ID
}
