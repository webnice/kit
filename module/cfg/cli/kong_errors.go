package cli

import "strings"

const (
	cUnknownError                                 = "unknown error: %s"
	cNoCommandSelected                            = "no command selected"
	cExpected                                     = "expected %s"
	cExpectedOneOf                                = "expected one of %s"
	cUnexpectedArgument                           = "unexpected argument %s"
	cUnexpectedFlagArgument                       = "unexpected flag argument %q"
	cUnexpectedToken                              = "unexpected token %s"
	cUnknownFlag                                  = "неизвестный флаг %q"
	cMissingFlags                                 = "не указан флаг %s"
	cHelperErrorDidYouMean                        = "%s, возможно, вы хотели указать %s?"
	cHelperErrorDidYouMeanOneOf                   = "%s, возможно, вы хотели указать один из %s?"
	cDefaultValueFor                              = "default value for %s: %s"
	cEnumValueFor                                 = "enum value for %s: %s"
	cEnvValueFor                                  = "env value for %s: %s"
	cHelpFor                                      = "help for %s: %s"
	cIsRequired                                   = "%s is required"
	cEnumSliceOrValue                             = "enum can only be applied to a slice or value"
	cMustBeOneOfButGot                            = "%s must be one of %s but got %q"
	cMissingPositionalArguments                   = "missing positional arguments %s"
	cFailField                                    = "%s.%s: %s"
	cRegexInputCannotBeEmpty                      = "regex input cannot be empty"
	cRegexUnableToCompile                         = "unable to compile regex: %s"
	cKongMustBeConfiguredWithConfiguration        = "kong must be configured with kong.Configuration"
	cUndefinedVariable                            = "undefined variable ${%s}"
	cExpectedPointerToStructButGot                = "expected a pointer to a struct but got %T"
	cExpectedAPointer                             = "expected a pointer"
	cCantMixPositionalArgumentsBranchingArguments = "can't mix positional arguments and branching arguments on %T"
	cDecodeBoolValue                              = "булево значение должно быть %s, %s, %s, но получено значение %q"
	cDecodeBoolType                               = "ожидался булев тип, но получен тип %q (%T)"
	cDecodeDurationValue                          = "ожидалось значение продолжительности времени, но получено значение %q: %s"
	cDecodeDurationType                           = "ожидался тип продолжительности времени, но получено %q"
	cDecodeIntValue                               = "ожидалось значение для %d бит, типа int, но получено значение %q"
	cDecodeIntType                                = "ожидался тип int, но получен тип %q (%T)"
	cDecodeUintValue                              = "ожидалось значение для %d бит, типа uint, но получено значение %q"
	cDecodeUintType                               = "ожидался тип uint, но получен тип %q (%T)"
	cDecodeFloatValue                             = "ожидалось значение типа float, но получено значение %q (%T)"
	cDecodeFloatType                              = "ожидался тип float, но получен типа %q (%T)"
	cDecodeValueError                             = "%s: %s"
	cDecodeValueEnv                               = "%s (из переменной окружения %s=%q)"
)

