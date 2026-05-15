package test

import (
	"OnlinePrictice/Helper"
	"fmt"
	"testing"
)

func TestUUID(t *testing.T) {
	s := Helper.GetUUID()
	fmt.Println(s, len(s))
}
