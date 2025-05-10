package struct_method_check_fix

import (
	"time"
)

func NewOtherWayMyStruct() MyStruct {
	return MyStruct{Name: "John"}
}

func NewTimeStruct() time.Time {
	return time.Now()
}
