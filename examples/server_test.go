package examples

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	usersv1 "github.com/grpc-methods-descriptor-example/components/users/v1"
)

func newUsersServer() *grpc.Server {
	svc := grpc.NewServer()
	usersv1.RegisterUsersServiceServer(svc, &usersServer{})
	return svc
}

type usersServer struct {
	usersv1.UnimplementedUsersServiceServer
}

func (s *usersServer) GetUser(ctx context.Context, in *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	return &usersv1.GetUserResponse{}, nil
}

func Example() {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	s := newUsersServer()
	go s.Serve(l)

	conn, err := grpc.NewClient(l.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	client := usersv1.NewUsersServiceClient(conn)
	_, err = client.GetUser(context.Background(), &usersv1.GetUserRequest{})
	fmt.Println(status.Code(err))
	// Output: OK
}
