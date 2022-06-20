package testcase_test

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"testing"
)

func ExampleStubTB_testingATestHelper() {
	stub := &testcase.StubTB{}

	myTestHelper := func(tb testing.TB) {
		tb.FailNow()
	}

	var tb testing.TB
	assert.Must(tb).Panic(func() {
		myTestHelper(stub)
	})
	assert.Must(tb).True(stub.IsFailed)
}