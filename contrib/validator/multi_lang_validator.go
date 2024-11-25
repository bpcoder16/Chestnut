package validator

import (
	"errors"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"strings"
	"sync"
)

type MultiLangValidator struct {
	Locale   string
	TagName  string
	trans    ut.Translator
	once     sync.Once
	validate *validator.Validate
}

var _ binding.StructValidator = &MultiLangValidator{}

func (v *MultiLangValidator) ValidateStruct(obj interface{}) error {
	value := reflect.ValueOf(obj)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}

	if valueType == reflect.Struct {
		v.lazyInit()
		if err := v.validate.Struct(obj); err != nil {
			errList := err.(validator.ValidationErrors)
			var sliceErrList []string
			for _, e := range errList {
				sliceErrList = append(sliceErrList, e.Translate(v.trans))
			}
			return errors.New(strings.Join(sliceErrList, ";"))
		}
	}
	return nil
}

func (v *MultiLangValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

func (v *MultiLangValidator) lazyInit() {
	v.once.Do(func() {
		v.validate = validator.New()
		if len(v.TagName) > 0 {
			v.validate.SetTagName(v.TagName)
		}

		utp := ut.New(en.New(), zh.New())
		switch v.Locale {
		case "zh":
			v.trans, _ = utp.GetTranslator("zh")
			_ = zhTranslations.RegisterDefaultTranslations(v.validate, v.trans)
		default:
			v.trans, _ = utp.GetTranslator("en")
			_ = enTranslations.RegisterDefaultTranslations(v.validate, v.trans)
		}
	})
}
