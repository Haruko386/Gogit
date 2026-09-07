# Gogit 后续开发计划

本文档记录 Gogit 当前尚未完成的功能、已知限制、建议优先级与验收标准。

当前版本已经具备持久 PTY Shell、基础行编辑、会话内历史、Git 静态候选、Tab 补全、参数值提示，以及包含虚拟环境和当前目录的动态 Prompt。后续工作的总体原则是：不破坏普通命令行行为，只在 Git 输入场景中增加便利。

## P0：发布前必须解决

### 1. 候选窗口限制与滚动

当前 Renderer 会绘制全部候选。随着 Git 目录扩大，候选可能超过终端高度，破坏输入区域。

需要实现：

- 默认最多显示 5～8 条候选；
- 选中项超出可见区域时滚动窗口；
- 窗口顶部或底部显示仍有隐藏候选；
- 终端高度不足时自动缩小候选窗口；
- 清除旧候选时不能留下残影。

验收标准：输入 `git ` 后候选不会超出当前终端，连续按上下键可以访问所有候选。

### 2. 多行粘贴与输入队列

当前同一个输入数据块中遇到 Enter 后会退出按键循环，Enter 后面的内容可能被丢弃。

需要实现：

- 保留 Enter 后尚未处理的输入；
- 支持一次粘贴多条命令；
- Shell 执行期间缓存后续完整命令，或采用明确的透传规则；
- 正确处理 `\r`、`\n` 和 `\r\n`；
- 明确 bracketed paste 的处理方式。

验收标准：一次粘贴下面两行时，两条命令都按顺序执行且没有字符丢失：

```text
git status
go test ./...
```

### 3. Linux 用户 Shell 配置兼容

当前 Bash 使用 `--norc`，Zsh 使用 `-f`，可能导致用户的 alias、函数、PATH、Conda 初始化及 Shell 插件不可用。

需要实现：

- 加载用户原有的 `.bashrc`、`.zshrc` 等配置；
- 在用户配置加载完成后安装 Gogit Prompt 协议；
- 避免 Conda、venv 或主题再次覆盖协议 Prompt；
- 保留用户 alias、函数和 PATH；
- 为 Bash、Zsh 和基础 POSIX sh 分别验证。

验收标准：进入 Gogit 后，用户原有 alias 和 `conda activate` 可用，环境名与当前目录仍能正确更新。

### 4. Prompt 协议恢复能力

用户重新定义 Prompt，或环境激活脚本覆盖 Prompt 后，Shell 可能不再输出 Gogit frame，UI 会一直停留在执行模式。

需要实现：

- 检测 Prompt 协议是否丢失；
- 设计协议重装或恢复机制；
- 避免正常的长时间命令被误判为协议失败；
- 保证随机 marker 不会显示给用户；
- 协议异常时提供可理解的错误或安全降级到透传模式。

验收标准：覆盖 Shell Prompt 后，Gogit 不会永久卡死，并能恢复编辑或明确降级。

### 5. Unicode 终端显示宽度

当前光标位置按 rune 数量计算，但中文、全角字符和部分 Emoji 通常占用多个终端单元格。

需要实现：

- 区分 UTF-8 字节、rune 索引和终端 cell width；
- Prompt、输入行和光标移动统一使用显示宽度；
- 处理组合字符与常见 Emoji；
- 保持 Editor 内部继续使用 rune 边界，避免拆开 UTF-8 字符。

验收标准：目录和命令中包含中文或 Emoji 时，光标、删除和重绘位置正确。

## P1：Git 辅助核心能力

### 6. 扩充静态 Git 命令目录

当前只覆盖部分常见子命令和 option。需要逐步增加：

- `clone`、`init`、`remote`、`tag`；
- `reset`、`revert`、`cherry-pick`；
- `bisect`、`worktree`、`submodule`；
- 每个命令常用的长参数和短参数别名；
- 参数说明、是否需要值、是否可重复及互斥关系。

不要一次机械录入全部 Git 参数。优先按常用程度扩展，并为每组数据补充行为验证。

验收标准：常用 Git 工作流可以只依赖 Gogit 提示发现主要命令和 option。

### 7. 扩展参数值提示

当前参数值提示主要覆盖 `git commit --message/-m`。

下一批建议支持：

```text
git commit --author <author>
git commit --date <date>
git commit --file <file>
git switch --create <branch>
git branch --delete <branch>
git remote add <name> <url>
git clone <repository> <directory>
```

参数元数据需要表达：

- 是否需要值；
- 值名称和说明；
- 是否允许 `--option=value`；
- 是否允许重复；
- 是否存在短参数别名；
- 是否与其他参数互斥；
- 值来自用户输入、文件、分支、remote 还是其他动态来源。

验收标准：需要值的 option 不会错误显示普通 option 列表，也不会把 `<value>` 占位文本插入命令。

### 8. 动态 Git 候选

静态目录只能解释命令和 option，无法提示当前仓库中的真实对象。

需要支持：

