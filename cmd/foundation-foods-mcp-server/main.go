package main

import (
	"fmt"
	"os"

	"github.com/noot-app/foundation-foods-mcp-server/internal/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
