/* 
For this application, navigating between Trip and its child resources is best 
handled as a single page application. Therefore, the API will send back data as JSON 
when user is interacting with trip and its child resources (Point and PointConstraint) 
for client to process and update the single page.

To make it consistent for both client code and server code, this API will always return
JSON, and it is the responsibility of the client dev to read the doc of this API to 
know how to construct the next URL(s)

Admittedly, the fact that 1) client dev using this API have to read the API doc
to know how to use it and 2) no HATEOAS: the response is not hypermedia (think HTML or JSON+HAL)
but pure data (think pure JSON) implies that this is not a "true" REST API. However,
such is also 99% of APIs that claim to be RESTful today.

*/
abstract class TripPlannerApi {
    static version = "v1.0.0";
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

    @post("/{id=users/*/trips/*}:copy" | "/{id=trips/*}:copy")
    CopyTrip(req: CopyTripRequest) : Trip;

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


    // Resource type: GeoPoint, read-only: does not support write-based actions (create, replace, update, delete)
    @get("/{id=geopoints/*}")
    GetGeoPoint(req: GetGeoPointRequest) : GeoPoint;

    @get("/geopoints/*")
    ListGeoPoint(req: ListGeoPointRequest) : GeoPoint[];

    // singleton subresource: Email
    @post("/{resource.id=users/*/email}:change")
    ChangeEmail(req: ChangeEmailRequest) : void;

    // singleton subresource: Password
    @post("/{resource.id=users/*/password}:change")
    ChangePassword(req: ChangePasswordRequest) : void;    

    @post("/{id=users/*/password}:reset")
    ResetPassword(req: ResetPasswordRequest) : void;
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

interface ChangeEmailRequest {
    id: string;

    oldEmail: string;
    password: string;
    newEmail: string;
}

interface ChangePasswordRequest {
    id: string;

    oldPassword: string;
    newPassowrd: string;
}

interface ResetPasswordRequest {
    id: string;
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

interface GetGeoPointRequest {
    id: string;
}

interface ListGeoPointRequest {
    filter: string;
}

interface CopyTripRequest {
    id: string;
    destinationParentId: string;
    destinationId: string;
}


/***********************************************
            Resource type definitions
***********************************************/
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
    planResult: Edge[];
}

interface Point {
    id: string;

    tripId: string;
    geoPointId: string;
    arrivalConstraint: PointArrivalConstraint;
    durationConstraint: PointDurationConstraint;
    beforeConstraint: PointBeforeConstraint;
    afterConstraint: PointAfterConstraint;
}

interface GeoPoint {
    id: string;

    lat: string;
    lon: string;
    tags: KeyValuePair[];
}

// Polymorphic resource
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
interface PointArrivalConstraint {
    from: Datetime;
    to: Datetime;
}

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
    unit: 'JPY' | 'USD';
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
    timezone: string;
}

interface Edge {
    pointId: string;
    nextPointId: string;
    start: Datetime;
    duration: Duration;
    cost: Cost;
    transportMode: 'train' | 'bus' | 'walk'
}

interface KeyValuePair {
    key: string;
    value: string;
}