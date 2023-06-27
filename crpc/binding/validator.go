package binding

import (
	"encoding/json"
	"errors"
	"fmt"
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
		if len(sliceValidateError) == 0 {
			return nil
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

func validate(model any) error {
	return Validator.ValidateStruct(model)
}

// 结构体参数完整性校验
// 反射
func validateParam(model any, decoder *json.Decoder) error {
	m := reflect.ValueOf(model)
	if m.Kind() != reflect.Pointer {
		return errors.New("this model is not a pointer")
	}
	elem := m.Elem().Interface()
	of := reflect.ValueOf(elem)
	switch of.Kind() {
	case reflect.Struct:
		return checkParam(of, model, decoder)
	case reflect.Slice, reflect.Array:
		ele := of.Type().Elem()
		fmt.Println(ele.Kind(), reflect.Struct)
		if ele.Kind() == reflect.Struct {
			return checkParamSlice(ele, model, decoder)
		} else if ele.Kind() == reflect.Pointer {
			fmt.Println(reflect.ValueOf(ele.Elem()), ele.Elem())
			return checkParam(reflect.ValueOf(ele.Elem()), model, decoder)
		}
	default:
		err := decoder.Decode(model)
		if err != nil {
			return err
		}
	}
	return nil
}
