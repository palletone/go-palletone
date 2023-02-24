package adaptor

import (
	"errors"
)

var ErrNotFound = errors.New("not found")         //未找到对应的结果
var ErrApiException = errors.New("api exception") //API端发生异常
