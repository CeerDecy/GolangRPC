package crpc_error

type CrError struct {
	err    error
	ErrFun ErrorFunc
}

func Default() *CrError {
	return &CrError{}
}

func (c *CrError) Error() string {
	return c.err.Error()
}

// Put 将错误放入CrError并抛出异常
func (c *CrError) Put(err error) {
	c.check(err)
}

// 检查是否为空
func (c *CrError) check(err error) {
	if err != nil {
		c.err = err
		panic(c)
	}
}

// ErrorFunc 异常处理函数
type ErrorFunc func(c *CrError)

// Result 将异常处理暴露给用户自定义
func (c *CrError) Result(fun ErrorFunc) {
	c.ErrFun = fun
}

// ExecResult 执行异常处理函数
func (c *CrError) ExecResult() {
	c.ErrFun(c)
}
