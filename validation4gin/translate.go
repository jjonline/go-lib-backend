package validation4gin

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"strconv"
	"strings"
)

// 定义常量
const (
	KindKey       = "kind"
	AttributeFlag = ":"
)

// Translate
//  - message的键 由 结构体字段名、点字符、验证器验证规则、星通配符构成
//  - message的值 给出响应文案内容
//  - message自定义消息map键构成形式的优先级定义如下 (Message Map key define use for priority)
//    - Field.rule
//    - Field.*
//    - rule
func Translate(err error, message Message, fieldMap FieldMap) MessageBag {
	if nil == err {
		return MessageBag{}
	}

	switch err.(type) {
	case *validator.InvalidValidationError:
		// 检查目标类型错误，譬如：要求检查结构体却传参了一个map
		// 这种场景的代码是必须修正的，传参都错误了代码严重的bug，文案仅提示给开发人员不宜展示给用户
		return MessageBag{
			fmt.Sprintf(ValidationRuleTypeErrorMessage, err.(*validator.InvalidValidationError).Type.String()),
		}
	case validator.ValidationErrors:
		return translateValidErr(err.(validator.ValidationErrors), message, fieldMap)
	case *strconv.NumError:
		// 一般是转换为数值类型时发生错误：无法转换 or 超过数值类型的长度
		// https://github.com/gin-gonic/gin/issues/2334
		// strconv 库返回的错误本身无法获取到是那个字段转换失败的
		return MessageBag{toPriorityMessage("", KindKey, message, fieldMap)}
	case *json.UnmarshalTypeError:
		return translateUnmarshalErr(err.(*json.UnmarshalTypeError), message, fieldMap)
	default:
		// 其他类型错误：例如解析PostForm出错，如果有*通配则返回*文案否则返回默认noCover
		if _, ok := message["*"]; ok {
			return MessageBag{toPriorityMessage("", "*", message, fieldMap)}
		}
		return MessageBag{ValidationNoCoverMessage}
	}
}

// translateValidErr 翻译表单检查错误
func translateValidErr(err validator.ValidationErrors, message Message, fieldMap FieldMap) MessageBag {
	ms := MessageBag{}
	for _, item := range err {
		ms = append(ms, toPriorityMessage(item.Field(), item.Tag(), message, fieldMap))
	}

	return ms
}

// translateUnmarshalErr 翻译表单绑定时参数类型错误
//  一般是参数绑定映射到结构体时出现错误
func translateUnmarshalErr(err *json.UnmarshalTypeError, message Message, fieldMap FieldMap) MessageBag {
	return MessageBag{toPriorityMessage(err.Field, KindKey, message, fieldMap)}
}

// toPriorityMessage 按优先级分析获取消息体
//  - field   字段字符串-结构体的Filed名
//  - rule	  检查规则
//  - message 自定义消息体
func toPriorityMessage(field, rule string, message Message, fieldMap FieldMap) string {
	var result string

	// default message
	result = fmt.Sprintf(ValidationDefaultMessage, field, rule)

	// -step1 Field.rule
	if val, ok := message[field+"."+rule]; ok {
		result = val
	} else {
		// -step2 Field.*
		if val, ok = message[field+".*"]; ok {
			result = val
		} else {
			// -step3 rule
			if val, ok = message[rule]; ok {
				result = val
			} else {
				// -step4 *
				if val, ok = message["*"]; ok {
					result = val
				}
			}
		}
	}

	// 转换字段名映射的自定义名称
	if strings.Contains(result, AttributeFlag) {
		for _filed, name := range fieldMap {
			result = strings.ReplaceAll(result, AttributeFlag+_filed, name)
		}
	}

	return result
}
