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
	tb1 := tokenbucket.NewBucket(2*time.Second, 10)
	tb2 := tokenbucket.NewBucket(1*time.Second, 10)
	w1 := tokenbucket.NewWriter(os.Stdout, tb1)
	w2 := tokenbucket.NewWriter(os.Stdout, tb2)

	go func() {
		var i int
		for {
			i++
			if i == 3 {
				//dynamically change speed midway
				tb2.ChangeInterval(100 * time.Millisecond)
			}
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
