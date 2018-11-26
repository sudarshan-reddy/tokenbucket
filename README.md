# tokenbucket
Token bucket algorithm implementation in Go

## Sample usage
```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sudarshan-reddy/tokenbucket"
)

func main() {
	w1 := tokenbucket.NewWriter(os.Stdout, tokenbucket.NewBucket(5*time.Millisecond, 10))
	w2 := tokenbucket.NewWriter(os.Stdout, tokenbucket.NewBucket(10*time.Millisecond, 10))

	go func() {
		var i int
		for {
			i++
			w2.Write([]byte(fmt.Sprintf("w2: %d\n", i)))
		}
	}()

	var j int
	for {
		j++
		w1.Write([]byte(fmt.Sprintf("w1: %d\n", j)))
	}

}

```