- 本地分支；
- 远程分支；
- tag；
- remote；
- 已修改文件与未跟踪文件；
- 适合当前命令的 revision 或路径。

示例：

```text
git switch fe
    feature/login
    feature/parser

git add int
    internal/editor/editor.go
    internal/suggest/engine.go
```

动态查询必须使用持久子 Shell 的真实当前目录，而不是 Gogit 父进程的工作目录。还需要避免每次按键都启动昂贵查询，可使用缓存、取消和防抖。

验收标准：执行 `cd` 进入不同仓库后，候选立即对应新仓库，慢查询不会阻塞键盘输入。

### 9. 更完整的 Git 与 Shell token 解析

当前解析只适合简单 Git 命令，引号和转义支持也只是最小实现。

需要逐步处理：

```text
git -C ../repo status
git --no-pager log
cd repo; git status
git status | less
git add "file with spaces.txt"
git commit -m 'fix message'
```

需要注意：

- PowerShell、Bash 和 Zsh 的引用及转义规则不同；
- 光标可能位于 token 中间；
- `|`、重定向、`;`、`&&` 和 `||` 会创建新的命令边界；
- 不应尝试一次实现完整 Shell Parser，应按实际候选需求逐步扩展。

验收标准：解析失败时宁可不提示，也不能替换错误的文本范围或生成危险命令。

## P2：一般命令行体验

### 10. 常用编辑快捷键

建议支持：

- Ctrl+L：清屏并重绘；
- Ctrl+A / Ctrl+E：移动到行首或行尾；
- Ctrl+W：删除前一个单词；
- Ctrl+U：清除光标前内容；
- Ctrl+K：清除光标后内容；
- Ctrl+R：搜索历史；
- Esc：退出候选选择；
- Tab 补全后按上下文决定是否添加空格。

验收标准：快捷键行为尽量与主流 PowerShell/Bash 行编辑体验一致。

### 11. 历史持久化

当前历史仅保存在本次 Gogit 进程内，且没有容量上限。

需要实现：

- 限制内存历史条数；
- 保存到用户数据目录；
- 启动时加载，退出或执行后安全写入；
- 避免并发 Gogit 进程互相覆盖；
- 提供关闭持久化的配置；
- 考虑 token、密码等敏感命令的过滤策略。

验收标准：重启 Gogit 后可浏览历史，历史文件损坏或不可写时不会影响 Shell 使用。

### 12. Prompt 与显示配置

需要支持：

- 自定义颜色和候选数量；
- 显示或隐藏环境名、当前目录和 Gogit 前缀；
- 可选显示上一条命令退出码；
- 小窗口和浅色/深色终端适配；
- 对目录名和环境名中的控制字符进行安全处理。

验收标准：配置错误时使用安全默认值，Prompt 不会破坏终端状态。

## P3：可靠性、测试与发布

### 13. 交互式程序兼容性

需要实际验证：

- `vim` 或其他全屏编辑器；
- `less`、`top` 等全屏程序；
- Python/Node REPL；
- `ssh`；
- 需要 Ctrl+C、Ctrl+D 或窗口 resize 的程序；
- 修改终端 echo 模式的程序。

验收标准：程序退出后 Gogit Prompt、光标、颜色和 raw mode 能正确恢复。

### 14. 端到端与并发测试

当前测试主要覆盖纯逻辑，仍需补充：

- PTY 启动、读写、关闭和子进程回收；
- PowerShell 5.1 与 PowerShell 7；
- Bash、Zsh 和 POSIX sh；
- Prompt frame 被拆分、混入大量输出及连续出现；
- 快速输入、慢命令和 resize 同时发生；
- Shell 意外退出；
- 输出管道关闭和终端恢复错误；
- 多行粘贴与交互模式切换。

验收标准：正常退出、异常退出及 Ctrl+C 后，宿主终端均不会处于损坏状态。

### 15. 安装、版本与发布

需要完成：

- `gogit --version` 与基础帮助；
- Windows、Linux、macOS 构建；
- release artifact 与校验值；
- 安装和卸载说明；
- 配置文件与用户数据目录约定；
- README 中已实现功能与计划功能同步；
- CI 中运行测试、vet 和跨平台构建。

验收标准：新用户可以按照 README 安装、运行、升级和卸载 Gogit。

## 推荐实施顺序

建议后续按以下顺序推进：

1. 候选窗口限制与滚动；
2. 多行粘贴与输入队列；
3. Linux 用户 Shell 配置兼容；
4. Prompt 协议恢复能力；
5. 参数值提示扩展；
6. 动态分支、remote、tag 和文件候选；
7. 更完整的 Git/Shell token 解析；
8. Unicode 显示宽度与编辑快捷键；
9. 历史持久化；
10. 端到端测试、配置与发布流程。

每完成一个阶段，都应至少执行：

```powershell
gofmt -w <changed-files>
go test ./...
go vet ./...
```

涉及终端交互时，还必须在真实 PowerShell/Bash/Zsh 中进行手动验收，不能只依赖单元测试。
