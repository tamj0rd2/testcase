# testcase [GoDoc](https://godoc.org/github.com/adamluzsi/testcase)

## Yes, but why?

I made a list of requirements for myself, and then looked trough the available testing frameworks in golang:
* works perfectly well with `go test` command out of the box
* allow me to run one test edge case easily from the specification
* don't build singleton objects outside of my test function scope
* allow me to run test cases in concurrent execution for specification where I know that no side effect expected.
  * this is especially important me, because I love quick test feedback loops
* allow me to define variables in a way, that they receive concrete value later
  * this help me build spec coverage, where if I forgot a edge case regarding a variable, the spec will simply panic about early on.
* I want to use [stretchr/testify](https://github.com/stretchr/testify), so assertions not necessary for me
  * or more precisely, I needed something that guaranteed to allow me the usage of that pkg

While I liked the solutions, I felt that the way I would use them would leave out one or more point from my requirements.
So I ended up making a small design about how it would be great for me to test.
I took great inspiration from [rspec](https://github.com/rspec/rspec),
as I loved the time I spent working with that framework.

This is how this pkg is made.

The package coverage is 100%, and considered stable.

This package implements two approaches to help you to do nested
"BDD" style testing in golang.

## The Spec based approach

spec structure is a simple wrapping around the testing.T#Run.
It does not use any global singleton cache object or anything like that.
It does not force you to use global variables.

It uses the same idiom as the core go testing pkg also provide you.
You can use the same way as the core testing pkg
> go run ./... -v -run "the/name/of/the/test/it/print/out/in/case/of/failure"

It allows you to do context preparation for each test in a way,
that it will be safe for use with testing.T#Parallel.

You receive a new pointer structure called `*testcase.V`
which will represent values that you configured for your test case.
As mentioned above, the values in `*testcase.V` are safe to use during T#Parallel,
so as long your construct does not have any side-effect,
you are free to make run your code on concurrent goroutines.

In case you work with something that depends on side-effects,
such as database tests, you can use the hooks,
to create clean-ahead / clean-up blocks.

Also if you use gomock, you can use the spec#Around function,
to set up the mock with a controller, and in the teardown function,
call the gomock.Controller#Finish function,
so your test cases will be only about
what is the different behavior from the rest of the test cases.

It will panic if you use hooks or variable preparation in an ambiguous way,
or when you try to access variable that doesn't exist in the context where you do so.
It tries to panic with friendly and supportive messages, but that is highly subjective.

### My Rule of Thumbs

#### Subject of the Describe

To me, I found it useful, that I always created a `subject`/`asResult` variable
with a function that takes `*testcase.V` right after each Spec#Describe function block.
This function signature always shared the same signature as the function/method I test within it.
This also help me force myself to build up the right context that the subject block depends on in a form of intput.

It is also really helped me to have more descriptive test cases, easier refactoring and in my opinion an easy way to setup edge cases by using `testcase.Spec#Let`.
You can see an example of this in the GoDoc.

#### each when/and has its own Let or Before/Around to setup the testing context 

When I create when/and block, I describe the reason for the context,
and then add a Let or a Before/Around that setup the testing context according to the description.

and describe it in the description of what test runtime context I wanted to create by that.
If you have a dependency object that not exist in the first level of the nesting,
don't worry, because using `*testcase.Spec#Let` allow you to do it later,
in the right context.

#### Black-box testing
 
I usually only test exported functions, so to me black-box testing worked out the best.
Trough this I tend to write specs that feel more like examples in the end about the usage,
and I'm forced to use the pkg as a user of that pkg.
> to do black-box testing, just append _test to your current pkg name where you do the testing.

#### each if represented with two `When`/`And` block

When the code requires an if,
I usually try to create a context with `when`/`and` blocks,
to justify and describe when can that if path triggered, and how.

When the specification complexity becomes too big,
that is usually a sign to me that the component has a big responsibility (not SRP).

I usually then read through the specs,
and then extract nested loops into a separate structures/funcs,
and refer to those dependencies through as an interface.
By this the required mind model can be made smaller.

Based on this assumption, the size and complexity of the specification
is usually in 1:1 ratio with the size of the mind model needed to understand the code.

#### Cover Repetitive test cases with shared specification
  
Sometimes however it is unavoidable to repeat test coverage in different testing contexts,
and for those cases, I usually create a function that takes *testcase.Spec as a receiver,
and do the specification in that function, so it can be referenced from many places.

Such a typical example for that is when you need to test error cases,
and then in the error cases shared spec you swap out the dependency that is fallible 
with a mock through using the `Let`,
then you can setup expectations with `Before`/`Around`

### Example

In the examples the subject of the will be a dummy structure

```go
type MyType struct {
    Field1 string
}

func (mt *MyType) IsLower() bool {
    return strings.ToLower(mt.Field1) == mt.Field1
}

func (mt *MyType) Fallible() (string, error) {
	return "", nil
}
```

#### Basic example with Describe+When+Then

```go
func TestMyType(t *testing.T) {
    s := testcase.NewSpec(t)

    // when no side effect expected,
    // you can use Spec#Parallel for make all test edge case run on different goroutine
    s.Parallel()

    myType := func(v *testcase.V) *MyType {
        return &MyType{Field1: v.I(`input`).(string)}
    }

    s.Describe(`IsLower`, func(s *testcase.Spec) {
        subject := func(v *testcase.V) bool { return myType(v).IsLower() }

        s.When(`input string has lower case characters`, func(s *testcase.Spec) {
            s.Let(`input`, func(v *testcase.V) interface{} { return `all lower case` })

            s.Then(`it will return true`, func(t *testing.T, v *testcase.V) {
                t.Parallel()

                if subject(v) != true {
                    t.Fatalf(`it was expected that the %q will re reported to be lowercase`, v.I(`input`))
                }
            })

            s.And(`the first character is capitalized`, func(s *testcase.Spec) {
                s.Let(`input`, func(v *testcase.V) interface{} { return `First character is uppercase` })

                s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
                    if subject(v) != false {
                        t.Fatalf(`it was expected that %q will be reported to be not lowercase`, v.I(`input`))
                    }
                })
            })
        })
    })
}
```

#### Variables

in your spec, you can use the `*testcase.V` object,
for fetching values for your objects.
Using them is gives you the ability to create value for them,
only when you are in the right testing scope that responsible 
for providing an example for the expected value.

```go
func ExampleSpec_Let(t *testing.T) {
	myType := func(v *testcase.V) *MyType { return &MyType{Field1: v.I(`input`).(string)} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(v *testcase.V) bool { return myType(v).IsLower() }

		s.When(`input characters are all lowercase`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} {
				return "all lowercase"
			})

			s.Then(`it will report true`, func(t *testing.T, v *testcase.V) {
				require.True(t, subject(v))
			})
		})

		s.When(`input is a capitalized`, func(s *testcase.Spec) {
			s.Let(`input`, func(v *testcase.V) interface{} {
				return "Capitalized"
			})

			s.Then(`it will report false`, func(t *testing.T, v *testcase.V) {
				require.False(t, subject(v))
			})
		})
	})
}
```

if your variable can fail, you can use the *V#T function to retrieve the current test run `*testing.T` object.

```go
s.Let(`input`, func(v *testcase.V) interface{} {
	require.True(v.T(), true, `my important test assertion regarding this input variable`)
    return "value"
})
```

#### Hooks 

Hooks help you setup common things for each test case.
For example clean ahead, clean up, mock expectation configuration,
and similar things can be done in hooks,
so your test case blocks with `Then` only represent the expected result(s).

##### Before

Before give you the ability to run a block before each test case.
This is ideal for doing clean ahead before each test case.
The received *testing.T object is the same as the Then block *testing.T object
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
func ExampleSpec_Before(t *testing.T) {
	myType := func(input string) *MyType { return &MyType{Field1: input} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(input string) bool { return myType(input).IsLower() }

		s.Before(func(t *testing.T, v *testcase.V) {
			// this will run before the test cases.
			// this hook applied to this scope and anything that is nested from here.
			// hooks can be stacked with each call.
		})

		s.Then(`it will report whether Field1 is lower or not`, func(t *testing.T, v *testcase.V) {
			require.True(t, subject(`all lower case character`))
			require.False(t, subject(`Capitalized`))
		})
	})
}
```

##### After 

After give you the ability to run a block after each test case.
This is ideal for running cleanups.
The received *testing.T object is the same as the Then block *testing.T object
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
func ExampleSpec_After(t *testing.T) {
	myType := func(input string) *MyType { return &MyType{Field1: input} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(input string) bool { return myType(input).IsLower() }

		s.After(func(t *testing.T, v *testcase.V) {
			// this will run after the test cases.
			// this hook applied to this scope and anything that is nested from here.
			// hooks can be stacked with each call.
		})

		s.Then(`it will report whether Field1 is lower or not`, func(t *testing.T, v *testcase.V) {
			require.True(t, subject(`all lower case character`))
			require.False(t, subject(`Capitalized`))
		})
	})
}
```

##### Around

Around give you the ability to create "Before" setup for each test case,
with the additional ability that the returned function will be deferred to run after the Then block is done.
This is ideal for setting up mocks, and then return the assertion request calls in the return func.
This hook applied to this scope and anything that is nested from here.
All setup block is stackable.

```go
func ExampleSpec_Around(t *testing.T) {
	myType := func(input string) *MyType { return &MyType{Field1: input} }

	s := testcase.NewSpec(t)

	s.Describe(`IsLower`, func(s *testcase.Spec) {
		subject := func(input string) bool { return myType(input).IsLower() }

		s.Around(func(t *testing.T, v *testcase.V) func() {
			// this will run before the test cases

			// this hook applied to this scope and anything that is nested from here.
			// hooks can be stacked with each call
			return func() {
				// The content of the returned func will be deferred to run after the test cases.
			}
		})

		s.Then(`it will report whether Field1 is lower or not`, func(t *testing.T, v *testcase.V) {
			require.True(t, subject(`all lower case character`))
			require.False(t, subject(`Capitalized`))
		})
	})
}
```

## Steps based approach

Steps is an easier idiom, that allows you to work with your favorite testing idiom.
It builds on the foundation of variable scoping.
If you use it for setting up variables for your test cases,
you should be aware, that for that purpose, it is not safe to use on concurrent goroutines.

```go
func TestSomething(t *testing.T) {
    var value string

    var steps = testcase.Steps{}
    t.Run(`on`, func(t *testing.T) {
        steps := steps.Add(func(t *testing.T) { value = "1" })

        t.Run(`each`, func(t *testing.T) {
            steps := steps.Add(func(t *testing.T) { value = "2" })

            t.Run(`nested`, func(t *testing.T) {
                steps := steps.Add(func(t *testing.T) { value = "3" })

                t.Run(`layer`, func(t *testing.T) {
                    steps := steps.Add(func(t *testing.T) { value = "4" })

                    t.Run(`it will setup and break down the right context`, func(t *testing.T) {
                        steps.Setup(t)

                        require.Equal(t, "4", value)
                    })
                })

                t.Run(`then`, func(t *testing.T) {
                    steps.Setup(t)

                    require.Equal(t, "3", value)
                })
            })

            t.Run(`then`, func(t *testing.T) {
                steps.Setup(t)

                require.Equal(t, "2", value)
            })
        })

        t.Run(`then`, func(t *testing.T) {
            steps.Setup(t)

            require.Equal(t, "1", value)
        })
    })
}
```

## Reference Projects
* [FeatureFlags](https://github.com/adamluzsi/FeatureFlags)
  * root cause why I created this in the first place.
