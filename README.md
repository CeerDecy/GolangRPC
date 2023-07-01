## 实现一个Golang的RPC微服务框架——crpc
### 一、路由的实现

1. 路由前缀树
2. 路由表

### 二、中间件

#### 要求

1. 不耦合在用户的代码中
2. 独立存在，能拿到上下文，并能做出影响（闭包）

#### 实现

1. ~~前置中间件~~
2. ~~后置中间件~~
3. 通用中间件
4. 路由级别中间件

### 三、页面渲染

#### HTML
1. 加载html文件

   提前加载到内存中
#### JSON
#### XML
#### 文件下载
#### 重定向

### 四、参数处理
#### 1. 优化Context
sync.Pool用于存储那些被分配了但没有被使用，但是未来可能被使用的值，这样可以不用再次分配内存，提高效率

sync.Pool大小是可伸缩的，高负载是会动态扩容，存放在池中不活跃的对象会被自动清理
#### 2. Query参数
#### 3. Post表单参数
#### 4. 文件上传
#### 5. Json参数解析
#### 6. 参数验证器Validate
### 三、日志处理
#### 1. 日志中间件
#### 2. 分级日志 ["debug","info","error"]
#### 3. 日志持久化存储
### 四、错误处理
#### 1. Recovery中间件
### 五、协程池
#### 1. 协程池
1. 创建固定数量的协程
2. 有一个任务队列，等待协程调度执行
3. 协程用完时，其他任务处于等待状态，一旦有协程空闲，立即执行任务
4. 协程长时间空闲，则清理，以免占用空间
5. 设置任务超时时间，若一个任务长时间完成不了，就主动让出协程
#### 2. 引入sync.pool
#### 3. 引入sync.Cond
sync.Cond是基于互斥锁/读写锁实现的条件变量，用来协调那些想要访问共享资源的那些Goroutine

此处用于等待获取空闲的Worker，替换原有的for死循环

#### 并发量在2000000及以下的时候可以有比较不错的优势，在时间基本持平的情况下内存占用大大减少，并发数越高时间消耗会更多
### 六、认证
#### 1. HTTPS支持





# 我的收获
1. 单例模式实现参数验证器减少频繁New带来的系统开销
2. 用Golang的反射机制实现对Json参数的解析以及判定
3. 使用闭包完成中间件