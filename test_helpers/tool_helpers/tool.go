package tool_helpers

import (
	"fmt"
	"time"
)

func Retry(attempts int, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return nil
		}
		if i >= (attempts - 1) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
