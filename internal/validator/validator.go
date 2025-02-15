package validator

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// Validator 结构体用于存储字段验证错误
// FieldErrors 是一个 map，key 为字段名，value 为错误信息
type Validator struct {
	FieldErrors map[string]string
}

// Valid 方法检查是否有任何验证错误
// 返回 true 如果没有错误，false 如果有错误
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

// AddFieldError 方法添加一个字段错误到 FieldErrors map 中
// key 为字段名，message 为错误信息
// 如果 FieldErrors 为 nil，会先初始化
// 如果该字段已经存在错误，则不会覆盖
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// CheckField 方法检查字段是否通过验证
// 如果 ok 为 false，则调用 AddFieldError 添加错误信息
// key 为字段名，message 为错误信息
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank 函数检查字符串是否为空或仅包含空白字符
// 返回 true 如果字符串不为空，false 如果为空
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars 函数检查字符串的字符数是否小于等于 n
// 使用 utf8.RuneCountInString 来正确计算多字节字符
// 返回 true 如果字符数 <= n，false 如果大于 n
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// PermittedValue 泛型函数检查值是否在允许的值列表中
// 使用 slices.Contains 进行检查
// 返回 true 如果值在允许列表中，false 如果不在
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
