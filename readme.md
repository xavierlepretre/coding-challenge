# Coding Challenge

Your challenge is to build a fees API in encore (encore.dev) that uses a temporal (temporal.io) workflow started at the beginning of a fee period, and allows for the progressive accrual of fees. At the end of the billing period, the total invoice and bill summation should be available.

Requirements:

1. Able to create new bill
2. Able to add line item to an existing open bill
3. Able to close an active bill
    a. indicate total amount being charged
    b. indicate all line item being charged
4. Reject line item addition if bill is closed (bill already charged)
5. Able to query open and closed bill
6. Able to handle different types of currency, for the sake of simplicity, assume GEL and USD.

You are free to design the RESTFul API that will be used by other services. The above requirements are not exhaustive, you’re free to add in requirements that you think are helpful in making the Fees API more flexible and complete.

Your solution should:

* Use encore and temporalite (a local temporal)
* Use temporal signals
* Include a Readme
* Have unit tests

Hint: https://encore.dev/docs/how-to/temporal

Hint: you should use the Temporalite version to make it easy https://github.com/temporalio/temporalite

## Install

You need:

* Go version 1.24.1 or compatible.
* [Temporal CLI](https://docs.temporal.io/cli).
* Mockgen:

    ```sh
    go install github.com/golang/mock/mockgen@v1.6.0
    ```

## Run unit tests

```sh
go test ./pkg/model/... -v
go test ./pkg/workflow/... -v
```

Or:

```sh
docker run --rm -it -v $(pwd):/app -w /app golang:1.24.1 go test ./pkg/model/... -v
docker run --rm -it -v $(pwd):/app -w /app golang:1.24.1 go test ./pkg/workflow/... -v
```

For the Encore.dev part:

```sh
encore test ./...
```

Or:

```sh
docker run --rm -it -v $(pwd):/app -w /app encoredotdev/encore:1.46.1 test ./... -v
```

If you need to recreate the mocks:

```sh
go generate ./...
```

And do not forget to adjust the created file as per the comment in [`gen_command.go`](./pkg/rest/mocks/gen_command.go).

## Run a local live test

In terminal 1, launch Temporal CLI:

```sh
temporal server start-dev --namespace billing --db-filename temporal.db
```

In terminal 2, launch a billing worker:

```sh
go run main/billing_worker.go --task-queue local-billing
```

In terminal 3, launch the REST API:

```sh
encore run
```

### Create a new bill

In the [opened browser](http://localhost:9400/sfet4/requests):

* Pick `rest.OpenNewBill`.
* Enter request as:

    ```json
    {
        "currency_code": "USD",
        "close_time": "2025-03-31T23:59:59Z"
    }
    ```

* Use `token-alice` as your authentication data.
* Press <kbd>CALL API</kbd>

It should return something like:

```json
{"id":"4ba283ee-1d1d-4146-9b67-3dc5b2a21328"}
```

### Get the bill again

In the [opened browser](http://localhost:9400/sfet4/requests):

* Pick `rest.GetBill`.
* Enter path as: `/bill/4ba283ee-1d1d-4146-9b67-3dc5b2a21328` or whichever value you had in the previous step.
* Use `token-alice` as your authentication data.
* Press <kbd>CALL API</kbd>

It should return something like:

```json
{"id":"4ba283ee-1d1d-4146-9b67-3dc5b2a21328","currency_code":"USD","status":0,"line_item_count":0,"total_ok":"y","total":0}
```

### Add a line item

In the [opened browser](http://localhost:9400/sfet4/requests):

* Pick `rest.AddLineItem`.
* Enter path as: `/bill/4ba283ee-1d1d-4146-9b67-3dc5b2a21328/line_items` or whichever value you had in the previous step.
* Use `token-alice` as your authentication data.
* Enter request as:
    
    ```json
    {
        "description": "Matchbox",
        "amount": 100,
        "currency_code": "USD"
    }
    ```

* Press <kbd>CALL API</kbd>

It should return something like:

```json
{"id":"fb93e3c7-e2ae-4ce1-9e4b-023dde5d0185","currency_code":"USD","line_item_count":1,"total_ok":"y","total":100}
```

### Close the bill

In the [opened browser](http://localhost:9400/sfet4/requests):

* Pick `rest.CloseBill`.
* Enter path as: `/bill/4ba283ee-1d1d-4146-9b67-3dc5b2a21328/close` or whichever value you had in the previous step.
* Use `token-alice` as your authentication data.
* Press <kbd>CALL API</kbd>

It should return something like:

```json
{"currency_code":"USD","line_item_count":1,"total_ok":"y","total":100}
```
