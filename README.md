```
go run ./server.go
```


```
docker run -p 8080:8080 \
  -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_USER=web \
  -e DB_PASSWORD=17051989 \
  -e DB_NAME=forozon \
  myapp
```

# GraphQL - создание поста:

```
mutation {
  createPost(title: "My New Post", content: "This is the content of the new post.", authorID: "123", commentsDisabled: false) {
    id
    title
    content
    authorID
    createdAt
    commentsDisabled
  }
}
```

# GraphQL - добавление комментария к посту (укажите postID):
```
mutation {
  addComment(postID: "HERE", authorID: "123", content: "This is a comment on the new post.") {
    id
    postID
    parentID
    authorID
    content
    createdAt
  }
}
```

# GraphQL - добавление вложенного комментария (укажите postID, parentID)
```
mutation {
  addComment(postID: "HERE", parentID: "HERE", authorID: "124", content: "This is a reply to the first comment.") {
    id
    postID
    parentID
    authorID
    content
    createdAt
  }
}
```

# GraphQL Query - список постов (без комментариев)
```
{
  posts {
    id
    title
    content
    authorID
    createdAt
    commentsDisabled  
  }
}
```

# GraphQL Query - список вложенных комментариев к посту и сам пост:
```
query  {
  post(id: "4ffedb71-b1db-4a99-80b5-a7843423d495") {
    id
    title
    content
    authorID
    createdAt
    comments {
      id
      content
      authorID
      createdAt
      replies {
        id
        content
        authorID
        createdAt
        replies {
          id
          content
          authorID
          createdAt
        }
      }
    }
  }
}
```

# GraphQL Query - список вложенных комментариев к посту:
```
query {
  comments(postID: "fe921932-df3e-4ae0-a0de-30d77c906d92", limit: 1000) {
    edges {
      id
      content
      authorID
      createdAt
      parentID
      replies {
        id
        content
        authorID
        createdAt
        parentID
        replies {
          id
          content
          authorID
          createdAt
          parentID
          replies {
            id
            content
            authorID
            createdAt
            parentID
            # Добавляйте больше уровней, если нужно
          }
        }
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```





