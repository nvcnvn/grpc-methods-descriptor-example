# grpc-methods-descriptor-example

We have a simple gRPC service:
```proto
message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string id = 1;
  string name = 2;
  string email = 3;
}

service UsersService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```
gRPC is great in many ways, its performance is great, its ecosystem cannot be compared.  
But imho, top over all of that is a typed contract its provide.  
I'm a backend engineer, I and my mobile and web friends can sit down, discuss, come up with an agreement then we generate the stub client code in flutter and es for mock implementation, regroup after 3 days.  
A good and performance day!  

But wait, we're missing something!
### What exactly user role can call this API?
```proto
service UsersService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```
`GetUser` method almost have everything, but still not enough for describing the authorization requirement.
- `Get`: this is the verb, the action
- `User`: this is the resource
And we only missing the `Role` part to describe a **RBAC** rule.  
Tooooo bad, we come so close, just if we can do something, something type-safe, something generic... :cry:
## Your wish will come true, with the help of proto descriptor
### First, extend the `google.protobuf.MethodOptions` to describe your policy
```proto
enum Role {
  ROLE_UNSPECIFIED = 0;
  ROLE_CUSTOMER = 1;
  ROLE_ADMIN = 2;
}

message RoleBasedAccessControl {
  repeated Role allowed_roles = 1;
  bool allow_unauthenticated = 2;
}

extend google.protobuf.MethodOptions {
  optional RoleBasedAccessControl access_control = 90000; // I don't know about this 90000, seem as long as it's unique
}
```
### Then, use the new `MethodOptions` in your Methods definition (mind the import path)
```proto
service UsersService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (components.rbac.v1.access_control) = {
      allowed_roles: [
        ROLE_CUSTOMER,
        ROLE_ADMIN
      ]
      allow_unauthenticated: false
    };
  }

  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (components.rbac.v1.access_control) = {
      allowed_roles: [ROLE_ADMIN]
      allow_unauthenticated: false
    };
  }
}
```
### then fork and modify the grpc-go proto generator command
Just kidding, the night is late and I need to be in the office before 08:00 AM :cry:
### just write an interceptor and use it with your server
Load the methods descriptor:
```go
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
```
Find our `access_control` option:
```go
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
```
### example:
Add or append the newly created interceptor to your backend
```go
func newUsersServer() *grpc.Server {
	svc := grpc.NewServer(grpc.UnaryInterceptor(interceptors.RBACUnaryInterceptor))
	usersv1.RegisterUsersServiceServer(svc, &usersServer{})
	return svc
}
```
Then test it (just run my Example):
```go
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
```