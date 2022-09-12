package argparse

import (
	"fmt"
	"strconv"
)

type Options[T comparable] struct {
	Required  bool
	Help      string
	Validator func(T) error
	Default   T
}

type OptionsList[T comparable] struct {
	Required  bool
	Help      string
	Validator func(T) error
	Default   []T
}

func (o *Options[T]) toCaster() *optionsCaster {
	c := &optionsCaster{
		Required: o.Required,
		Help:     o.Help,
	}

	var empty T
	if o.Default != empty {
		c.Default = o.Default
	}
	if o.Validator != nil {
		c.validator = o.Validator
	}
	c.init(empty)

	return c
}

func (o *OptionsList[T]) toCaster() *optionsCaster {
	c := &optionsCaster{
		Required: o.Required,
		Help:     o.Help,
		Default:  o.Default,
		isList:   true,
	}
	var empty T

	if len(o.Default) != 0 {
		c.Default = o.Default
	}
	if o.Validator != nil {
		c.validator = o.Validator
	}

	c.init(empty)

	return c
}

type argType int

const (
	argTypeBool argType = iota
	argTypeInt
	argTypeString
	argTypeFloat
)

type optionsCaster struct {
	Default  any
	Help     string
	Required bool
	ArgType  argType

	validator any
	isList    bool
}

func (o *optionsCaster) validate(value any) error {
	if o.validator == nil {
		return nil
	}

	switch o.ArgType {
	case argTypeBool:
		return o.validator.(func(bool) error)(value.(bool))
	case argTypeInt:
		return o.validator.(func(int) error)(value.(int))
	case argTypeString:
		return o.validator.(func(string) error)(value.(string))
	case argTypeFloat:
		return o.validator.(func(float64) error)(value.(float64))
	}

	return fmt.Errorf("unknown type: %d", o.ArgType)
}

func (o *optionsCaster) valueEmpty(value any) bool {
	switch o.ArgType {
	case argTypeBool:
		return value.(bool) == false
	case argTypeInt:
		return value.(int) == 0
	case argTypeString:
		return value.(string) == ""
	case argTypeFloat:
		return value.(float64) == 0
	default:
		panic(fmt.Errorf("unknown type: %d", o.ArgType))
	}
}

func (o *optionsCaster) cast(value string, ptr any) error {
	switch o.ArgType {
	case argTypeBool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		if o.isList {
			*ptr.(*[]bool) = append(*ptr.(*[]bool), b)
		} else {
			*ptr.(*bool) = b
		}
	case argTypeInt:
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		if o.isList {
			*ptr.(*[]int) = append(*ptr.(*[]int), i)
		} else {
			*ptr.(*int) = i
		}
	case argTypeString:
		if o.isList {
			*ptr.(*[]string) = append(*ptr.(*[]string), value)
		} else {
			*ptr.(*string) = value
		}
	case argTypeFloat:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		if o.isList {
			*ptr.(*[]float64) = append(*ptr.(*[]float64), f)
		} else {
			*ptr.(*float64) = f
		}
	default:
		panic(fmt.Errorf("unknown type: %d", o.ArgType))
	}

	return nil
}

func (o *optionsCaster) init(v any) {
	switch v.(type) {
	case bool:
		o.ArgType = argTypeBool
	case int:
		o.ArgType = argTypeInt
	case string:
		o.ArgType = argTypeString
	case float64:
		o.ArgType = argTypeFloat
	default:
		panic(fmt.Sprintf("unknown type %T", o.Default))
	}
}

func (o *optionsCaster) strType() string {
	var t string
	switch o.ArgType {
	case argTypeBool:
		t = "bool"
	case argTypeInt:
		t = "int"
	case argTypeString:
		t = "string"
	case argTypeFloat:
		t = "float"
	default:
		panic(fmt.Sprintf("unknown type: %d", o.ArgType))
	}

	if o.isList {
		t += "s"
	}

	return t
}

func (o *optionsCaster) castValue(value any, ptr any) {
	switch o.ArgType {
	case argTypeBool:
		if o.isList {
			*ptr.(*[]bool) = value.([]bool)
		} else {
			*ptr.(*bool) = value.(bool)
		}
	case argTypeInt:
		if o.isList {
			*ptr.(*[]int) = value.([]int)
		} else {
			*ptr.(*int) = value.(int)
		}
	case argTypeString:
		if o.isList {
			*ptr.(*[]string) = value.([]string)
		} else {
			*ptr.(*string) = value.(string)
		}
	case argTypeFloat:
		if o.isList {
			*ptr.(*[]float64) = value.([]float64)
		} else {
			*ptr.(*float64) = value.(float64)
		}
	default:
		panic(fmt.Errorf("unknown type: %d", o.ArgType))
	}
}
