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