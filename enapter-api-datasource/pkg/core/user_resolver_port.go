package core

import "context"

type ResolveUserRequest struct {
	Email string
}

type ResolveUserResponse struct {
	ID string
}

type UserResolverPort interface {
	ResolveUser(
		context.Context, *ResolveUserRequest,
	) (*ResolveUserResponse, error)
}