var (
	kongErrSingleton                                = &KongError{}
	errUnknownError                                 = err{tpl: cUnknownError}
	errNoCommandSelected                            = err{tpl: cNoCommandSelected}
	errExpected                                     = err{tpl: cExpected}
	errExpectedOneOf                                = err{tpl: cExpectedOneOf}
	errUnexpectedArgument                           = err{tpl: cUnexpectedArgument}
	errUnexpectedFlagArgument                       = err{tpl: cUnexpectedFlagArgument}
	errUnexpectedToken                              = err{tpl: cUnexpectedToken}
	errUnknownFlag                                  = err{tpl: cUnknownFlag}
	errMissingFlags                                 = err{tpl: cMissingFlags}
	errHelperErrorDidYouMean                        = err{tpl: cHelperErrorDidYouMean}
	errHelperErrorDidYouMeanOneOf                   = err{tpl: cHelperErrorDidYouMeanOneOf}
	errDefaultValueFor                              = err{tpl: cDefaultValueFor}
	errEnumValueFor                                 = err{tpl: cEnumValueFor}
	errEnvValueFor                                  = err{tpl: cEnvValueFor}
	errHelpFor                                      = err{tpl: cHelpFor}
	errIsRequired                                   = err{tpl: cIsRequired}
	errEnumSliceOrValue                             = err{tpl: cEnumSliceOrValue}
	errMustBeOneOfButGot                            = err{tpl: cMustBeOneOfButGot}
	errMissingPositionalArguments                   = err{tpl: cMissingPositionalArguments}
	errFailField                                    = err{tpl: cFailField}
	errRegexInputCannotBeEmpty                      = err{tpl: cRegexInputCannotBeEmpty}
	errRegexUnableToCompile                         = err{tpl: cRegexUnableToCompile}
	errKongMustBeConfiguredWithConfiguration        = err{tpl: cKongMustBeConfiguredWithConfiguration}
	errUndefinedVariable                            = err{tpl: cUndefinedVariable}
	errExpectedPointerToStructButGot                = err{tpl: cExpectedPointerToStructButGot}
	errExpectedAPointer                             = err{tpl: cExpectedAPointer}
	errCantMixPositionalArgumentsBranchingArguments = err{tpl: cCantMixPositionalArgumentsBranchingArguments}
	errDecodeBoolValue                              = err{tpl: cDecodeBoolValue}
	errDecodeBoolType                               = err{tpl: cDecodeBoolType}
	errDecodeDurationValue                          = err{tpl: cDecodeDurationValue}
	errDecodeDurationType                           = err{tpl: cDecodeDurationType}
	errDecodeIntValue                               = err{tpl: cDecodeIntValue}
	errDecodeIntType                                = err{tpl: cDecodeIntType}
	errDecodeUintValue                              = err{tpl: cDecodeUintValue}
	errDecodeUintType                               = err{tpl: cDecodeUintType}
	errDecodeFloatValue                             = err{tpl: cDecodeFloatValue}
	errDecodeFloatType                              = err{tpl: cDecodeFloatType}
	errDecodeValueError                             = err{tpl: cDecodeValueError}
	errDecodeValueEnv                               = err{tpl: cDecodeValueEnv}
)

// ERRORS: Implementation of errors with the ability to compare errors with each other

// UnknownError Unknown Error
func (e *KongError) UnknownError(err error) Err { return newErr(&errUnknownError, err) }

// NoCommandSelected no command selected
func (e *KongError) NoCommandSelected() Err { return newErr(&errNoCommandSelected) }

// Expected children ...
func (e *KongError) Expected(s string) Err { return newErr(&errExpected, s) }

// ExpectedOneOf Expected one of ...
func (e *KongError) ExpectedOneOf(s string) Err { return newErr(&errExpectedOneOf, s) }

// UnexpectedArgument Unexpected argument ...
func (e *KongError) UnexpectedArgument(argument string) Err {
	return newErr(&errUnexpectedArgument, argument)
}

// UnexpectedFlagArgument unexpected flag argument
func (e *KongError) UnexpectedFlagArgument(argument any) Err {
	return newErr(&errUnexpectedFlagArgument, argument)
}

// UnexpectedToken Unexpected token ...
func (e *KongError) UnexpectedToken(token Token) Err { return newErr(&errUnexpectedToken, token) }

// UnknownFlag unknown flag ...
func (e *KongError) UnknownFlag(flag string) Err { return newErr(&errUnknownFlag, flag) }

// MissingFlags missing flags: ...
func (e *KongError) MissingFlags(flag string) Err { return newErr(&errMissingFlags, flag) }

// HelperErrorDidYouMean ..., did you mean ...?
func (e *KongError) HelperErrorDidYouMean(err error, hypothesis string) Err {
	return newErr(&errHelperErrorDidYouMean, err, hypothesis)
}

// HelperErrorDidYouMeanOneOf ..., did you mean one of ...?
func (e *KongError) HelperErrorDidYouMeanOneOf(err error, hypothesis []string) Err {
	const hypothesisDelimiter = ", "
	return newErr(&errHelperErrorDidYouMeanOneOf, err, strings.Join(hypothesis, hypothesisDelimiter))
}

// DefaultValueFor default value for ...
func (e *KongError) DefaultValueFor(s string, err error) Err {
	return newErr(&errDefaultValueFor, s, err)
}

// EnumValueFor enum value for ...
func (e *KongError) EnumValueFor(s string, err error) Err { return newErr(&errEnumValueFor, s, err) }

// EnvValueFor env value for ...
func (e *KongError) EnvValueFor(s string, err error) Err { return newErr(&errEnvValueFor, s, err) }

// HelpFor help for ...
func (e *KongError) HelpFor(s string, err error) Err { return newErr(&errHelpFor, s, err) }

// IsRequired ... is required
func (e *KongError) IsRequired(required string) Err { return newErr(&errIsRequired, required) }

// EnumSliceOrValue Enum can only be applied to a slice or value
func (e *KongError) EnumSliceOrValue() Err { return newErr(&errEnumSliceOrValue) }

