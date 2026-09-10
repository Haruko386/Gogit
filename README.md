<div align="center"><h1><b>Gogit</b></h1></div>

![Gogit](https://socialify.git.ci/Haruko386/Gogit/image?custom_language=Go&description=1&font=Source+Code+Pro&forks=1&issues=1&language=1&name=1&owner=1&pattern=Floating+Cogs&pulls=1&stargazers=1&theme=Auto)

<div align="center">
    <img height="32" src="https://cdn.simpleicons.org/go">
    <img height="32" src="https://cdn.simpleicons.org/git">
    <img height="32" src="https://cdn.simpleicons.org/github">
</div>

<div align="center">

[![Typing SVG](https://readme-typing-svg.demolab.com?font=JetBrains+Mono&size=22&duration=4500&pause=1400&color=728295&center=true&vCenter=true&repeat=false&width=435&lines=Git+is+easy+until+you+use+it+%3A%28)](https://git.io/typing-svg)
</div>

### Imagine this:

You just started your first internship as a developer.

#### Day one:

```bash
git clone ...
```

Easy.

---

#### Day two:

```bash
git checkout -b feat/something
```

Still easy.

---

#### Then one day, your mentor walks over:

> "Main has moved forward. Rebase your commits onto the latest main,  
> resolve the conflicts, and update your remote branch safely."

You:

> "Sure."

Your brain:

> **What the hell is a rebase?**

So you search Google, Stack Overflow, GitHub — or just ask **`ChatGPT`**.

A minute later, you somehow end up with:

```bash
git fetch origin
git rebase origin/main

# resolve conflicts...

git add .
git rebase --continue
git push --force-with-lease
```

It works!!!

For a brief moment, you think you understand Git.

----

#### Until the next day:

> "Remove yesterday's commit, but keep the changes in your working tree."

You stare at the terminal.

```bash
git reset ???
```

`--soft`? `--mixed`? `--hard`?

And somehow you're back to searching:

> **difference between git reset soft mixed hard**

---

This was basically my experience while interning on **[infiniflow/ragflow](https://github.com/infiniflow/ragflow)**.

Git doesn't have a shortage of commands.  
If anything, it has **way too many of them**.

Usually, the problem isn't:

> "I don't know what I want to do."

It's:

> "I know exactly what I want to do.  
> I just don't know what that damn Git command is called."

So I built **Gogit** — because memorizing Git commands shouldn't be part of the job.
At least not for a beginner

---

<div align="center"><h1><b>What is Gogit?</b></h1></div>

**Gogit** is a Git CLI assistant written in Go.

<div align=center>
    <img width="80%" src="./external/gif/demo.gif">
</div>

## Install and run from anywhere

Gogit requires **Go 1.25 or later** and **Git**. Clone this repository, enter
its directory, and install the executable into your Go binary directory.

```bash
git clone https://github.com/Haruko386/Gogit.git
cd Gogit
```

### Windows PowerShell

```powershell
$installDir = go env GOBIN
if ([string]::IsNullOrWhiteSpace($installDir)) {
    $installDir = Join-Path (go env GOPATH) "bin"
}

New-Item -ItemType Directory -Force $installDir | Out-Null
go build -o (Join-Path $installDir "gogit.exe") .

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($userPath -split ";") -notcontains $installDir) {
    $newPath = if ($userPath) { "$userPath;$installDir" } else { $installDir }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
}
$env:Path += ";$installDir"
```

### macOS or Linux

```bash
install_dir="$(go env GOBIN)"
if [ -z "$install_dir" ]; then
    install_dir="$(go env GOPATH)/bin"
fi

mkdir -p "$install_dir"
go build -o "$install_dir/gogit" .
```

Add that directory to your shell's `PATH` if it is not already available:

```bash
# Bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Zsh (the default shell on modern macOS)
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

If you configured a custom `GOBIN`, add that directory instead of
`$(go env GOPATH)/bin`.

Verify the installation from a new terminal or any directory:

```bash
gogit
```

This opens the Gogit-assisted shell. Type Git commands normally and press
`Ctrl+D` on an empty input line when you want to leave Gogit.

It simply tries to help when you're staring at this:

```bash
git branch --sho|
```

and wondering what comes next.

Gogit is planned to suggest:

```text
--show-current
    Print the name of the current branch.
```

Press `Enter`:

```bash
git branch --show-current
```

Done.

No need to:

1. open a browser
2. search the Git documentation
3. open Stack Overflow
4. ask an AI
5. copy an answer written in 2014
6. pray it doesn't delete your working tree

The goal is to turn this:

```text
I know what I want to do
          ↓
       git ...
          ↓
    Gogit helps
          ↓
         Done
```

instead of this:

```text
I know what I want to do
          ↓
       Google
          ↓
   Stack Overflow
          ↓
          AI
          ↓
  git reset --hard
          ↓
    Wait... WHAT?
```

---

## The Idea

The planned autocomplete will understand the current Git command context
and suggest available options while you type.

For example:

```bash
git branch --
```

Gogit may show:

```text
--show-current
    Show the name of the current branch.

--merged
    List branches already merged into the specified commit.

--no-merged
    List branches that have not yet been merged.

--delete
    Delete a branch.
```

So instead of only telling you:

> **what you can type**

Gogit also tells you:

> **what the hell it actually does**

For dangerous commands, Gogit should eventually be able to tell you
that you're about to do something... interesting:

```text
--force
    Force push to the remote repository.

    ⚠ This may overwrite remote history.

--force-with-lease
    Force push only when the remote branch has not unexpectedly changed.

    ✓ Usually safer than --force.
```

Because sometimes the most useful feature of a CLI assistant
is not helping you type:

```bash
git reset --hard
```

faster.

It's stopping you for half a second before you do it.

---

<div align="center"><h1><b>Why "Gogit"?</b></h1></div>

Because it's written in **Go**.

And it's for **Git**.

Go + Git.

**Gogit.**

Yes.

I spent considerably more time Googling Git commands than naming this project.

---



> [!important]
> If you use **agent** to operate git, bro, this is not what you need.
> 
> Just use your agent and let your brain go to waste.
> 
> (I'm not mean AI is harmful or that I want to ban it, I just think we should not let AI do everything, so that we won't forget something basic skill)
