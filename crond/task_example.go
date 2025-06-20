package crond

type taskExample struct {
}

func (t taskExample) Signature() string {
	return "taskExample"
}

func (t taskExample) Desc() string {
	return "定时任务示例"
}

// Rule 定时规则：`Second | Minute | Hour | Dom (day of month) | Month | Dow (day of week)`
//   - 注意：是精确到秒的，如果你的任务不需要精确到秒，则秒规则给一个确切的数字比如0，给*表示在你的规则内每秒触发1次
//   - ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//     *
//   - *       *    *    *    *    *
//   - -	   -    -    -    -    -
//   - |	   |    |    |    |    |
//   - |	   |    |    |    |    |
//   - |	   |    |    |    |    +----- day of week (0 - 7) (Sunday=0 or 7)
//   - |	   |    |    |    +---------- month (1 - 12)
//   - |	   |    |    +--------------- day of month (1 - 31)
//   - |	   |    +-------------------- hour (0 - 23)
//   - |	   +------------------------- min (0 - 59)
//   - +--------------------------------- sec (0 - 59)
//     *
//   - ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
func (t taskExample) Rule() string {
	// 每5分钟执行1次
	// 此处请注意，定时规则是精确到秒的，秒位不能给*，否则表示在你指定规则内的每1秒触发1次
	return "0 */5 * * * *"
}

func (t taskExample) Execute() error {
	// 你的定时被执行的逻辑
	// 返回nil执行成功，返回error执行失败，发生panic执行失败
	return nil
}
