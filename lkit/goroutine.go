package lkit

import (
	"context"
	"fmt"
	"runtime/debug"

	"tiktok_tool/llog"
)

// SafeGo 安全地启动一个协程，自动添加panic处理
// 用法: SafeGo(func() { ... })
func SafeGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				llog.Error(fmt.Sprintf("%v\n %s\n", r, stack))
			}
		}()
		f()
	}()
}

// SafeGoWithRecover 安全地启动一个协程，自动添加panic处理，并允许自定义recover处理函数
// 用法: SafeGoWithRecover(func() { ... }, func(r interface{}) { ... })
func SafeGoWithRecover(f func(), recoverFunc func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if recoverFunc != nil {
					recoverFunc(r)
				} else {
					stack := debug.Stack()
					llog.Error(fmt.Sprintf("%v\n %s\n", r, stack))
				}
			}
		}()
		f()
	}()
}

// SafeGoWithContext 带上下文的安全协程
// 用法: SafeGoWithContext(ctx, func(ctx context.Context) { ... })
func SafeGoWithContext(ctx context.Context, f func(context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				llog.Error(fmt.Sprintf("%v\n %s\n", r, stack))
			}
		}()
		f(ctx)
	}()
}

// SafeGoWithArgs 支持任意参数的安全协程
// 用法: SafeGoWithArgs(func(args ...interface{}) { ... }, arg1, arg2, ...)
func SafeGoWithArgs(f func(args ...interface{}), args ...interface{}) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				llog.Error(fmt.Sprintf("%v\n %s\n", r, stack))
			}
		}()
		f(args...)
	}()
}

// SafeGoWithCallback 安全地启动一个协程，并在执行完成后调用回调函数
// 用法: SafeGoWithCallback(func() { ... }, func(err interface{}) { ... })
// 参数:
//   - f: 要执行的函数
//   - callback: 回调函数，在f执行完成后调用，如果f发生panic，err为panic的值，否则err为nil
func SafeGoWithCallback(f func(), callback func(err interface{})) {
	go func() {
		var err interface{}
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				llog.Error(fmt.Sprintf("%v\n %s\n", r, stack))
				err = r
			}
			// 执行回调函数
			if callback != nil {
				SafeGo(func() {
					callback(err)
				})
			}
		}()
		f()
	}()
}
