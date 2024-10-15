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
