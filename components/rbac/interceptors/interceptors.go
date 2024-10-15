package interceptors

import (
	"context"
	"strings"

	rbacv1 "github.com/grpc-methods-descriptor-example/components/rbac/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	ErrMissingRole     = status.Errorf(codes.Unauthenticated, "missing role")
)

func RBACUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// the info give us: /components.users.v1.UsersService/GetUser
	// we need to convert it to components.users.v1.UsersService.GetUser
	methodName := strings.Replace(strings.TrimPrefix(info.FullMethod, "/"), "/", ".", -1)
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(methodName))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "method not found descriptor")
	}

	method, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, status.Errorf(codes.Internal, "some hoe this is not a method")
	}

	var policy *rbacv1.RoleBasedAccessControl
	method.Options().ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.FullName() != rbacv1.E_AccessControl.TypeDescriptor().FullName() {
			// continue finding the AccessControl field
			return true
		}

		b, err := proto.Marshal(v.Message().Interface())
		if err != nil {
			// TODO: better handle this as an Internal error
			// but for now, just return PermissionDenied
			return false
		}

		policy = &rbacv1.RoleBasedAccessControl{}
		if err := proto.Unmarshal(b, policy); err != nil {
			// same as above, better handle this as an Internal error
			return false
		}

		// btw I think this policy can be cached

		return false
	})

	if policy == nil {
		// secure by default, DENY_ALL if no control policy is found
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	if policy.AllowUnauthenticated {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMissingMetadata
	}

	if len(md["role"]) == 0 {
		return nil, ErrMissingRole
	}

	currentRoles := make(map[string]struct{})
	for _, role := range md["role"] {
		currentRoles[role] = struct{}{}
	}

	foundAllowed := false
	for _, role := range policy.AllowedRoles {
		if _, ok := currentRoles[role.String()]; ok {
			foundAllowed = true
		}
	}

	if !foundAllowed {
		return nil, status.Errorf(codes.PermissionDenied, "permission denied")
	}

	return handler(ctx, req)
}
