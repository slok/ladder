package types

import "fmt"

//I2Int64 will take an int interface and return an int64 value
func I2Int64(n interface{}) int64 {
	switch n := n.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return int64(n)
	}
	panic(fmt.Sprintf("%v is not a valid int type", n))
}
