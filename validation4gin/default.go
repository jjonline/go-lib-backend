package validation4gin

// 定义默认message
var (
	// ValidationRuleTypeErrorMessage go-playground/validator表单检查类型错误
	//  出现此种类型错误表明代码写法有严重问题，需要手动修正
	//  gin框架下几乎不会出现
	ValidationRuleTypeErrorMessage = "参数类型错误请检查代码：%s"
	// ValidationDefaultMessage 未定义rule规则下文案时默认文案
	//  当未定义rule规则对应的消息体时的默认消息体
	ValidationDefaultMessage = "字段%s规则%s不通过"
)
