package examples

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/grpc-methods-descriptor-example/components/rbac/interceptors"
	usersv1 "github.com/grpc-methods-descriptor-example/components/users/v1"
)

func newUsersServer() *grpc.Server {
	svc := grpc.NewServer(grpc.UnaryInterceptor(interceptors.RBACUnaryInterceptor))
	usersv1.RegisterUsersServiceServer(svc, &usersServer{})
	return svc
}

// usersServer is just a dummy implementation to avoid the unwanted error
type usersServer struct {
	usersv1.UnimplementedUsersServiceServer
}

func (s *usersServer) GetUser(ctx context.Context, in *usersv1.GetUserRequest) (*usersv1.GetUserResponse, error) {
	return &usersv1.GetUserResponse{}, nil
}

func (s *usersServer) DeleteUser(context.Context, *usersv1.DeleteUserRequest) (*usersv1.DeleteUserResponse, error) {
	return &usersv1.DeleteUserResponse{}, nil
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

	// assuming the role is ROLE_CUSTOMER
	// in the real world, this should be done by the authentication, JWT, session, etc.
	peasantCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "ROLE_CUSTOMER"))
	_, err = client.GetUser(peasantCtx, &usersv1.GetUserRequest{})
	fmt.Println(status.Code(err))

	_, err = client.DeleteUser(peasantCtx, &usersv1.DeleteUserRequest{})
	fmt.Println(status.Code(err))

	knightlyAdminCtx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("role", "ROLE_ADMIN"))
	_, err = client.GetUser(knightlyAdminCtx, &usersv1.GetUserRequest{})
	fmt.Println(status.Code(err))

	_, err = client.DeleteUser(knightlyAdminCtx, &usersv1.DeleteUserRequest{})
	fmt.Println(status.Code(err))
	// Output:
	// OK
	// PermissionDenied
	// OK
	// OK
}
