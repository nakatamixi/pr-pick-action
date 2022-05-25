package main

import (
	"fmt"

	"golang.org/x/xerrors"
)

func main() {
	fmt.Println(xerrors.New("test"))
}
