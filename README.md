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