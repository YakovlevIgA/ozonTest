package graph

import (
	"context"
	"fmt"
	"log"

	"github.com/YakovlevIgA/forozon/graph/model"
)

// Storage интерфейс хранилища
type Storage interface {
	GetPosts(ctx context.Context) ([]*model.Post, error)
	GetPostByID(ctx context.Context, id string) (*model.Post, error)
	CreatePost(ctx context.Context, title, content, authorID string, commentsDisabled bool) (*model.Post, error)
	AddComment(ctx context.Context, postID string, parentID *string, authorID string, content string) (*model.Comment, error)
	GetCommentsForPost(ctx context.Context, postID string, limit int, cursor *string) ([]*model.CommentWithReplies, error)
}

// Resolver сервис для работы с постами и комментариями
type Resolver struct {
	Storage Storage
}

// NewResolver создает новый экземпляр Resolver
func NewResolver(storage Storage) *Resolver {
	return &Resolver{storage}
}

// CreatePost создание поста
func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, authorID string, commentsDisabled bool) (*model.Post, error) {
	post, err := r.Storage.CreatePost(ctx, title, content, authorID, commentsDisabled)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %v", err)
	}

	return post, nil
}

// AddComment создание комментария для поста
func (r *mutationResolver) AddComment(ctx context.Context, postID string, parentID *string, authorID string, content string) (*model.Comment, error) {
	comment, err := r.Storage.AddComment(ctx, postID, parentID, authorID, content)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}

	return comment, nil
}

// Post получение поста по id
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	post, err := r.Storage.GetPostByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %v", err)
	}

	if post == nil {
		return nil, fmt.Errorf("post not found")
	}

	return post, nil
}

// Posts получение списка всех постов
func (r *queryResolver) Posts(ctx context.Context) ([]*model.Post, error) {
	log.Println("Fetching posts")

	posts, err := r.Storage.GetPosts(ctx)
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		return nil, fmt.Errorf("failed to get posts: %v", err)
	}

	log.Println("Successfully fetched posts")
	return posts, nil
}

// Comments получение комментариев с пагинацией
func (r *queryResolver) Comments(ctx context.Context, postID string, limit *int32, cursor *string) (*model.CommentConnection, error) {
	var limitInt int
	if limit != nil {
		limitInt = int(*limit)
	}

	if limitInt == 0 {
		limitInt = 1000
	}

	comments, err := r.Storage.GetCommentsForPost(ctx, postID, limitInt, cursor)
	if err != nil {
		return nil, err
	}

	return &model.CommentConnection{
		Edges:    comments,
		PageInfo: &model.PageInfo{},
	}, nil
}
