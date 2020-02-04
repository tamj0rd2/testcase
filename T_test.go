package testcase_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/adamluzsi/testcase"
)

func TestT_Let_canBeUsedDuringTest(t *testing.T) {
	s := testcase.NewSpec(t)

	s.Context(`runtime define`, func(s *testcase.Spec) {
		s.Let(`n-original`, func(t *testcase.T) interface{} { return rand.Intn(42) })
		s.Let(`m-original`, func(t *testcase.T) interface{} { return rand.Intn(42) + 100 })

		var exampleMultiReturnFunc = func(t *testcase.T) (int, int) {
			return t.I(`n-original`).(int), t.I(`m-original`).(int)
		}

		s.Context(`Let being set during test runtime`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				n, m := exampleMultiReturnFunc(t)
				t.Let(`n`, n)
				t.Let(`m`, m)
			})

			s.Test(`let values which are defined during runtime present in the test`, func(t *testcase.T) {
				require.Equal(t, t.I(`n`), t.I(`n-original`))
				require.Equal(t, t.I(`m`), t.I(`m-original`))
			})
		})
	})

	s.Context(`runtime update`, func(s *testcase.Spec) {
		var initValue = rand.Intn(42)
		s.Let(`x`, func(t *testcase.T) interface{} { return initValue })

		s.Before(func(t *testcase.T) {
			t.Let(`x`, t.I(`x`).(int)+1)
		})

		s.Before(func(t *testcase.T) {
			t.Let(`x`, t.I(`x`).(int)+1)
		})

		s.Test(`let will returns the value then override the runtime variables`, func(t *testcase.T) {
			require.Equal(t, initValue+2, t.I(`x`).(int))
		})
	})

}

func TestT_Defer(t *testing.T) {
	s := testcase.NewSpec(t)

	var res []int

	s.Context(``, func(s *testcase.Spec) {
		s.Before(func(t *testcase.T) {
			res = append(res, 0)
		})

		s.After(func(t *testcase.T) {
			res = append(res, -1)
		})

		s.Context(``, func(s *testcase.Spec) {
			s.Around(func(t *testcase.T) func() {
				res = append(res, 1)
				return func() { res = append(res, -2) }
			})

			s.Context(``, func(s *testcase.Spec) {
				s.Let(`with defer`, func(t *testcase.T) interface{} {
					t.Defer(func() { res = append(res, -3) })
					return 42
				})

				s.Before(func(t *testcase.T) {
					// calling a variable that has defer will ensure
					// that the deferred function call will be executed
					// as part of the *T#defer stack, and not afterwards
					require.Equal(t, 42, t.I(`with defer`).(int))
				})

				s.Test(``, func(t *testcase.T) {
					t.Defer(func() { res = append(res, -4) })
				})
			})
		})
	})

	require.Equal(t, []int{0, 1, -4, -3, -2, -1}, res)
}

func TestT_Defer_willRunEvenIfSomethingForceTheTestToStopEarly(t *testing.T) {
	s := testcase.NewSpec(t)
	var ran bool
	s.Before(func(t *testcase.T) { t.Defer(func() { ran = true }) })
	s.Test(``, func(t *testcase.T) { t.Skip(`please stop early`) })
	require.True(t, ran)
}