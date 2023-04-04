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

    // resource type: Trip
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

    // singleton subresource: UserPassword
    @patch("/{resource.id=users/*/password}")
    UpdateUserPassword(req: UpdateUserPasswordRequest) : UserPassword;

    @post("/{id=users/*/password}:reset")
    ResetUserPassword(req: ResetUserPasswordRequest) : ResetUserPasswordResponse;
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

interface ResetUserPasswordRequest {
    id: string;

    email: string;
}

interface ResetUserPasswordResponse {
    email: string;
}

interface UpdateUserPasswordRequest {
    resource: UserPassword;
}

interface User {
    id: string;

    name: string;
    email: string;
    joinDate: Datetime;
}

// singleton subresource for password for security reason
interface UserPassword {
    id: string;

    oldPassword: string;
    password: string;
}

// Polymorphic resource, can be either trip planned by an anonymous user or a registered user
interface Trip {
    id: string;
    type: "anonymous" | "registered";

    userId?: string;
    name?: string;
    dateExpected: Datetime;
    dateCreated: Datetime;
    budgetLimit: Cost;
    preferredTransportMode: 'train' | 'bus' | 'walk';
}

interface Point {
    id: string;

    tripId: string;
    geoPointId: string;
    constraints: PointConstraints;
    priority?: number;
}

// Data types
interface PointConstraints {
    timeSpendExpected: Duration;

    afterPoints: string[];
}

interface Cost {
    amount: number;
    unit: string;
}

interface Duration {
    hour: number;
    minute: number;
}

interface Datetime {
    year: number;
    month: number;
    day: number;
    hour: number;
    min: number;
    sec: number;
    timezone: string;
}