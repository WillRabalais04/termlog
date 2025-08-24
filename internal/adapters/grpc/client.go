package grpc

import (
	"google.golang.org/grpc"
)

func NewClient() (*grpc.ClientConn, error) {

	conn, err := grpc.NewClient("")

	return conn, err
}
