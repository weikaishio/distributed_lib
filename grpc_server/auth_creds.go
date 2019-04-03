package grpc_server

import (
	"context"
)

const (
	AuthName = "username"
	AuthPwd  = "password"
)

type AuthCreds struct {
	UserName string
	Password string
	IsTLS    bool
}

func (a *AuthCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		AuthName: a.UserName,
		AuthPwd:  a.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires
// transport security.
func (a *AuthCreds) RequireTransportSecurity() bool {
	return a.IsTLS
}
