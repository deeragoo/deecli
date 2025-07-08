# PowerShell install script for deecli

$installDir = "$env:USERPROFILE\.deecli\bin"

if (-Not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

# Assume deecli.exe is in current directory
Copy-Item -Path .\deecli.exe -Destination $installDir -Force

# Add to PATH in PowerShell profile
$profilePath = $PROFILE

if (-Not (Test-Path $profilePath)) {
    New-Item -ItemType File -Path $profilePath -Force | Out-Null
}

$pathEntry = "`$env:USERPROFILE\.deecli\bin"

$content = Get-Content $profilePath -ErrorAction SilentlyContinue
if ($content -notcontains $pathEntry) {
    Add-Content -Path $profilePath -Value "`n# Add deecli to PATH`n`$env:PATH += ';' + $pathEntry"
    Write-Host "Added $pathEntry to PATH in PowerShell profile."
} else {
    Write-Host "$pathEntry already in PATH in PowerShell profile."
}

Write-Host "Installation complete. Please restart PowerShell or run:`n. $profilePath"