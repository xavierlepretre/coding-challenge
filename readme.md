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

## Run unit tests

```sh
go test ./... -v
```

Or:

```sh
docker run --rm -it -v $(pwd):/app -w /app golang:1.24.1 go test ./... -v
```