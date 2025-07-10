## rtm模块
```
    // ① 长时间任务处理
	go func() { ... p.dealLongTimeProcess() }()
	// ② 定时分表
	go func() { ... p.subTable() }()
	// ③ 定时更新任务位置信息
	go func() { ... p.UpdateSchedulePos() }()
```

`rtm.go` 文件实现了短任务的运行时（TaskRuntime），主要用于定时管理和调度异步任务表的分表、位置更新和超时处理。下面分块详细解释其所有核心逻辑：

---

## 1. 结构体定义

```go
type TaskRuntime struct {}
```
这是一个空结构体，用来承载运行时相关的方法。

---

## 2. 启动入口

```go
func (p *TaskRuntime) Run() {
	p.run()
}
```
Run 方法公开对外，内部实际调用 run 方法。

---

## 3. 多协程调度

```go
func (p *TaskRuntime) run() {
	// ① 长时间任务处理
	go func() { ... p.dealLongTimeProcess() }()
	// ② 定时分表
	go func() { ... p.subTable() }()
	// ③ 定时更新任务位置信息
	go func() { ... p.UpdateSchedulePos() }()
}
```
- 启动后会并发运行三个核心流程，每个流程都有自己的异常恢复（recover）逻辑，保证 panic 不会导致主进程崩溃，并且打印详细调用栈。

---

## 4. UpdateSchedulePos（定时更新任务位置）

```go
func (p *TaskRuntime) UpdateSchedulePos() {
	for {
		// 定时器，间隔 configurable
		t := time.NewTimer(time.Duration(config.Conf.Task.MoveInterval) * time.Second)
		<-t.C
		p.UpdateSchedulePosProcessing()
	}
}
```
- 按照配置的间隔定时触发。
- 调用 `UpdateSchedulePosProcessing` 处理实际的任务位置表更新。

### 具体处理逻辑

```go
func (p *TaskRuntime) UpdateSchedulePosProcessing() {
	taskPosList, err := db.TaskPosNsp.GetTaskPosList(db.DB)
	for _, taskPos := range taskPosList {
		// 获取当前类型任务的完成数量与总数量
		finishNum := db.TaskNsp.GetFinishTaskCount(...)
		count := db.TaskNsp.GetAllTaskCount(...)
		// 如果所有任务都完成，并且 beginPos < endPos，则更新 beginPos
		if finishNum == count && taskPos.ScheduleBeginPos < taskPos.ScheduleEndPos {
			taskPos.ScheduleBeginPos++
			db.TaskPosNsp.Save(db.DB, taskPos)
		}
	}
}
```
- 作用：移动任务调度位置（beginPos），保证 beginPos 总是指向未完成的任务区间。

---

## 5. subTable（定时检查分表）

```go
func (p *TaskRuntime) subTable() {
	for {
		t := time.NewTimer(time.Duration(config.Conf.Task.SplitInterval) * time.Second)
		<-t.C
		p.subTableProcessing()
	}
}
```
- 按配置间隔定期检查是否需要分表。

### 具体处理逻辑

```go
func (p *TaskRuntime) subTableProcessing() {
	taskPosList, err := db.TaskPosNsp.GetTaskPosList(db.DB)
	for _, taskPos := range taskPosList {
		count := db.TaskNsp.GetAllTaskCount(...)
		if count >= config.Conf.Task.TableMaxRows {
			// 超过最大行数，创建新表
			nextPos := db.TaskPosNsp.GetNextPos(...)
			db.TaskNsp.CreateTable(db.DB, taskType, nextPos)
			taskPos.ScheduleEndPos++
			db.TaskPosNsp.Save(db.DB, taskPos)
		}
	}
}
```
- 作用：当任务表行数超过阈值时自动分表，并更新表的 endPos 指针。

---

## 6. dealLongTimeProcess（定时处理长时间未完成的任务）

```go
func (p *TaskRuntime) dealLongTimeProcess() {
	for {
		t := time.NewTimer(time.Duration(config.Conf.Task.LongProcessInterval) * time.Second)
		<-t.C
		p.dealTimeoutProcessing()
	}
}
```
- 定期检测和处理长时间未完成的任务。

### 具体处理逻辑

```go
func (p *TaskRuntime) dealTimeoutProcessing() {
	taskTypeCfgList, err := db.TaskTypeCfgNsp.GetTaskTypeCfgList(db.DB)
	for _, taskTypeCfg := range taskTypeCfgList {
		p.dealTimeoutProcessingWithType(taskTypeCfg)
	}
}
```
- 遍历所有任务类型配置，分别处理。

#### 单类型处理细节

```go
func (p *TaskRuntime) dealTimeoutProcessingWithType(taskCfg *db.TaskScheduleCfg) {
	taskPos := db.TaskPosNsp.GetTaskPos(db.DB, taskCfg.TaskType)
	maxProcessTime := config.Conf.Task.MaxProcessTime
	if int64(taskCfg.MaxProcessingTime) == 0 {
		maxProcessTime = taskCfg.MaxProcessingTime
	}
	taskList := db.TaskNsp.GetLongTimeProcessing(db.DB, ...)
	for _, task := range taskList {
		if task.CrtRetryNum >= taskCfg.MaxRetryNum {
			db.TaskNsp.SetStatus(db.DB, task.TaskId, db.TASK_STATUS_FAILED)
			continue
		}
		db.TaskNsp.SetStatusAndRetryNumIncrement(db.DB, task.TaskId, db.TASK_STATUS_PENDING)
	}
}
```
- 作用：对于超时未完成的任务，如果已达最大重试次数则置为失败，否则重试次数+1并重置为待执行。

