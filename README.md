## Entain BE Technical Test

This test has been designed to demonstrate your ability and understanding of technologies commonly used at Entain. 

Please treat the services provided as if they would live in a real-world environment.

### Directory Structure

- `api`: A basic REST gateway, forwarding requests onto service(s).
- `racing`: A very bare-bones racing service.

```
entain/
├─ api/
│  ├─ proto/
│  ├─ main.go
├─ racing/
│  ├─ db/
│  ├─ proto/
│  ├─ service/
│  ├─ main.go
├─ README.md
```

### Getting Started

1. Install Go (latest).

```bash
brew install go
```

... or [see here](https://golang.org/doc/install).

2. Install `protoc`

```
brew install protobuf
```

... or [see here](https://grpc.io/docs/protoc-installation/).

2. In a terminal window, start our racing service...

```bash
cd ./racing

go build && ./racing
➜ INFO[0000] gRPC server listening on: localhost:9000
```

3. In another terminal window, start our api service...

```bash
cd ./api

go build && ./api
➜ INFO[0000] API server listening on: localhost:8000
```

4. Make a request for races... 

```bash
curl -X "POST" "http://localhost:8000/v1/list-races" \
     -H 'Content-Type: application/json' \
     -d $'{
  "filter": {}
}'
```

### Changes/Updates Required

- We'd like to see you push this repository up to **GitHub/Gitlab/Bitbucket** and lodge a **Pull/Merge Request for each** of the below tasks.
- This means, we'd end up with **5x PR's** in total. **Each PR should target the previous**, so they build on one-another.
- Alternatively you can merge each PR/MR after each other into master.
- This will allow us to review your changes as well as we possibly can.
- As your code will be reviewed by multiple people, it's preferred if the repository is **publicly accessible**. 
- If making the repository public is not possible; you may choose to create a separate account or ask us for multiple email addresses which you can then add as viewers. 

... and now to the test! Please complete the following tasks.

1. Add another filter to the existing RPC, so we can call `ListRaces` asking for races that are visible only.
   > We'd like to continue to be able to fetch all races regardless of their visibility, so try naming your filter as logically as possible. https://cloud.google.com/apis/design/standard_methods#list
2. We'd like to see the races returned, ordered by their `advertised_start_time`
   > Bonus points if you allow the consumer to specify an ORDER/SORT-BY they might be after. 
3. Our races require a new `status` field that is derived based on their `advertised_start_time`'s. The status is simply, `OPEN` or `CLOSED`. All races that have an `advertised_start_time` in the past should reflect `CLOSED`. 
   > There's a number of ways this could be implemented. Just have a go!
4. Introduce a new RPC, that allows us to fetch a single race by its ID.
   > This link here might help you on your way: https://cloud.google.com/apis/design/standard_methods#get
5. Create a `sports` service that for sake of simplicity, implements a similar API to racing. This sports API can be called `ListEvents`. We'll leave it up to you to determine what you might think a sports event is made up off, but it should at minimum have an `id`, a `name` and an `advertised_start_time`.

> Note: this should be a separate service, not bolted onto the existing racing service. At an extremely high-level, the diagram below attempts to provide a visual representation showing the separation of services needed and flow of requests.
> 
> ![](example.png)


**Don't forget:**

> Document and comment! Please make sure your work is appropriately documented/commented, so fellow developers know whats going on.

**Note:**

To aid in proto generation following any changes, you can run `go generate ./...` from `api` and `racing` directories.

Before you do so, please ensure you have the following installed. You can simply run the following command below in each of `api` and `racing` directories.

```
go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 google.golang.org/genproto/googleapis/api google.golang.org/grpc/cmd/protoc-gen-go-grpc google.golang.org/protobuf/cmd/protoc-gen-go
```

### Good Reading

- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Google API Design](https://cloud.google.com/apis/design)
- [Go Modules](https://golang.org/ref/mod)
- [Ubers Go Style Guide](https://github.com/uber-go/guide/blob/2910ce2e11d0e0cba2cece2c60ae45e3a984ffe5/style.md)



---------

## Changes

### Task 1: only visible filter

`ListRaces` accepts an `onlyVisible` filter. Set it to `true` to get visible races only. Leave it out, or set it to `false`, to get all races as before.

```bash
curl -X "POST" "http://localhost:8000/v1/list-races" \
     -H 'Content-Type: application/json' \
     -d $'{"filter": {"onlyVisible": true}}'
```

Checked against the seeded database (100 races, 54 visible):

| Request | Before | After |
| --- | --- | --- |
| `{"filter": {}}` | 100 | 100 |
| `{"filter": {"onlyVisible": true}}` | 100 | 54 |

#### Changing the filter

A request goes through two services: `api` receives the JSON, then passes it on to `racing`. Each service has its own copy of the proto file:

- `api/proto/racing/racing.proto`
- `racing/proto/racing/racing.proto`

When you add or change a filter field, please update both files, then run `go generate ./...` in `api` and in `racing`.

If `api/proto/racing/racing.proto` doesn't have the field, `api` ignores it without any error, so the filter does nothing and every race comes back. That's why `onlyVisible: true` returned all 100 races before this change.

#### Tests

`racing/db/races_filter_test.go` tests the filter in two ways:

- A unit test checks the SQL that `applyFilter` builds.
- An integration test runs `List` against an in-memory SQLite database with three fixed races, so the result doesn't depend on the seeded data.

Run them from the project root:

```bash
cd racing && go test ./...
```

### Task 2: order races by advertised start time

`ListRaces` accepts an `orderBy` field. It uses the `order_by` format from [AIP-132](https://google.aip.dev/132#ordering): field names separated by commas, each one optionally followed by ` desc`.

```bash
curl -X "POST" "http://localhost:8000/v1/list-races" \
     -H 'Content-Type: application/json' \
     -d $'{"orderBy": "meeting_id, advertised_start_time desc"}'
```

- Use the proto field names: `id`, `meeting_id`, `name`, `number`, `visible`, `advertised_start_time`.
- Only lowercase `desc` is accepted. Leave it out to sort in ascending order.
- When `orderBy` is empty, races are sorted by `advertised_start_time`, earliest first.
- `id` is always the last sort key, so races with the same values come back in the same order every time.
- Start times are compared with SQLite's `julianday()`, so times with different UTC offsets are ordered by the actual time, not as text.
- An unknown field, a field listed twice, or any other direction returns `400` (gRPC `INVALID_ARGUMENT`) instead of being ignored.

Before, races came back in `id` order. Now they come back by start time:

| Position | Before | After |
| --- | --- | --- |
| 1st | id 1, `2021-03-03T01:30:57Z` | id 23, `2021-02-28T06:46:12Z` |
| 2nd | id 2, `2021-03-02T19:16:58Z` | id 75, `2021-02-28T07:00:50Z` |
| 3rd | id 3, `2021-03-02T08:30:16Z` | id 76, `2021-02-28T07:09:10Z` |

An invalid value is rejected:

| Request | Response |
| --- | --- |
| `{"orderBy": "price"}` | `400`, `unsupported order_by field "price"` |

#### Code changes

- `RacesRepo.List` in `racing/db/races.go` now takes an `orderBy` argument after the filter. Anything that calls `List` needs to pass it.
- `ListRaces` in `racing/service/racing.go` checks `order_by` and passes it to `List`. Invalid values are rejected here, before the database is called.

#### Tests

`racing/db/races_test.go` is renamed to `racing/db/races_filter_test.go`, so each test file is named after what it tests.

- `racing/db/races_order_test.go`: a unit test checks the `ORDER BY` that `applyOrder` builds. An integration test runs `List` against four races in an in-memory SQLite database. Their start times use different UTC offsets, and two of them start at the same time, to check the time comparison and the `id` tie-break.
- `racing/service/racing_order_test.go`: checks how `parseOrderBy` reads `order_by`, and that `ListRaces` returns `INVALID_ARGUMENT` for `price` without calling the database.

Run them the same way as task 1.

### Task 3: race status

Every race now has a `status`: `CLOSED` if its `advertised_start_time` is in the past, otherwise `OPEN`. A race that starts exactly now is still `OPEN`.

```json
{
  "id": "23",
  "meetingId": "6",
  "name": "Idaho wolves",
  "number": "10",
  "visible": false,
  "advertisedStartTime": "2021-02-28T06:46:12Z",
  "status": "CLOSED"
}
```

- The status is worked out each time races are listed, using the current time. It isn't stored in the database.
- Every race in the seeded database starts in 2021, so they all come back as `CLOSED`. `OPEN` is covered by the tests below.
- You can't filter or sort by `status`. The default order, `advertised_start_time`, already lists `CLOSED` races before `OPEN` ones.

| | Before | After |
| --- | --- | --- |
| `status` of the 100 seeded races | Not in the response | `CLOSED` for all 100 |

#### Code changes

- `Race` in both proto files has a new `status` field. Its values are `OPEN` and `CLOSED`. `STATUS_UNSPECIFIED` is only the default for an unset field; this service always sets `OPEN` or `CLOSED`.
- `ListRaces` in `racing/service/racing.go` sets the status of each race. It reads the current time once per request, so races with the same start time always get the same status.

#### Tests

`racing/service/racing_status_test.go`:

- `TestRaceStatus` checks the rule against a fixed time: one second before is `CLOSED`, exactly now and one second after are `OPEN`.
- `TestListRacesSetsStatus` checks that `ListRaces` sets the status of every race. The start times are in 2000 and 2999, so the result doesn't depend on when the test runs.

#### CI

`.github/workflows/test.yml` runs on every pull request with Go 1.27.4. It runs `go vet` and `go test` in `racing`, and `go build` in `api`.

### Task 4: get a single race

`GetRace` returns one race by its id. Through the REST gateway it is `GET /v1/races/{id}`:

```bash
curl "http://localhost:8000/v1/races/1"
```

```json
{
  "id": "1",
  "meetingId": "5",
  "name": "North Dakota foes",
  "number": "2",
  "visible": false,
  "advertisedStartTime": "2021-03-03T01:30:57Z",
  "status": "CLOSED"
}
```

- The response is the race itself, in the same shape as each race from `ListRaces`, including `status`.
- A race is returned whether or not it is visible.
- An id of zero or less returns `400` instead of `404`. Zero is the default value in proto3, so it usually means the caller forgot to set the id.

| Request | Before | After |
| --- | --- | --- |
| `GET /v1/races/1` | `404`, the route didn't exist | `200`, race 1 |

Errors:

| Request | Response |
| --- | --- |
| `GET /v1/races/999` | `404`, `race 999 not found` |
| `GET /v1/races/0` | `400`, `id must be greater than zero, got 0` |
| `GET /v1/races/abc` | `400`, from the gateway, because the id isn't a number |

#### Code changes

- `GetRace` and its request message, `GetRaceRequest`, are added to `api/proto/racing/racing.proto` and `racing/proto/racing/racing.proto`.
- Only `api/proto/racing/racing.proto` has the `GET /v1/races/{id}` route, because the `api` service is the one that turns HTTP requests into gRPC calls to `racing`.
- `RacesRepo` in `racing/db/races.go` has a new `Get` method, which returns `ErrNotFound` when no race has the id. It uses the same query as `List` with `WHERE id = ?`. `List` itself is unchanged.
- `GetRace` in `racing/service/racing.go` checks the id, calls `Get`, and sets the status the same way as `ListRaces`.
- The fake repositories in `racing/service/racing_order_test.go` and `racing/service/racing_status_test.go` now have a `Get` method, because `RacesRepo` requires it.

#### Tests

- `racing/db/races_get_test.go`: runs `Get` against two races in an in-memory SQLite database. It checks that a race is returned even when it isn't visible, and that a missing id returns `ErrNotFound`.
- `racing/service/racing_get_test.go`: checks that `GetRace` sets the status, returns `NOT_FOUND` for a missing race, and returns `INVALID_ARGUMENT` for a zero or negative id without calling the database.

### Task 5: sports service

`sports` is a separate gRPC service with one RPC, `ListEvents`. The `api` service sends `POST /v1/list-events` to it, the same way it sends `POST /v1/list-races` to `racing`.

Start all three services, each in its own terminal:

```bash
cd ./racing && go build && ./racing
cd ./sports && go build && ./sports
cd ./api && go build && ./api
```

Then list the events:

```bash
curl -X "POST" "http://localhost:8000/v1/list-events" \
     -H 'Content-Type: application/json' \
     -d $'{}'
```

```json
{
  "events": [
    {
      "id": "1",
      "name": "Philadelphia Phillies vs Chicago Cubs",
      "advertisedStartTime": "2026-11-25T13:31:18Z",
      "sport": "Baseball",
      "competition": "MLB",
      "homeTeam": "Chicago Cubs",
      "awayTeam": "Philadelphia Phillies"
    }
  ]
}
```

The real response has 100 events; this shows the first one.

- An event has `id`, `name`, `advertised_start_time`, `sport`, `competition`, `home_team` and `away_team`. The teams have their own fields, so clients don't need to split `name`, which is only for display. `sport` and `competition` are plain text, so adding one doesn't need a proto change.
- `sports/db/sports.db` holds 100 MLB games between the 30 teams, made once by the seed. It is committed, like `racing/db/racing.db`, so running `sports` gives the same games as the example above.
- The tests don't use `sports/db/sports.db`. They build their own data in an in-memory database, so they don't depend on the date.
- `ListEvents` has no filter or ordering. Tasks 1 to 4 were asked for racing, not for sports.
- `sports` listens on `localhost:9001`, and `api` sends sports requests to that address. To use a different address, pass the new address to both services, for example:

  ```bash
  ./sports -grpc-endpoint=localhost:9101
  ./api -sports-grpc-endpoint=localhost:9101
  ```

#### Code changes

- `sports/` is a new Go module with the same layout as `racing/`: `main.go`, `service/sports.go`, `db/` and `proto/`.
- `api/proto/sports/sports.proto` is the `api` copy of the sports proto, with the `POST /v1/list-events` route. `api/main.go` registers it.
- `.github/workflows/test.yml` now also runs `go vet` and `go test` in `sports`.
- `.gitignore` ignores the `sports/sports` binary.

#### Tests

- `sports/db/events_list_test.go`: runs `List` against two events in an in-memory SQLite database and checks every field, including a start time with a `+10:00` offset.
- `sports/db/db_seed_test.go`: runs the seed on an empty in-memory database and checks that it makes 100 MLB games, each with two different MLB teams as home and away, a name made from them, and a start time within the expected range.
- `sports/service/sports_list_test.go`: checks that `ListEvents` returns the events from the repository, and that a database error reaches the client as `INTERNAL` without the database's own message.

Run them with `cd sports && go test ./...`.

### Out of scope

These are in the original code and aren't part of the five tasks, so they're left as they are to keep each PR to its task. `sports` doesn't repeat them.

1. **Racing can leave database connections open.** In `racing/db/races.go`, `List` never calls `rows.Close()`. Rows close by themselves once every row is read, but if scanning returns early with an error, the connection stays open, and these can build up. `Get`, added in Task 4, closes its rows with `defer`. *Found while reading the racing code before Task 1.*
2. **The generator install command doesn't install the generators on current Go.** The Note above the Changes section uses `go get`, which since Go 1.18 only updates `go.mod`. `go install` is needed instead. *Found while setting up the project before Task 1.*
3. **Racing can return part of the results as if they were all of them.** In `racing/db/races.go`, `scanRaces` doesn't check `rows.Err()` after the loop. If reading stops because of an error, the races read so far are returned with no error. *Found while writing the `sports` repository in Task 5.*
4. **The racing seed can hide a failed insert.** In `racing/db/db.go`, `seed` carries on after an error and only returns the result of the last insert, so the service can start with races missing and no error. *Found while writing the `sports` seed in Task 5.*
