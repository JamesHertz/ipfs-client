package utils 

import (
	"os"
	"fmt"
)

func Eprintf(format string, args ...any){
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}