---

## 7. 全局异常保护

每个 goroutine 都用了 `defer recover`，防止异常崩溃，并详细记录错误和调用栈。

---

## 总结

- **TaskRuntime** 作为短任务的调度与管理核心，负责定时：
  - **分表**：自动扩展表容量，避免单表过大。
  - **轮转调度位置**：保证 beginPos/endPos 始终指向正确的任务区间。
  - **处理超时任务**：重试或标记为失败，保证任务系统健壮性。
- **高可用**：所有流程都用 recover 防御 panic，保证服务稳定运行。

如需对每个函数或某段逻辑进一步细化，请随时告知！


根据你提供的 flowsvr/src/rtm/rtm.go 源码，和其依赖（如 config、db、martlog、requestid 等）来看，rtm.go 的实现细节有以下值得注意的地方：

---

## 1. 是否使用了 Redis？

**rtm.go 文件本身并未直接引用或操作 Redis。**
- 文件里没有显式引入 redis 客户端包（如 github.com/go-redis/redis 等）。
- 也没有看到类似 redis.Get、redis.Set、redisClient 等调用。
- 但是，部分注释中提到 “get lock” “unlock”，比如：
  ```go
  // lockKey := "subTable"
  // lockKey := SHORT_TASK_LONGTIME_DEAL_LOCK_KEY
  ```
  这说明在分布式部署下，理论上需要用分布式锁（如 Redis 实现的 Redlock），但当前代码注释掉了锁相关实现，实际并未加锁。  
  **结论：本文件没有实际用 Redis，但如果你要部署多实例，强烈建议加上分布式锁，通常会用 Redis。**

---

## 2. 错误恢复与日志

- 所有 goroutine 都用 defer recover() + debug.PrintStack()，保证即使 panic 也能捕捉并打印详细堆栈，防止服务崩溃。
- 日志系统（martlog）用于打印所有关键流程的 info 和 error，便于排查异常。

---

## 3. 数据库依赖

- 所有任务表、位置表的增删查改，全部通过 db 包实现（如 db.TaskNsp, db.TaskPosNsp）。
- 所涉及的表包括：
  - t_task_type_beginPos / t_task_type_endPos
  - t_schedule_pos
  - 任务主表
- 代码未指定数据库类型（MySQL、Postgres 等），但一般是关系型数据库。

---

## 4. 配置依赖

- 定时器的时间间隔、表最大行数、最大处理时间等，全部通过 config.Conf.Task.* 配置项动态配置。
- 这意味着你可以不用重启服务，直接通过配置文件调整调度参数。

---

## 5. 并发与定时

- 用了 go func() 启动多个独立 goroutine 并发处理（分表、位置更新、超时处理），互不阻塞。
- 每个周期都是通过 time.NewTimer 实现定时调度。

---

## 6. 任务超时重试/失败

- dealTimeoutProcessingWithType 里对超时任务的处理逻辑：
  - 如果任务已重试到最大次数，直接置为失败（TASK_STATUS_FAILED）。
  - 否则重试次数 +1，状态置为待执行（TASK_STATUS_PENDING）。
- 这保证了任务不会无限重试，且失败任务会被清晰标记。

---

## 7. 任务分表

- 分表逻辑是以任务数量为阈值（config.Conf.Task.TableMaxRows），自动扩展新表，维护表的 endPos。
- 确保单表不会因数据量太大影响性能。

---

## 8. RequestID 跟踪

- dealLongTimeProcess 的 goroutine 启动时会生成唯一 requestID，便于分布式链路追踪和日志追溯。

---

## 9. 代码中的“锁”逻辑

- 目前所有“加锁”相关代码都被注释掉了，意味着如果你有多实例部署，**可能会有并发安全问题**（如重复分表、重复位置更新等）。
- 实际生产环境建议补充分布式锁（如 Redis、Zookeeper、Etcd 等），否则多实例下会有竞态风险。

---

## 10. 其他注意事项

- 数据库操作未做批量处理，表数据量大时可能有性能隐患。
- 配置和表结构变更需同步更新，不然定时任务可能异常。
- 没有直接的健康检查，但可以通过日志监控 goroutine 是否正常调度。
- 没有显式退出机制，所有 for 循环为死循环，除非主进程崩溃。

---

## 总结

- rtm.go 本身未用 Redis，但分布式部署建议加锁（常用 Redis）。
- 所有核心操作依赖 db 层，定时参数可配置。
- goroutine + recover + 日志保证高可用。
- 没有锁的话多实例部署不安全，需特别注意。
- 业务表、分表、任务位置表等均需要数据库支持。

如要了解某个细节的实现或有其它问题，欢迎继续追问！