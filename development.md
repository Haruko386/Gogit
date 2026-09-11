# Gogit 后续开发计划

更新时间：2026-09-11

本文档记录 Gogit 当前已经实现的能力、仍然存在的限制、后续优先级与验收标准。状态以当前代码为准；完成功能应同时具备实现和相应测试，不能只更新目录数据。

## 当前进度概览

- 核心 MVP：约 70%。
- Git 辅助核心能力（P1）：约 50%。
- 面向正式发布的整体完成度：约 45%～50%。
- P0 Prompt 协议恢复：已完成。
- P1 静态命令、参数元数据和分支候选：进行中。
- P2 编辑体验、历史持久化和配置：大部分待实现。
- P3 跨平台 CI 与 Release workflow：已建立，完整端到端验收和发布体验仍待补齐。

当前版本已经具备持久 PTY Shell、基础行编辑、会话内历史、Git 静态候选、Tab 补全、参数值提示、本地与远程跟踪分支候选、多行粘贴、Unicode 终端宽度处理，以及包含虚拟环境和当前目录的动态 Prompt。

Windows 使用 PowerShell；Linux 和 macOS 支持 Bash、Zsh 与 POSIX sh，并在安装 Gogit Prompt 前加载用户 Shell 配置。普通命令和前台交互程序运行期间，输入直接交给 PTY，Gogit 只在自己的编辑状态下处理补全和行编辑。

## P0：发布前必须解决

### 1. Prompt 协议与恢复能力

状态：已完成。

已实现并验证：

- 使用会话随机 marker 传输 Prompt 状态，不向终端显示协议内容。
- Prompt 被用户配置或命令覆盖后，通过会话专属恢复函数重新安装协议 Prompt。
- 恢复不依赖命令超时，不会把长时间运行的命令误判为 Prompt 损坏。
- PowerShell wrapper 保留 PowerShell `$?` 和原生程序 `$LASTEXITCODE` 的失败状态。
- Scanner 支持 marker 拆分、连续 frame、损坏 frame、嵌套 begin marker 和超大 frame 后重新同步。
- 超大 frame 的 end marker 即使跨多次读取到达，也会被丢弃，不会泄漏到可见终端输出。
- 恢复函数不可用时显示错误，并继续提供 PTY 透传能力。
- Bash、Zsh、POSIX sh 和 PowerShell 初始化脚本均有自动化覆盖；真实 Shell 的完整平台矩阵仍属于 P3 验收范围。

## P1：Git 辅助核心能力

### 2. 静态 Git 命令目录

状态：进行中。

已实现一级命令：

```text
status      add         commit      branch      switch
checkout    restore     log         diff        merge
rebase      fetch       pull        push        stash
clone       init        remote      tag
```

已实现 `git remote` 二级命令：

```text
add         rename      remove      set-head    set-branches
get-url     set-url     show        prune       update
```

上述命令已经包含一批常用 option、说明、短参数别名和行为测试。候选会根据当前 token 前缀过滤，不会在没有匹配项时固定占用终端高度。

下一批计划：

- 增加 `reset`、`revert`、`cherry-pick` 的一级命令和常用 option。
- 增加 `bisect`、`worktree`、`submodule` 及其二级命令。
- 继续补齐现有命令常用的长参数和短参数别名。
- 为危险操作补充清楚的风险说明，例如强制推送、硬重置和删除分支。

验收标准：常用 Git 工作流可以依赖 Gogit 发现主要命令和 option；新增的每组目录数据都有匹配、插入和排除行为测试。

### 3. 参数元数据与参数值提示

状态：进行中。

已实现的参数元数据可以表达：

- option 是否需要值；
- 值名称和显示说明；
- 长短参数别名；
- option 是否允许重复；
- option 之间的互斥关系；
- `--option=value` 形式；
- 参数值提示只用于显示，不会把 `<value>` 占位文本插入命令。

当前已覆盖部分 `commit`、`clone`、`init`、`remote` 和 `tag` 参数，包括 `remote` 二级命令的重复参数与互斥参数处理。

