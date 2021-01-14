package httpspec_test

import (
	"context"
	"net/http"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func ExampleLetContext_withValue() {
	s := testcase.NewSpec(testingT)

	httpspec.SubjectLet(s, func(t *testcase.T) http.Handler { return MyHandler{} })

	s.Before(func(t *testcase.T) {
		// this is ideal for representing middleware prerequisite
		// in the form of a value in the context that is guaranteed by a middleware.
		// Use this only if you cannot make it part of the specification level context value deceleration with ContextLet.
		ctx := t.I(httpspec.ContextVarName).(context.Context)
		ctx = context.WithValue(ctx, `foo`, `bar`)
		t.Let(httpspec.ContextVarName, ctx)
	})

	s.Test(`the *http.Request#Context() will have foo-bar`, func(t *testcase.T) {
		httpspec.SubjectGet(t)
	})
}