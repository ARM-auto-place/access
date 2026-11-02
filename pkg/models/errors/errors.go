package models

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/valyala/fasthttp"
)

const (
	CodeInvalidParams = -32602
	CodeInternalError = -32603
	CodeRoleNotFound  = 1001
	CodeUserNotFound  = 1002
	publicTextPrefix  = "access.errors"
)

var defaultFactory = &Factory{}

type Factory struct{ serviceName, subsystem string }

func NewFactory(serviceName, subsystem string) *Factory {
	return &Factory{serviceName: serviceName, subsystem: subsystem}
}
func SetServiceInfo(serviceName, subsystem string) {
	defaultFactory = NewFactory(serviceName, subsystem)
}

func (f *Factory) New(msg string, statusCode int) *Error {
	return &Error{ErrorText: msg, statusCode: statusCode, rpcCode: CodeInternalError, publicText: buildPublicText(f.serviceName, f.subsystem, msg)}
}

func (f *Factory) Wrap(err error, publicReason string) *Error {
	if err == nil {
		return nil
	}
	result := f.New("internal server error", fasthttp.StatusInternalServerError)
	var existing *Error
	if errors.As(err, &existing) {
		result.ErrorText = existing.ErrorText
		result.statusCode = existing.statusCode
		result.rpcCode = existing.rpcCode
		result.publicText = existing.publicText
		for reason, detail := range existing.Cause {
			result.addCause(reason, detail)
		}
	}
	return result.WithWrapped(err, publicReason)
}

type Error struct {
	ErrorText           string         `json:"-"`
	Cause               map[string]any `json:"-"`
	publicText          string
	wrap                error
	statusCode, rpcCode int
}

func New(msg string, statusCode int) *Error {
	return defaultFactory.New(msg, statusCode)
}

func Wrap(err error, publicReason string) *Error {
	return defaultFactory.Wrap(err, publicReason)
}

func (e *Error) Error() string {
	if len(e.Cause) > 0 {
		return fmt.Sprintf("%s. Causes: %v", e.ErrorText, e.Cause)
	}
	return e.ErrorText
}
func (e *Error) Unwrap() error                 { return e.wrap }
func (e *Error) Code() int                     { return e.rpcCode }
func (e *Error) WithCode(code int) *Error      { e.rpcCode = code; return e }
func (e *Error) SetStatusCode(code int) *Error { e.statusCode = code; return e }
func (e *Error) GetStatusCode() int            { return e.statusCode }
func (e *Error) GetPublicText() string         { return e.publicText }
func (e *Error) GetPublicMessage() string      { return e.ErrorText }
func (e *Error) WithWrapped(err error, publicReason string) *Error {
	if err != nil {
		e.wrap = err
		e.addCause(publicReason, err.Error())
	}
	return e
}
func (e *Error) addCause(reason string, detail any) *Error {
	if e.Cause == nil {
		e.Cause = make(map[string]any)
	}
	e.Cause[reason] = detail
	return e
}
func HasCause(err error, cause string) bool {
	var e *Error
	if !errors.As(err, &e) {
		return false
	}
	for reason, detail := range e.Cause {
		if strings.Contains(reason, cause) || strings.Contains(fmt.Sprint(detail), cause) {
			return true
		}
	}
	return false
}
func Is(one, two error) bool {
	if one == nil || two == nil {
		return one == two
	}
	var a, b *Error
	if errors.As(one, &a) && errors.As(two, &b) {
		return a.Code() == b.Code() && a.ErrorText == b.ErrorText
	}
	return errors.Is(one, two)
}
func InvalidParams() *Error {
	return New("invalid parameters", fasthttp.StatusBadRequest).WithCode(CodeInvalidParams)
}
func InternalServerError() *Error {
	return New("internal server error", fasthttp.StatusInternalServerError).WithCode(CodeInternalError)
}
func ErrRoleNotFound() *Error {
	return New("role not assigned", fasthttp.StatusNotFound).WithCode(CodeRoleNotFound)
}
func ErrUserNotFound() *Error {
	return New("user not found", fasthttp.StatusNotFound).WithCode(CodeUserNotFound)
}
func buildPublicText(serviceName, subsystem, msg string) string {
	parts := []string{publicTextPrefix}
	if serviceName != "" {
		parts = append(parts, serviceName)
	}
	if subsystem != "" {
		parts = append(parts, subsystem)
	}
	words := strings.FieldsFunc(msg, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	if len(words) > 0 {
		for i := 1; i < len(words); i++ {
			r := []rune(words[i])
			r[0] = unicode.ToUpper(r[0])
			words[i] = string(r)
		}
		parts = append(parts, strings.Join(words, ""))
	}
	return strings.Join(parts, ".")
}
