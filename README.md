# Well-defined rollouts

Imagine your service rollout just being:

```go
err := rollout.ServiceRollout(myService)
if err != nil {
	svcErr, ok := errors.As(err, rollout.ServiceErr)
	if !ok {
	    fmt.Printf("unexpected failed deployment: %s\n", err)
	}
	
	// Do any detailed handling depending on which stage failed.
	fmt.Printf("service rollout failed: %s\n", err)
}

fmt.Printf("service rollout succeeded!")
```

This library gives you an interface for a `Service` that can be
rolled out in a safe manner.

It works by defining a `Service` that has three methods:

- `Rollout() error`
- `CheckHealth() error`
- `Rollback() error`

With that we can create a safe way of rolling out that works like:

- Attempt `Rollout`
- If successful `CheckHealth`
- If successful done

Or invoke `Rollback` whenever a step failed. This of course mean that
the `Rollback` method should be as safe as possible.

# Detecting errors

The rollout package comes with a special error `ServiceErr` that
wraps the errors from the three methods of a service.

The error message with vary depending on which part failed.
If you're throwing typed errors or well identifiable ones in your
methods, you can use `errors.Is` to check for them.