// MustBeOneOfButGot ... must be one of ... but got ...
func (e *KongError) MustBeOneOfButGot(summary string, enums string, item any) Err {
	return newErr(&errMustBeOneOfButGot, summary, enums, item)
}

// MissingPositionalArguments missing positional arguments ...
func (e *KongError) MissingPositionalArguments(arguments string) Err {
	return newErr(&errMissingPositionalArguments, arguments)
}

// FailField Error for check field
func (e *KongError) FailField(parent, name, value string) Err {
	return newErr(&errFailField, parent, name, value)
}

// RegexInputCannotBeEmpty Regex input cannot be empty
func (e *KongError) RegexInputCannotBeEmpty() Err { return newErr(&errRegexInputCannotBeEmpty) }

// RegexUnableToCompile Unable to compile regex: ...
func (e *KongError) RegexUnableToCompile(err error) Err { return newErr(&errRegexUnableToCompile, err) }

// KongMustBeConfiguredWithConfiguration Kong must be configured with kong.Configuration
func (e *KongError) KongMustBeConfiguredWithConfiguration() Err {
	return newErr(&errKongMustBeConfiguredWithConfiguration)
}

// UndefinedVariable undefined variable ...
func (e *KongError) UndefinedVariable(name string) Err { return newErr(&errUndefinedVariable, name) }

// ExpectedPointerToStructButGot expected a pointer to a struct but got ...
func (e *KongError) ExpectedPointerToStructButGot(ast any) Err {
	return newErr(&errExpectedPointerToStructButGot, ast)
}

// ExpectedAPointer Expected a pointer
func (e *KongError) ExpectedAPointer() Err { return newErr(&errExpectedAPointer) }

// CantMixPositionalArgumentsBranchingArguments Can't mix positional arguments and branching arguments on ...
func (e *KongError) CantMixPositionalArgumentsBranchingArguments(ast any) Err {
	return newErr(&errCantMixPositionalArgumentsBranchingArguments, ast)
}

// DecodeBoolValue Булево значение должно быть ..., ..., ..., но получено значение ...
func (e *KongError) DecodeBoolValue(v1, v2, v3 string, value string) Err {
	return newErr(&errDecodeBoolValue, v1, v2, v3, value)
}

// DecodeBoolType Ожидался булев тип, но получен тип ... (...)
func (e *KongError) DecodeBoolType(t any) Err { return newErr(&errDecodeBoolType, t, t) }

// DecodeDurationValue Ожидалось значение продолжительности времени, но получено значение ...: ...
func (e *KongError) DecodeDurationValue(value string, err error) Err {
	return newErr(&errDecodeDurationValue, value, err)
}

// DecodeDurationType Ожидался тип продолжительности времени, но получено ...
func (e *KongError) DecodeDurationType(value any) Err {
	return newErr(&errDecodeDurationType, value)
}

// DecodeIntValue Ожидалось значение для ... бит, типа int, но получено значение ...
func (e *KongError) DecodeIntValue(num int, tpe string) Err {
	return newErr(&errDecodeIntValue, num, tpe)
}

// DecodeIntType Ожидался тип int, но получен тип ... (...)
func (e *KongError) DecodeIntType(t string, v any) Err { return newErr(&errDecodeIntType, t, v) }

// DecodeUintValue Ожидалось значение для ... бит, типа uint, но получено значение ...
func (e *KongError) DecodeUintValue(num int, tpe string) Err {
	return newErr(&errDecodeUintValue, num, tpe)
}

// DecodeUintType Ожидался тип uint, но получен тип ... (...)
func (e *KongError) DecodeUintType(tpe string, v any) Err {
	return newErr(&errDecodeUintType, tpe, v)
}

// DecodeFloatValue Ожидалось значение типа float, но получено значение ... (...)
func (e *KongError) DecodeFloatValue(val string, tpe any) Err {
	return newErr(&errDecodeFloatValue, val, tpe)
}

// DecodeFloatType Ожидался тип float, но получен типа ... (...)
func (e *KongError) DecodeFloatType(val string, tpe any) Err {
	return newErr(&errDecodeFloatType, val, tpe)
}

// DecodeValueError Ошибка разбора аргумента, флага или параметра: ...: ...
func (e *KongError) DecodeValueError(short string, err error) Err {
	return newErr(&errDecodeValueError, short, err)
}

// DecodeValueEnv Ошибка разбора аргумента, флага или параметра из переменной окружения
func (e *KongError) DecodeValueEnv(err error, variable string, value string) Err {
	return newErr(&errDecodeValueEnv, err, variable, value)
}
