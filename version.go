package main

import (
	"fmt"
	"runtime"
)

const version = "1.0.0"

func printVersion() {
	fmt.Printf("free-ddns version: %s (%s)\n", version, runtime.Version())
}
