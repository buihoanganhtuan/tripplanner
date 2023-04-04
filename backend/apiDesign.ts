abstract class TripPlannerApi {
    static version = "v1";
    static title = "Trip Planner API"

    // resource type: User
    @post("/users")
    CreateUser(req: CreateUserRequest) : User;

    @get("/{id=users/*}")
    GetUser(req: GetUserRequest) : User;

    @get("/users")
    ListUsers(req: ListUsersRequest) : User[];

    @patch("/{resource.name=users/*}")
    UpdateUser(req: UpdateUserRequest) : User;

    @put("/{resource.name=users/*}")
    ReplaceUser(req: ReplaceUserRequest) : User;

    @delete("/{id=users/*}")
    DeleteUser(req: DeleteUserRequest) : void;

    // resource type: User.Trip
    @post("/{resource.userId=users/*}/trips")
    CreateTrip(req: CreateTripRequest) : Trip;

    @get("/{id=users/*/trips/*}")
    GetTrip(req: GetTripRequest) : Trip;

    @patch("/{resource.id=users/*/trips/*}")
    UpdateTrip(req: UpdateTripRequest) : Trip;

    @put("/{resource.id=users/*/trips/*}")
    ReplaceTrip(req: ReplaceTripRequest) : Trip;

    @get("{parent=users/*}/trips")
    ListTrips(req: ListTripRequest) : Trip[];

    @delete("/{id=users/*/trips/*}")
    DeleteTrip(req: DeleteTripRequest) : void;

}

interface CreateUserRequest {
    resource: User;
}

// fieldMask shall be sent as query string to avoid violating constraint on standard update method
interface UpdateUserRequest {
    resource: User;
    fieldMask: string[];
}

interface ReplaceUserRequest {
    resource: User;
}

interface GetUserRequest {
    id: string;
}

interface ListUsersRequest {
    filter: string;
}

interface DeleteUserRequest {
    name: string;
}

interface CreateTripRequest {
    resource: Trip;
}

interface UpdateTripRequest {
    resource: Trip;
    fieldMask: string[];
}

interface ReplaceTripRequest {
    resource: Trip;
}

interface GetTripRequest {
    id: string;
}

interface ListTripRequest {
    parent: string;
    filter: string;
}

interface DeleteTripRequest {
    id: string;
}

interface User {
    id: string;

    name: string;
    password: string;
    email: string;
    joinDate: Datetime;
}

// Polymorphic resource, can be either trip planned by an anonymous user or a registered user
interface Trip {
    id: string;
    type: "anonymous" | "registered";

    userId?: string;
    name?: string;
    dateExpected: Datetime;
    dateCreated: Datetime;
}

// Data types
interface Datetime {
    year: number;
    month: number;
    day: number;
    hour: number;
    min: number;
    sec: number;
    timezone: string;
}