# go-error-buffer

A tiny library that makes it easy to make sure you don't have too many errors happening too fast.

For example, say you have a long running process talking to a DB. A few errors here and there could be normal, e.g. if your network is not always reliable, so you just want to keep re-trying when errors happen; but not if they start happening too fast.

## Usage

Specify how many errors is too much in how much time when you create your `ErrorBuffer`:

```go
// import  "github.com/wk8/go-error-buffer"

buffer := errorbuffer.NewErrorBuffer(3, time.Minute)
```

Then add errors as they come in. If it's too much too fast, then the buffer will overflow:
```go
for {
	err := myFunc()

	if err := buffer.Add(err); err != nil {
		// there were more than 3 errors in the last minute
	}
}
```
