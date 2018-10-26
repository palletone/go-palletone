package storage

import "github.com/palletone/go-palletone/common"

//<<<<<<< Utils <<<<<<<
//YiRan
//Returns true when the contents of the two []byte are exactly the same
func BytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) { //[]int{} != []int(nil)
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

//YiRan
//Returns true when the contents of the two Address are exactly the same
func AddressEqual(a, b common.Address) bool {
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

//YiRan
// this function connect multiple []byte to single []byte.
func KeyConnector(keys ...[]byte) []byte {
	var res []byte
	for _, key := range keys {
		res = append(res, key...)
	}
	return res
}

//YiRan
//print error if exist.
func ErrorLogHandler(err error, errType string) error {
	if err != nil {
		println(errType, "error", err.Error())
		return err
	}
	return nil
}

//>>>>>>> Utils >>>>>>>
