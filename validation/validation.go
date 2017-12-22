package validation

import (
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

	var verr validationError

	for i := range errs {
		switch t := errs[i].(type) {
		case govalidator.Errors:
			if len(t) == 0 {
				continue
			}
			e, ok := t[0].(govalidator.Error)
			if !ok {
				continue
			}

			verr = AddError(verr, e.Name, e.Err).(validationError)
		case govalidator.Error:
			verr = AddError(verr, t.Name, t.Err).(validationError)
		}
	}

	return verr
}