下一批计划：

```text
git commit --author <author>
git commit --date <date>
git commit --file <file>
git switch --create <branch>
git branch --delete <branch>
git remote add <name> <url>
git clone <repository> <directory>
```

还需要让元数据明确参数值来自自由输入、文件、分支、remote、tag、revision 或其他动态来源。

验收标准：需要值的 option 不会错误显示普通 option 列表；已使用且不可重复的 option 不会再次出现；互斥 option 不会同时建议。

### 4. 动态 Git 候选

状态：部分完成。

已实现：

- 当前仓库本地分支候选。
- 远程跟踪分支候选。
- 当前分支标识。
- 分支查询取消、超时和仓库目录切换。
- `switch`、`checkout`、`merge`、`rebase`、`reset`、`log`、`diff`、`pull` 和 `push` 等上下文中的分支候选路由。

下一批计划：

- remote 名称候选；
- tag 候选；
- 已修改、已删除和未跟踪文件候选；
- 适合当前命令的 revision、commit 和路径候选；
- 缓存、防抖和统一的异步取消机制，避免每次按键都启动 Git 进程。

示例：

```text
git add int
    internal/editor/editor.go
    internal/suggest/engine.go

git remote remove ori
    origin

git tag --delete v1
    v1.0.0
```

验收标准：执行 `cd` 进入不同仓库后，候选立即对应新仓库；慢查询不会阻塞键盘输入；查询失败不会影响 Shell 使用。

### 5. Git 与 Shell token 解析

状态：待扩展。

当前 Parser 只处理空白分隔的简单 token，并支持光标位于 token 中间时替换正确的 rune 范围。引号、转义、Git 全局 option 和 Shell 命令边界仍未完整处理。

下一批场景：

```text
git -C ../repo status
git --no-pager log
cd repo; git status
git status | less
git add "file with spaces.txt"
git commit -m 'fix message'
```

实现时需要区分 PowerShell、Bash 和 Zsh 的引用与转义规则，并识别 `|`、重定向、`;`、`&&` 和 `||` 创建的新命令边界。不计划一次实现完整 Shell Parser，而是按真实补全需求逐步扩展。

验收标准：解析失败时宁可不提示，也不能替换错误文本范围或生成危险命令。

## P2：一般命令行体验

### 6. 行编辑与快捷键

状态：部分完成。

已实现：

- 左右方向键、Home、End 和 Delete；
- Backspace、Enter 和 Tab；
- Ctrl+C 与 Ctrl+D；
- 上下方向键浏览候选或当前会话历史；
- Tab 补全、候选选择和草稿恢复；
- 多行 bracketed paste；
- Unicode rune 编辑和终端单元格宽度计算。

下一批计划：

- Ctrl+L：清屏并重绘；
- Ctrl+A / Ctrl+E：移动到行首或行尾；
- Ctrl+W：删除前一个单词；
- Ctrl+U / Ctrl+K：清除光标前后内容；
- Ctrl+R：搜索历史；
- Escape：退出候选选择；
- 根据补全上下文决定是否自动添加空格。

验收标准：快捷键行为尽量与主流 PowerShell/Bash 行编辑体验一致，并且不会截获前台交互程序的输入。

### 7. 历史持久化

状态：待实现。

当前历史只保存在本次 Gogit 进程内，会避免连续记录相同命令，并能在结束浏览后恢复用户原有草稿。

后续需要：

- 限制内存历史条数；
- 保存到操作系统用户数据目录；
- 启动时加载，执行后或退出前安全写入；
- 避免多个 Gogit 进程相互覆盖；
- 提供关闭持久化的配置；
- 过滤 token、密码和其他敏感命令。

验收标准：重启 Gogit 后可以浏览历史；历史文件损坏或不可写时不影响 Shell 使用。

### 8. Prompt 与显示配置

状态：待实现。

计划支持：

