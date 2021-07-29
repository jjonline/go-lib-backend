# validation4gin

## 一、说明

````
type ValidRuleRequest struct {
    ID uint `form:"id" json:"id" binding:"required,min=10,max=1000"`
}

func ValidRule(ctx *gin.Context) {
    var req ValidRuleRequest
	
    // 这个err无法个性化自定义 需要转换翻译
    err := ctx.ShouldBindQuery(&req)
}
````
上述代码为基于gin框架的典型的参数绑定，并且具有表单检查的功能

 * 当参数绑定失败或表单检查不通过时返回的`error`信息有多种类型：绑定失败相关、表单检查相关 等
 * 直接`err.Error()`转换的字符串对非编程人员而言识读性差

本库尝试按一定规则来解析解析这个err并定制化返回文案

> 因 gin 框架底层默认使用的 `go-playground/validator` 这个表单检查库
> 实质上本库是在尝试翻译 `go-playground/validator` 的返回值而不是使用官方库既重又不够个性化的i18n翻译。

## 二、用法

样例

````
type ValidRuleRequest struct {
    ID uint `form:"id" json:"id" binding:"required,min=1,max=1000"`
    Name string `form:"name" json:"name" binding:"required,max=255"`
}

var message = validation4gin.

func ValidRule(ctx *gin.Context) {
    var req ValidRuleRequest
	
    err := ctx.ShouldBindQuery(&req)
    
    // 定义Rule规则下的错误文案
    message := validation4gin.Message{
        "ID.*"         :  ":ID错误",
        "ID.kind"      :  ":ID类型错误",
        "ID.min"       :  ":ID不得小于10",
        "ID.max"       :  ":ID不得大于100",
        "Name.required":  "用户:Name必填",
        "Name.max"     :  "用户:Name长度最大支持255个字符",
        "Name.kind"    :  "用户:Name参数类型错误",
    }
    // 定义表单相关字段对应值
    // 如果绑定的到struct则是结构体的字段名FieldName（而不是传参form表单字段名）
    // 如果绑定的到map，则是map的key
    fieldMap := validation4gin.FieldMap{
        "ID":   "编号",
        "Name": "名称",
    }
    if err != nil {
        msg := validation4gin.Translate(err, message, fieldMap)
		fmt.Println(msg.First()) // 将会打印出翻译转换后的文案
    }
}
````

`validation4gin.Message` 类型 
 - 键 构成规则是 `字段名.表单检查规则`
   - `FieldName.-rule-` 表示表单检查这个规则不通过时的文案(优先级1)；
   - `FieldName.kind` `kind`是一个特定的规则（即 `validation4gin.KindKey` 常量），表示参数绑定时传参类型与绑定的结构不符或越界时使用该文案(优先级1)；
   - `FieldName.*` 表示字段下通配，该字段下未定义的规则被触发时通配使用(优先级2)；
   - `-rule-` 表示全字段通配，某个字段未定义任何rule规则下则是哟好难过全字段通配规则文案(优先级3)；
 - 值 为自定义文案，文案中可使用 `:`开头的字段名作为变量，即`validation4gin.FieldMap`的键名并最终把变量替换掉；

`validation4gin.FieldMap` 类型
 - 键 为字段名，结构体的字段或map的键名
 - 值 为这个字段映射的自定义文案名
