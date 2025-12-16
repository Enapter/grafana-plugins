package core_test

import (
	"context"

	"github.com/stretchr/testify/suite"

	"github.com/Enapter/grafana-plugins/pkg/core"
)

type MockUserResolver struct {
	suite           *suite.Suite
	resolveUserFunc func(
		ctx context.Context,
		req *core.ResolveUserRequest,
	) (*core.ResolveUserResponse, error)
}

func NewMockUserResolver(s *suite.Suite) *MockUserResolver {
	m := &MockUserResolver{
		suite: s,
	}
	m.resolveUserFunc = m.unexpectedResolveUserCall
	return m
}

func (m *MockUserResolver) ExpectResolveUserAndReturn(
	wantReq *core.ResolveUserRequest,
	resp *core.ResolveUserResponse, err error,
) {
	m.resolveUserFunc = func(
		_ context.Context, haveReq *core.ResolveUserRequest,
	) (*core.ResolveUserResponse, error) {
		defer func() {
			m.resolveUserFunc = m.unexpectedResolveUserCall
		}()
		m.suite.Require().Equal(wantReq, haveReq)
		return resp, err
	}
}

func (m *MockUserResolver) ResolveUser(
	ctx context.Context,
	req *core.ResolveUserRequest,
) (*core.ResolveUserResponse, error) {
	return m.resolveUserFunc(ctx, req)
}

func (m *MockUserResolver) unexpectedResolveUserCall(
	ctx context.Context, req *core.ResolveUserRequest,
) (*core.ResolveUserResponse, error) {
	m.suite.FailNow("unexpected call")
	//nolint: nilnil // unreachable
	return nil, nil
}
