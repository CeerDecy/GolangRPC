package binding

import (
	"github.com/go-playground/validator/v10"
	crpcError "github/CeerDecy/RpcFrameWork/crpc/error"
	"reflect"
	"sync"
)

type StructValidator interface {
	ValidateStruct(any) error
	Engine() any
}

var Validator StructValidator = &defaultValidator{}

type defaultValidator struct {
	validate *validator.Validate
	once     sync.Once
}

func (d *defaultValidator) ValidateStruct(model any) error {
	of := reflect.ValueOf(model)
	switch of.Kind() {
	case reflect.Pointer:
		return d.ValidateStruct(of.Elem().Interface())
	case reflect.Struct:
		return d.validateStruct(model)
	case reflect.Slice, reflect.Array:
		count := of.Len()
		sliceValidateError := make(crpcError.SliceValidateError, 0)
		for i := 0; i < count; i++ {
			if err := d.validateStruct(of.Index(i).Interface()); err != nil {
				sliceValidateError = append(sliceValidateError, err)
			}
		}
		return sliceValidateError
	}
	return nil
}

// Engine 单例模式创建引擎
func (d *defaultValidator) Engine() any {
	d.lazyInit()
	return d.validate
}

// 验证结构体
func (d *defaultValidator) validateStruct(model any) error {
	d.lazyInit()
	return d.validate.Struct(model)
}

func (d *defaultValidator) lazyInit() {
	d.once.Do(func() {
		d.validate = validator.New()
	})
}
