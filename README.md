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