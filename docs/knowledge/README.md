# 知识内容源

这个目录存放可以导入 Dev Notebook 的学习内容，和 `project-playbook.md` 的项目过程记录分开。

每篇知识条目都是 Markdown 文件，并使用 YAML front matter 描述标题、分类、标签和摘要。后续网站实现导入功能时，可以读取这些字段并写入 SQLite。

目录按技术分类，例如：

```text
docs/knowledge/
  go/
  vue/
  database/
```

每篇笔记应包含：一句话结论、最小例子、项目中的实际用法、前端类比、容易误解和个人练习。
