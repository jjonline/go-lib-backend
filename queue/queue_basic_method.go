/*
 * @Time   : 2021/1/20 下午10:10
 * @Email  : jjonline@jjonline.cn
 */
package queue

import (
	"encoding/json"
)

// queueBasic 队列基础公用方法
type queueBasic struct{}

// region 获取队列相关名称私有方法

// name 获取队列名称
func (r *queueBasic) name(queue string) string {
	return queue
}

// reservedName 获取队列执行中zSet名称
func (r *queueBasic) reservedName(queue string) string {
	return queue + ":reserved"
}

// delayedName 获取队列延迟zSet名称
func (r *queueBasic) delayedName(queue string) string {
	return queue + ":delayed"
}

// marshalPayload 初始化创建生成队列内部存储的payload字符串
// @task	  队列任务类实例
// @taskParam 队列job参数
// @ID	      队列job编号ID（延迟队列）
func (r *queueBasic) marshalPayload(task TaskIFace, taskParam interface{}, ID string) ([]byte, error) {
	if ID == "" {
		// 补充一个尽量唯一的编号ID
		ID = FakeUniqueID()
	}
	return json.Marshal(Payload{
		Name:          task.Name(),
		ID:            ID,
		MaxTries:      task.MaxTries(),
		RetryInterval: task.RetryInterval(),
		Attempts:      0,
		Payload:       []byte(IFaceToString(taskParam)),
		PopTime:       0,
	})
}

// unmarshalPayload 解析生成队列内部存储的payload字符串为struct
// @payload 队列内部存储的payload字符串
func (r *queueBasic) unmarshalPayload(payload []byte, result *Payload) error {
	return json.Unmarshal(payload, result)
}

// endregion
