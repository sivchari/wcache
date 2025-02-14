package wcache_test

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/sivchari/wcache"
)

func ExampleWeakCache() {
	c := wcache.New[string, string]().
		WithLogger(slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
						if a.Key == "time" {
							return slog.String("time", "for-test")
						}
						return a
					},
				},
			),
		))

	value1 := "value1"
	c.Set("key1", &value1)
	value2 := "value2"
	c.Set("key-2", &value2)

	got1 := c.Get("key1")
	got2 := c.Get("key-2")

	fmt.Println("got1:", *got1)
	fmt.Println("got2:", *got2)

	value1 = ""
	value2 = ""

	runtime.GC()
	time.Sleep(1 * time.Second)

	newGot1 := c.Get("key1")
	newGot2 := c.Get("key-2")
	fmt.Println("newGot1:", newGot1)
	fmt.Println("newGot2:", newGot2)

	// Output:
	// got1: value1
	// got2: value2
	// {"time":"for-test","level":"INFO","msg":"deleting key","key":"key-2"}
	// {"time":"for-test","level":"INFO","msg":"deleting key","key":"key1"}
	// newGot1: <nil>
	// newGot2: <nil>
}

func TestWeakCache(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	c := wcache.New[string, User]()
	alice := &User{
		Name: "Alice",
		Age:  20,
	}
	c.Set("alice", alice)

	if got := c.Get("alice"); got != alice {
		t.Errorf("got %v, want %v", got, &alice)
	}
	t.Logf("alice: %p", alice)
	runtime.GC()

	// alice is still referenced.
	if got := c.Get("alice"); got != alice {
		t.Errorf("got %v, want %v", got, &alice)
	}
	t.Logf("alice: %p", alice)
	alice = nil
	runtime.GC()

	// alice is not referenced anymore.
	if got := c.Get("alice"); got != nil {
		t.Errorf("got %v, want %v", got, nil)
	} else {
		t.Logf("alice: %v", got)
	}
}
