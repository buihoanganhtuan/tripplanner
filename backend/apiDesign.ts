abstract class TripPlannerApi {
    static version = "v1";
    static title = "Trip Planner API"

    // resource: User
    @post("/users")
    CreateUser(req: CreateUserRequest) : User;

    @get("/{name=users/*}")
    GetUser(req: GetUserRequest) : User;

    @get("/users")
    ListUsers(req: ListUsersRequest) : User[];

    @patch("/{name=users/*}")
    UpdateUser(req: UpdateUserRequest) : User;

    @delete("/{name=/users/*}")
    DeleteUser(req: DeleteUserRequest) : void;

}

interface CreateUserRequest {
    user: User;
}

// Just to be user friendly. Technically, GetUserRequest should have contained id instead
interface GetUserRequest {
    name: string;
}

interface ListUsersRequest {
    filter: string;
}

// fieldMask shall be sent as query string
interface UpdateUserRequest {
    user: User;
    fieldMask: string[];
}

interface DeleteUserRequest {
    name: string;
}

interface User {
    id: string;

    name: string;
    password: string;
    email: string;
    joinDate: string;
}