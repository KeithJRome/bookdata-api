package main

import (
	"testing"

	"github.com/matt-FFFFFF/bookdata-api/loader"
)

// This test was to prove that without a mutex lock iterating and updating the books list could create a race condition.
func TestRace(t *testing.T) {

	go func() {
		books.GetAllBooks(0, 0)
	}()

	go func() {
		book := loader.BookData{
			BookID: "test001",
			ISBN:   "ABCDEFGHIJ",
		}
		books.InsertBook(&book)
	}()

}

/*
BEFORE FIX, THIS PRODUCED:

==================
WARNING: DATA RACE
Write at 0x00c00000e060 by goroutine 13:
  github.com/matt-FFFFFF/bookdata-api/datastore.(*Books).InsertBook()
      /workspaces/keith-bookdata-api/datastore/memory.go:167 +0x265
  github.com/matt-FFFFFF/bookdata-api.TestRace.func2()
      /workspaces/keith-bookdata-api/race_test.go:21 +0xb4

Previous read at 0x00c00000e060 by goroutine 12:
  github.com/matt-FFFFFF/bookdata-api/datastore.applyLimitAndSkipToBooks()
      /workspaces/keith-bookdata-api/datastore/memory.go:82 +0x56
  github.com/matt-FFFFFF/bookdata-api/datastore.(*Books).GetAllBooks()
      /workspaces/keith-bookdata-api/datastore/memory.go:106 +0x2f
  github.com/matt-FFFFFF/bookdata-api.TestRace.func1()
      /workspaces/keith-bookdata-api/race_test.go:13 +0x5a

Goroutine 13 (running) created at:
  github.com/matt-FFFFFF/bookdata-api.TestRace()
      /workspaces/keith-bookdata-api/race_test.go:16 +0x5a
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:991 +0x1eb

Goroutine 12 (finished) created at:
  github.com/matt-FFFFFF/bookdata-api.TestRace()
      /workspaces/keith-bookdata-api/race_test.go:12 +0x42
  testing.tRunner()
      /usr/local/go/src/testing/testing.go:991 +0x1eb
==================
Found 1 data race(s)
exit status 66
FAIL    github.com/matt-FFFFFF/bookdata-api     0.158s
*/
