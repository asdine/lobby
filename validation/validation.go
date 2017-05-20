package validation

import (
	"reflect"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Validate validates and saves all the govalidator errors in a ValidatorError.
func Validate(s interface{}) error {
	ok, err := govalidator.ValidateStruct(s)
	if ok {
		return nil
	}

	errs, ok := err.(govalidator.Errors)
	if !ok || errs == nil {
		return nil
	}

	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var verr validationError

	for i := range errs {
		e, ok := errs[i].(govalidator.Error)
		if !ok {
			continue
		}

		f, ok := typ.FieldByName(e.Name)
		if !ok {
			// shouldn't happen.
			panic("unknown field")
		}

		var name = e.Name

		tag := f.Tag.Get("json")
		if tag != "" {
			if idx := strings.Index(tag, ","); idx != -1 {
				name = tag[:idx]
			} else {
				name = tag
			}
		}

		verr = AddError(verr, name, e.Err).(validationError)
	}

	return verr
}
