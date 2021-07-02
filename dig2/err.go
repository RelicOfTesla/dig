package dig2

import (
	stdErr "errors"
)

var ErrNotFoundTargetProvider = stdErr.New("not found target provider")
var ErrNotFoundCreatorParser = stdErr.New("not found creator function parser")
var ErrLoopRequire = stdErr.New("loop require")
var ErrMustFuncType = stdErr.New("must func type")
var ErrMustFuncResult = stdErr.New("invalid factory creator, must return a value")
var ErrDuplicateProvider = stdErr.New("duplicate provider")
var ErrInvalidOutIndex = stdErr.New("invalid out index")
var ErrUBerDefine = stdErr.New("uber code define")
var ErrUnexportedField = stdErr.New("unexported fields not allowed")
var ErrCanNotSetFieldValue = stdErr.New("can not set field value")
var ErrNotInitHoldValue = stdErr.New("not init hold value")
var ErrNotDefineHoldType = stdErr.New("not define this hold type")

//var ErrInvalidHoldValueType = stdErr.New("invalid hold value type")
