package binding

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
)

type jsonBinding struct {
	disallowUnknownFields bool
	isValidate            bool
}

func (j *jsonBinding) Name() string {
	return "json"
}

func (j *jsonBinding) Bind(request *http.Request, model any) error {
	body := request.Body
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	if j.disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if j.isValidate {
		err := validateParam(model, decoder)
		if err != nil {
			return err
		}
	}
	err := decoder.Decode(model)
	if err != nil && err != io.EOF {
		return err
	}
	return validate(model)
}

// 验证数组、切片
func checkParamSlice(of reflect.Type, model any, decoder *json.Decoder) error {
	mv := make([]map[string]interface{}, 0)
	err := decoder.Decode(&mv)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		jsonTag := of.Field(i).Tag.Get("json")
		if jsonTag == "" {
			jsonTag = of.Field(i).Name
		}
		for _, v := range mv {
			if _, ok := v[jsonTag]; !ok && of.Field(i).Tag.Get("crpc") == "require" {
				return errors.New("miss field " + jsonTag)
			}
		}
	}
	marshal, err := json.Marshal(mv)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, model)
	if err != nil {
		return err
	}
	return nil
}

// 验证普通结构体
func checkParam(of reflect.Value, model any, decoder *json.Decoder) error {
	mv := make(map[string]any)
	err := decoder.Decode(&mv)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		jsonTag := of.Type().Field(i).Tag.Get("json")
		if jsonTag == "" {
			jsonTag = of.Type().Field(i).Name
		}
		if _, ok := mv[jsonTag]; !ok && of.Type().Field(i).Tag.Get("crpc") == "require" {
			return errors.New("miss field " + jsonTag)
		}
	}
	if err != nil && err != io.EOF {
		return err
	}
	marshal, err := json.Marshal(mv)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, model)
	if err != nil {
		return err
	}
	return nil
}
