package dig

import "github.com/RelicOfTesla/dig/dig2"

type UBerProviderMgr interface {
	Invoke(f any) error
	Provide(f any, _opts ...dig2.ProvideOption) error
}

func New(opts ...dig2.Option) UBerProviderMgr {
	return dig2.New(opts...)
}

type In = dig2.In
type Out = dig2.Out

var Name = dig2.Name
var Group = dig2.Group

var Cache = dig2.Cache
