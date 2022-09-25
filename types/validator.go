package types

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	NameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	EnvRegex  = regexp.MustCompile("^[a-zA-Z_]+[a-zA-Z0-9_]*$")
	UrlRegex  = regexp.MustCompile(`^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$`)
)

var converters = map[string]func(string) (interface{}, error){}

func RegisterConverter[T any](zero T, converter func(string) (T, error)) {
	converters[fmt.Sprintf("%T", zero)] = func(s string) (interface{}, error) {
		return converter(s)
	}
}

func init() {
	// string
	RegisterConverter("", func(s string) (string, error) {
		return s, nil
	})
	// int
	RegisterConverter(int(0), func(s string) (int, error) {
		return strconv.Atoi(s)
	})
	// int32
	RegisterConverter(int32(0), func(s string) (int32, error) {
		i, err := strconv.ParseInt(s, 10, 32)
		return int32(i), err
	})
	// int64
	RegisterConverter(int64(0), func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
	// uint
	RegisterConverter(uint(0), func(s string) (uint, error) {
		i, err := strconv.ParseUint(s, 10, 64)
		return uint(i), err
	})
	// uint32
	RegisterConverter(uint32(0), func(s string) (uint32, error) {
		i, err := strconv.ParseUint(s, 10, 32)
		return uint32(i), err
	})
	// uint64
	RegisterConverter(uint64(0), func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, 64)
	})
	// float32
	RegisterConverter(float32(0), func(s string) (float32, error) {
		f, err := strconv.ParseFloat(s, 64)
		return float32(f), err
	})
	// float64
	RegisterConverter(float64(0), func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
	// bool
	RegisterConverter(false, func(s string) (bool, error) {
		return strconv.ParseBool(s)
	})
	// *regex.Regexp
	RegisterConverter(UrlRegex, func(s string) (*regexp.Regexp, error) {
		return regexp.Compile(s)
	})
}

func fuzzyEqual[T any](item T) T {
	if str, ok := any(item).(string); ok {
		return any(strings.ToLower(strings.TrimSpace(str))).(T)
	}

	return item
}

type Validator[T any] interface {
	Validate(T) error
	Convert(string) (T, error)
}

type ValidatorFunction[T any] func(T) error

func MultiValidator[T any](validators ...Validator[T]) Validator[T] {
	return ValidatorFunction[T](func(item T) error {
		for _, validator := range validators {
			if err := validator.Validate(item); err != nil {
				if errors.Is(err, ErrValidtorStopValidation) {
					return nil
				}

				return err
			}
		}

		return nil
	})
}

func NotEqualValidator[T comparable](template Stringer, future Future[[]T]) Validator[T] {
	return ValidatorFunction[T](func(val T) error {
		val = fuzzyEqual(val)

		items, err := future.Get()
		if err != nil {
			return err
		}

		for _, item := range items {
			if val == fuzzyEqual(item) {
				return fmt.Errorf(template.String(), val)
			}
		}

		return nil
	})
}

func EqualValidator[T comparable](format Stringer, future Future[[]T]) Validator[T] {
	return ValidatorFunction[T](func(val T) error {
		val = fuzzyEqual(val)

		items, err := future.Get()
		if err != nil {
			return err
		}

		for _, item := range items {
			if val == fuzzyEqual(item) {
				return nil
			}
		}

		return fmt.Errorf(format.String(), val)
	})
}

func NameValidator(name string, allowEmpty bool) Validator[string] {
	if name == "" {
		name = "name"
	}

	return ValidatorFunction[string](func(s string) error {
		if allowEmpty && s == "" {
			return nil
		}

		if s == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}

		if NameRegex.MatchString(s) {
			return nil
		}

		return fmt.Errorf("%s %s", name, ErrInvalidName.Error())
	})
}

func PathValidator(name string, allowEmpty bool, allowDir bool, allowFile bool, allowStdIn bool) Validator[string] {
	if name == "" {
		name = "path"
	}

	return ValidatorFunction[string](func(s string) error {
		if allowEmpty && s == "" {
			return nil
		}

		if s == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}

		if s == "-" {
			if allowStdIn {
				return nil
			} else {
				return fmt.Errorf("%s cannot be stdin", name)
			}
		}

		var retErr error
		if s, err := os.Stat(s); err != nil {
			if os.IsNotExist(err) {
				retErr = ErrInvalidPathNotFound
			} else {
				retErr = err
			}
		} else if s.IsDir() && !allowDir {
			retErr = ErrInvalidPathIsDir
		} else if !s.IsDir() && !allowFile {
			retErr = ErrInvalidPathIsFile
		}

		if retErr == nil {
			return nil
		}

		return fmt.Errorf("%s %s", name, retErr.Error())
	})
}

func UrlValidator(name string, allowEmpty bool) Validator[string] {
	if name == "" {
		name = "url"
	}

	return ValidatorFunction[string](func(s string) error {
		if allowEmpty && s == "" {
			return nil
		}

		if s == "" {
			return fmt.Errorf("%s cannot be empty", name)
		}

		if UrlRegex.MatchString(s) {
			return nil
		}

		return fmt.Errorf("%s %s", name, ErrInvalidURL.Error())
	})
}

func EmptyValidator[T comparable](name string, empty bool) Validator[T] {
	if name == "" {
		name = "value"
	}

	return ValidatorFunction[T](func(s T) error {
		var a T

		if empty {
			if s != a {
				return fmt.Errorf("%s must be empty", name)
			}
		} else {
			if s == a {
				return fmt.Errorf("%s cannot be empty", name)
			}
		}

		return nil
	})
}

func OptionalEmptyValidator[T comparable]() Validator[T] {
	return ValidatorFunction[T](func(s T) error {
		var a T

		if s == a {
			return ErrValidtorStopValidation
		}

		return nil
	})
}

func ConditionalValidator[T comparable](validator func(input T) Validator[T]) Validator[T] {
	return ValidatorFunction[T](func(val T) error {
		validator := validator(val)
		if validator == nil {
			return nil
		}

		return validator.Validate(val)
	})
}

func (v ValidatorFunction[T]) Validate(val T) error {
	return v(val)
}

func (v ValidatorFunction[T]) Convert(val string) (T, error) {
	var zero T

	converter, ok := converters[fmt.Sprintf("%T", zero)]
	if !ok {
		return zero, fmt.Errorf("no converter for %T", zero)
	}

	ret, err := converter(val)
	if err != nil {
		return zero, err
	}

	return ret.(T), nil
}

func EnvValidator() Validator[string] {
	return ValidatorFunction[string](func(s string) error {
		if s == "" {
			return errors.New("name cannot be empty")
		}

		if EnvRegex.MatchString(s) {
			return nil
		}

		return errors.New("name must be a valid env variable name")
	})
}
