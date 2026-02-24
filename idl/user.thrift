namespace go user

struct User {
    1: i64 id
    2: string username
    3: string email
    4: string avatar
    5: string created_at
    6: string updated_at
}

struct RegisterRequest {
    1: string username (vt.min_size = "2", vt.max_size = "32")
    2: string email    (vt.min_size = "5", vt.max_size = "64")
    3: string password (vt.min_size = "6", vt.max_size = "128")
}

struct RegisterResponse {
    1: i64  user_id
    2: string token
}

struct LoginRequest {
    1: string username (vt.min_size = "2")
    2: string password (vt.min_size = "6")
}

struct LoginResponse {
    1: i64    user_id
    2: string token
}

struct GetUserRequest {
    1: i64 user_id
}

struct GetUserResponse {
    1: User user
}

struct UpdateUserRequest {
    1: i64    user_id
    2: string username
    3: string email
    4: string avatar
}

struct UpdateUserResponse {
    1: bool success
}

service UserService {
    RegisterResponse Register(1: RegisterRequest req)
    LoginResponse Login(1: LoginRequest req)
    GetUserResponse GetUser(1: GetUserRequest req)
    UpdateUserResponse UpdateUser(1: UpdateUserRequest req)
}
