package main

import "fmt"

type Note struct {
	Title string
}

func runSliceLesson() {
	// 创建一个空的 Note 切片。
	notes := make([]Note, 0)

	// 每次 append 追加一条笔记。
	notes = append(notes, Note{Title: "学习函数"})
	notes = append(notes, Note{Title: "学习结构体"})
	notes = append(notes, Note{Title: "学习切片"})

	// len 返回切片当前包含的元素数量。
	fmt.Println("笔记总数：", len(notes))

	// 下标从 0 开始，所以 notes[0] 是第一条笔记。
	fmt.Println("第一条笔记：", notes[0].Title)

	// range 依次取出每条笔记的下标和值。
	for index, note := range notes {
		fmt.Println("下标：", index, "标题：", note.Title)
	}
}
