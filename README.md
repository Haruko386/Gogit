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

### A little story:

> You are an internal in a company, this is your first work since you are an adult.
>
> **First day**, you boss told you to `Clone` the company's project from **GitHub** and get familiar with it.
> You know how to do it, just type:
> `git clone ...`
> 
> **Next day**, you need to submit some `feat` or `fix` **PR** to the upstream branch. Still ok:
> `git checkout -b feat/something`
> 
> **Few days later**, your boos walks over:
> A contributor's **PR** have conflict, you need to review his code and give him a `suggestion` to resolve the conflict.
> 
> Usually, on your local and your own branch, you just type: `git checkout branch` → `git pull upstream main --rebase` → `resolve the conflict in local` and `git push --force-with-lease origin branch`
> 
> But in this way, you found there are **two conflict**, but in GitHub, there should be only **one conflict** ???
> 
> Finally, you [**Boss**](https://github.com/JinHai-CN) told you how to do it **right**:

```shell
git remote add whhe git@github.com:whhe/ragflow
git fetch whhe
git checkout upstream/main
git merge whhe/feat-bedrock-api-key-auth
```

> Like dude, what fk is this bro?

This was basically my experience while interning on **[infiniflow/ragflow](https://github.com/infiniflow/ragflow)**.
And that **PR review** is [**here**](https://github.com/infiniflow/ragflow/pull/18301#pullrequestreview-4992126462)

To be honest, I only know 

```bash
git clone ...
git add .
git commit -m ""
git push # I like to use `--force`, although I know the consequence :)
```

before my internship in [**Infiniflow**](https://github.com/infiniflow)

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

It's stopping you for half a second before you do it.

> [!important]
> If you use **agent** to operate git, bro, this is not what you need.
> 
> Just use your agent and let your brain go to waste.
> 
> (I'm not mean AI is harmful or that I want to ban it, I just think we should not let AI do everything, so that we won't forget something basic skill)
