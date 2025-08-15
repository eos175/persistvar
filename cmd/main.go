package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eos175/persistvar"
	"github.com/eos175/persistvar/storage"
)

func main() {
	fs, _ := storage.NewBoltStorage("vars.db")
	mgr := persistvar.NewVarManager(fs)
	defer mgr.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mgr.AutoSync(ctx, 3*time.Minute)

	counter, _ := persistvar.NewVar(mgr, "counter", 0)
	name, _ := persistvar.NewVar(mgr, "username", "anon")

	counter.SetLazy(counter.Get() + 1)
	name.SetLazy("Emmanuel")

	fmt.Println("Counter:", counter.Get())
	fmt.Println("Name:", name.Get())
}
