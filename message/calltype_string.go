// Code generated by "stringer -type=CallType"; DO NOT EDIT

package message

import "fmt"

const _CallType_name = "CALLCALLALLTRIGGERTRIGGERALL"

var _CallType_index = [...]uint8{0, 4, 11, 18, 28}

func (i CallType) String() string {
	if i < 0 || i >= CallType(len(_CallType_index)-1) {
		return fmt.Sprintf("CallType(%d)", i)
	}
	return _CallType_name[_CallType_index[i]:_CallType_index[i+1]]
}