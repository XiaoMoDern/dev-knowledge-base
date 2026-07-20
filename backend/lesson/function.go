package main

import "fmt"

func multiply(a, b int) int {
	return a * b
}

func runFunctionLesson() {
	result := multiply(2, 3)
	fmt.Println("函数练习：", result)
}