- 自定义颜色和候选数量；
- 显示或隐藏环境名、当前目录和 Gogit 前缀；
- 可选显示上一条命令退出码；
- 小窗口和浅色/深色终端适配；
- 对目录名和环境名中的控制字符进行安全处理；
- 配置文件位置、格式、默认值和错误回退策略。

验收标准：配置错误时使用安全默认值，Prompt 不会破坏终端状态。

## P3：可靠性、测试与发布

### 9. 交互式程序兼容性

状态：部分完成。

已实现：

- 非编辑状态的键盘输入直接透传给前台 PTY 程序，包括 Enter 和控制字符。
- 多行粘贴结束后的命令不会与交互输入混淆。
- Windows PowerShell 的真实 PTY 测试会验证命令包装内容不回显到终端。
- `git diff` 使用的 pager 可以接收输入并通过 `q` 退出。

仍需在真实平台验证：

- Vim 或其他全屏编辑器；
- `less`、`top` 等全屏程序的完整操作；
- Python/Node REPL；
- SSH；
- Ctrl+C、Ctrl+D 和窗口 resize；
- 修改终端 echo/raw mode 的程序；
- PowerShell 5.1、PowerShell 7、Bash、Zsh 和 POSIX sh。

验收标准：交互程序退出后，Gogit Prompt、光标、颜色和 raw mode 均能正确恢复。

### 10. 自动化、端到端与并发测试

状态：进行中。

已实现：

- Editor、Decoder、Renderer、History 和 Suggest 的单元测试；
- 分支解析与动态分支候选测试；
- Prompt Scanner 拆分、损坏、嵌套、超限和恢复测试；
- Shell 配置加载与 Prompt 恢复测试；
- 多行粘贴、交互输入透传和 PowerShell PTY 回归测试；
- GitHub Actions 在 Windows、Linux 和 macOS 上执行 build、vet、test 与 gofmt 检查。

仍需补充：

- Unix 真实 PTY 的完整读写、关闭与子进程回收测试；
- PowerShell 5.1 与 PowerShell 7 的完整矩阵；
- 快速输入、慢命令、Prompt 恢复和 resize 并发场景；
- Shell 意外退出、输出管道关闭和终端恢复失败；
- 更完整的交互程序与多 Shell 端到端测试。

验收标准：正常退出、异常退出和 Ctrl+C 后，宿主终端均不会停留在损坏状态；CI 在全部目标平台稳定通过。

### 11. 安装、版本与发布

状态：部分完成。

已实现：

- README 提供 Windows、macOS 和 Linux 的构建安装与 PATH 配置说明。
- 安装完成后可以在任意目录输入 `gogit` 进入辅助 Shell。
- CI 覆盖 Windows、Linux 和 macOS。
- `v*.*.*` tag 会构建 Windows amd64、Linux amd64 和 macOS arm64 包。
- Release workflow 会生成压缩包、SHA-256 校验文件和 GitHub Release。

仍需完成：

- `gogit --version` 与基础 `--help`；
- 升级和卸载说明；
- 配置文件与用户数据目录约定；
- 确认更多需要发布的架构；
- 创建并验证第一个正式版本 tag；
- 保持 README、开发计划和实际功能同步。

验收标准：新用户可以按照 README 安装、运行、升级和卸载 Gogit，并能验证下载产物的校验值。

## 推荐实施顺序

接下来建议按以下顺序推进：

1. 动态 remote、tag 和文件候选；
2. 支持 Git 全局 option、引号、路径和 Shell 命令边界；
3. 补充 `reset`、`revert`、`cherry-pick`、`worktree`、`submodule` 等目录；
4. 常用编辑快捷键与历史持久化；
5. Prompt/显示配置；
6. 完整端到端测试、版本命令和首次正式发布。

每完成一个阶段，至少执行：

```powershell
gofmt -w <changed-files>
go test ./... -count=1
go vet ./...
git diff --check
```

涉及终端交互时，还必须在真实 PowerShell、Bash 和 Zsh 中进行手动验收，不能只依赖单元测试。
