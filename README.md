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