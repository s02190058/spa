# Single Page Application

## Overview

The purposes of the project are:

- to see how the Go app structure can be organized
- to understand how restful API works
- to upgrade my knowledge of the Go standard library (packages `net/http`, `database/sql`, ...)
- to work with `postgresql` database, `migrate` tool, `docker` containers

All files from `static` folder were already crated. I wrote the rest myself.

## Quick start

Before launching the application, you must specify all sensitive data
in the `.env` file (see `example.env`).

then enter the command below:

```shell
make compose-up
```

## API Endpoints

1) `POST /api/register` - user registration
2) `POST /api/login` - user login
3) `GET /api/posts/` - list of all posts
4) `POST /api/posts` - adding a post (`url/text`)
5) `GET /api/funny/{category_name}` - list of posts with the certain category
6) `GET /api/post/{post_id}` - certain post
7) `POST /api/post/{post_id}` - adding a comment
8) `DELETE /api/post/{post_id}/{comment_id}` - deleting a post
9) `GET /api/post/{post_id}/upvote` - upvote post rating
10) `GET /api/post/{post_id}/downvote` - downvote post rating
11) `GET /api/post/{post_id}/unvote` - unvote post rating
12) `DELETE /api/post/{post_id}` - deleting a post
13) `GET /api/user/{username}` - list of all posts of the certain user

## TODO

- write tests
- add gRPC, CLI transport
- create customError type like
`type CustomError struct {Err error, HTTPCode int}`
which will allow to remove all switch operators in a transport layer
