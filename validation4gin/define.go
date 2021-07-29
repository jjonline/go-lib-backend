package validation4gin

import "strings"

// Message 自定义validate失败时错误消息
type Message map[string]string

// FieldMap 自定义检查字段filed与自定义名称映射关系
type FieldMap map[string]string

// MessageBag 处理结果集响应值结构
type MessageBag []string

// IsEmpty 检查消息bag是否为空
func (m MessageBag) IsEmpty() bool {
	return len(m) == 0
}

// First 获取第一条文案，没有则返回空字符串
//  不要依赖返回空字符串判断 MessageBag 为空，而应该使用 IsEmpty 方法
func (m MessageBag) First() string {
	if !m.IsEmpty() {
		return m[0]
	}
	return ""
}

// All 获取所有文案
//  - sep 多条文案的分隔符，没有则返回空字符串
//  不要依赖返回空字符串判断 MessageBag 为空，而应该使用 IsEmpty 方法
func (m MessageBag) All(sep string) string {
	if !m.IsEmpty() {
		strings.Join(m, sep)
	}
	return ""
}
