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

    @patch("/{resource.id=users/*}")
    UpdateUser(req: UpdateUserRequest) : User;

    @put("/{resource.id=users/*}")
    ReplaceUser(req: ReplaceUserRequest) : User;

    @delete("/{id=users/*}")
    DeleteUser(req: DeleteUserRequest) : void;



    // resource type: Trip
    @post("/{parent=users/*}/trips" | "/trips")
    CreateTrip(req: CreateTripRequest) : Trip;

    @patch("/{resource.id=users/*/trips/*}" | "/{resource.id=trips/*}")
    UpdateTrip(req: UpdateTripRequest) : Trip;

    @put("/{resource.id=users/*/trips/*}" | "/{resource.id=trips/*}")
    ReplaceTrip(req: ReplaceTripRequest) : Trip;

    @get("/{id=users/*/trips/*}" | "/{id=trips/*}")
    GetTrip(req: GetTripRequest) : Trip;

    @get("{parent=users/*}/trips" | "/trips")
    ListTrips(req: ListTripRequest) : Trip[];

    @delete("/{id=users/*/trips/*}" | "/{id=trips/*}")
    DeleteTrip(req: DeleteTripRequest) : void;

    @post("/{id=users/*/trips/*}:plan" | "/{id=trips/*}:plan")
    PlanTrip(req: PlanTripRequest) : Operation<ResultT, MetadataT>; // custom method for planning the trip



    // Resource type: Point
    @post("/{parent=users/*/trips/*}/points" | "/{parent=trips/*}/points")
    CreatePoint(req: CreatePointRequest) : Point;

    @patch("/{resource.id=users/*/trips/*/points/*}" | "/{resource.id=trips/*/points/*}")
    UpdatePoint(req: UpdatePointRequest): Point;

    @put("/{resource.id=users/*/trips/*/points/*}" | "/{resource.id=trips/*/points/*}")
    ReplacePoint(req: ReplacePointRequest) : Point;

    @get("/{id=users/*/trips/*/points/*}" | "/{id=trips/*/points/*}")
    GetPoint(req: GetPointRequest) : Point;

    @get("/{parent=users/*/trips/*}/points" | "/{parent=trips/*}/points")
    ListPoints(req: ListPointRequest) : Point[];

    @delete("/{id=users/*/trips/*/points/*}" | "/{id=trips/*/points/*}")
    DeletePoint(req: DeletePointRequest) : void;

    @post("/{id=users/*/trips/*/points/*}:copy | {id=trips/*/points/*}:copy") // custom method for copy a point from one trip to another. Won't copy pointConstraints
    CopyPoint(req: CopyPointRequest) : Point;

    // singleton subresource: UserPassword
    @post("/{resource.id=users/*/password}")
    UpdateUserPassword(req: UpdateUserPasswordRequest) : UserPassword;

    @post("/{id=users/*/password}:reset")
    ResetUserPassword(req: ResetUserPasswordRequest) : ResetUserPasswordResponse;

    // Resource type: PointConstraint
    @post("/{parents=users/*/trips/*/points/*}/constraints" | "/{parent=trips/*/points/*}/constraints")
    CreatePointConstraint(req: CreatePointConstraintRequest) : PointConstraint;

    @patch("/{resource.id=users/*/trips/*/points/*/constraints/*}" | "/{resource.id=trips/*/points/*/constraints/*}")
    UpdatePointConstraint(req: UpdatePointConstraintRequest) : PointConstraint;

    @put("/{resource.id=users/*/trips/*/points/*/constraints/*}" | "/{resource.id=trips/*/points/*/constraints/*}")
    ReplacePointConstraint(req: ReplacePointConstraintRequest) : PointConstraint;

    @get("/{id=users/*/trips/*/points/*/constraints/*}" | "/{id=trips/*/points/*/constraints/*}")
    GetPointConstraint(req: GetPointConstraintRequest) : PointConstraint;

    @get("/{parent=users/*/trips/*/points/*}/constraints" | "/{parent=trips/*/points/*}/constraints")
    ListPointConstraints(req: ListPointConstraintsRequest) : PointConstraint[];

    @delete("/{id=users/*/trips/*/points/*/constraints/*}" | "/{id=trips/*/points/*/constraints/*}")
    DeletePointConstraint(req: DeletePointConstraintRequest) : void;
}


// **************************** Request/Response interface definitions *********************************
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
    parent?: string;
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
    parent?: string;
    filter: string;
}

interface DeleteTripRequest {
    id: string;
}

interface PlanTripRequest {
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

interface CreatePointRequest {
    parent: string;
    resource: Point;
}

interface UpdatePointRequest {
    resource: Point;
}

interface ReplacePointRequest {
    resource: Point;
}

interface GetPointRequest {
    id: string;
}

interface ListPointRequest {
    parent: string;
    filter: string;
}

interface DeletePointRequest {
    id: string;
}

interface CopyPointRequest {
    id: string;
    destinationTripId: string;
    destinationId: string;
}

interface CreatePointConstraintRequest {
    parent: string;
    resource: PointConstraint;    
}

interface UpdatePointConstraintRequest {
    resource: PointConstraint;
}

interface ReplacePointConstraintRequest {
    resource: PointConstraint;
}

interface GetPointConstraintRequest {
    id: string;
}

interface ListPointConstraintsRequest {
    parent: string;
    filter: string;
}

interface DeletePointConstraintRequest {
    id: string;
}


// ********************* Resource type definitions ************************
interface User {
    id: string;

    name: string;
    email: string;
    joinDate: Datetime;
}

// singleton subresource for password for security reason
interface UserPassword {
    id: string;

    value: string;
}

// Polymorphic resource, can be either trip planned by an anonymous user or a registered user
interface Trip {
    id: string;
    type: "anonymous" | "registered";

    userId?: string;
    name?: string;
    dateExpected: Datetime;
    dateCreated: Datetime;
    lastModified: Datetime;
    budgetLimit: Cost;
    preferredTransportMode: 'train' | 'bus' | 'walk';
}

interface Point {
    id: string;

    tripId: string;
    geoPointId: string;
    priority?: number;
}

// Polymorphic resource
interface PointConstraint {
    id: string;
    pointId: string;
    type: 'duration' | 'after' | 'before';
    value: PointDurationConstraint | PointAfterConstraint | PointBeforeConstraint;
}

interface Operation<ResultT, MetadataT> {
    id: string;

    done: Boolean;
    result: ResultT | OperationError
    metadata: MetadataT;
}

interface OperationError {
    code: string;
    messsage: string;
    details?: string;
}

// Data types
interface PointDurationConstraint {
    duration: number;
    unit: 'h' | 'm';
}

interface PointAfterConstraint {
    points: string[];
}

interface PointBeforeConstraint {
    points: string[];
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