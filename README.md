# go-test-runner

A simple tool to select individual Go subtests to execute.

Dependencies:

- [fzf](https://github.com/junegunn/fzf)

```
go get github.com/benjaminheng/go-test-runner
```

## Demo

Invoke `go-test-runner` with a package list (`./...`) or a glob pattern
(`./mypackage/*.go`).

```
$ go-test-runner ./...
```

A fzf process will be started showing the tests and subtests parsed from the
input files.

```
  TestOfferUpdater_UpdateOffersByID$/offers_are_updated$
  TestOfferUpdater_UpdateOffersByID$/validate$/invalid_column_in_preconditions$
  TestOfferUpdater_UpdateOffersByID$/validate$/invalid_column_in_changemap$
  TestOfferUpdater_UpdateOffersByID$/validate$/no_changemap$
  TestOfferUpdater_UpdateOffersByID$/validate$/no_ids$
  TestOfferUpdater_UpdateOffersByID$/validate$
  TestOfferUpdater_UpdateOffersByID$
  TestOfferUpdater_UpdateOffer$/offer_is_updated$
  TestOfferUpdater_UpdateOffer$/validate$/invalid_offer$
  TestOfferUpdater_UpdateOffer$/validate$/nil_offer$
> TestOfferUpdater_UpdateOffer$/validate$
  TestOfferUpdater_UpdateOffer$
  85/85
>
```

In this example we select `TestOfferUpdater_UpdateOffer$/validate$`. This will
be passed to the `-run` flag in `go test`. The full command:

```
go test -v ./... -run TestOfferUpdater_UpdateOffer$/validate$
```

Only the selected subtests are executed. The command output:

```
=== RUN   TestOfferUpdater_UpdateOffer
=== RUN   TestOfferUpdater_UpdateOffer/validate
=== RUN   TestOfferUpdater_UpdateOffer/validate/nil_offer
=== RUN   TestOfferUpdater_UpdateOffer/validate/invalid_offer
--- PASS: TestOfferUpdater_UpdateOffer (0.00s)
    --- PASS: TestOfferUpdater_UpdateOffer/validate (0.00s)
        --- PASS: TestOfferUpdater_UpdateOffer/validate/nil_offer (0.00s)
        --- PASS: TestOfferUpdater_UpdateOffer/validate/invalid_offer (0.00s)
PASS
ok      command-line-arguments  0.058s
```
