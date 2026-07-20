package main

import "fmt"

// Counter 用来演示指针：它的方法和项目里 Store 的方法原理一样。
type Counter struct {
	Count int
}

// 值传递：函数收到的是 c 的副本，改副本不影响原值。
func incrementByValue(c Counter) {
	c.Count++
}

// 指针传递：函数收到的是 c 的地址，能改到原值。
func incrementByPointer(c *Counter) {
	c.Count++ // 语法糖：c 是指针时，c.Count 等价于 (*c).Count
}

// 实现Double练习
func Double(c *Counter) {
	c.Count *= 2
}

func runPointerLesson() {
	c := Counter{Count: 0}

	// 值传递：改的是副本，原 c 不变。
	incrementByValue(c)
	fmt.Println("值传递后 Count =", c.Count) // 期望 0

	// &c 取 c 的地址传进去，原 c 被修改。
	incrementByPointer(&c)
	fmt.Println("指针传递后 Count =", c.Count) // 期望 1

	// ===== 练习 =====
	// 项目里 Store 的方法都用 *Store 接收者（如 func (store *Store) CreateNote），
	// 因为方法要修改 Store 内部状态。Counter 同理。
	//
	// 任务：在下面给 Counter 写一个 Double 方法，让 Count 翻倍。接收者必须用 *Counter。
	// 写完后取消下面三行注释运行，期望输出 2。
	//
	// c2 := Counter{Count: 1}
	// c2.Double()
	// fmt.Println("练习：Double 后 Count =", c2.Count) // 期望 2
	Double(&c)
	fmt.Println("指针传递后 Count =", c.Count) // 期望 2
}
