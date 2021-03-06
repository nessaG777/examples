package interceptor

import (
	"context"

	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor struct {
	JWTSecret string
}

// Auth interceptor for validating JWT token
func (interceptor *Interceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logrus.Infoln("method: ", info.FullMethod)
	err := interceptor.authorize(ctx, info.FullMethod, req)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

// Auth interceptor for validating JWT token for streams
func (interceptor *Interceptor) StreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	logrus.Infoln("method: ", info.FullMethod)
	err := interceptor.authorize(stream.Context(), info.FullMethod, srv)
	if err != nil {
		return err
	}
	return handler(srv, stream)
}

// Authorize validates JWT token
func (interceptor *Interceptor) authorize(ctx context.Context, method string, payload interface{}) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]

	token, err := jwt.Parse(accessToken, interceptor.keyFunc)

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// valid token, do something with claims
		return nil
	} else {
		return status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

}

// KeyFunc is a callback function to generate key based on the JWT token
func (interceptor *Interceptor) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(interceptor.JWTSecret), nil
}
