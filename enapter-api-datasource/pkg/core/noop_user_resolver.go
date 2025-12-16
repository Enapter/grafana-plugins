package core

import (
	"context"
)

type NoopUserResolver struct{}

func (NoopUserResolver) ResolveUser(
	_ context.Context, req *ResolveUserRequest,
) (*ResolveUserResponse, error) {
	return &ResolveUserResponse{
		ID: req.Email,
	}, nil
}
