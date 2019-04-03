package grpc_server

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log"
)

var (
	Err_AuthFail   = errors.New("auth failed")
	Err_NoAuthInfo = errors.New("No auth info")
)

type RPCSecurity struct {
	auths map[string]string
	Creds credentials.TransportCredentials
}

func NewRPCSecurity(auths map[string]string, certFile, keyFile string) *RPCSecurity {
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatal("RPCSvrRun RPCSecurity Failed to generate credentials NewServerTLSFromFile err:%v", err)
	}
	return &RPCSecurity{
		auths: auths,
		Creds: creds,
	}
}
func (s *RPCSecurity) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := s.authorize(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

func (s *RPCSecurity) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := s.authorize(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (s *RPCSecurity) authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md[AuthName]) > 0 &&
			len(md[AuthPwd]) > 0 {
			if pwd, ok := s.auths[md[AuthName][0]]; ok {
				if pwd == md[AuthPwd][0] {
					return nil
				}
			}
		}

		return Err_AuthFail
	}

	return Err_NoAuthInfo
}
