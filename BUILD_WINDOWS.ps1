# OpenDesk Free local Windows build script (PowerShell)
# Prerequisite: Go 1.26+ installed and on PATH. No administrator privileges required.
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot
New-Item -ItemType Directory -Path 'dist' -Force | Out-Null
Write-Host 'Resolving PDF dependencies ...'
go mod tidy
if ($LASTEXITCODE -ne 0) { throw 'Dependency download failed' }
Write-Host 'Building desktop application ...'
go build -trimpath -ldflags '-s -w -H=windowsgui' -o 'cmd/setup/payload/OpenDeskFree.exe' './cmd/opendesk'
if ($LASTEXITCODE -ne 0) { throw 'Desktop build failed' }
Write-Host 'Running automated tests ...'
go test ./...
if ($LASTEXITCODE -ne 0) { throw 'Automated tests failed' }
Copy-Item 'cmd/setup/payload/OpenDeskFree.exe' 'dist/OpenDeskFree_Portable_0.6.2_preview1.exe' -Force
Write-Host 'Building installer ...'
go build -trimpath -ldflags '-s -w -H=windowsgui' -o 'dist/OpenDeskFree_Setup_0.6.2_preview1.exe' './cmd/setup'
if ($LASTEXITCODE -ne 0) { throw 'Installer build failed' }
Get-ChildItem 'dist/*.exe' | ForEach-Object { "$((Get-FileHash $_.FullName -Algorithm SHA256).Hash)  $($_.Name)" } | Out-File 'dist/SHA256SUMS.txt' -Encoding utf8
Write-Host 'Windows build complete: dist\OpenDeskFree_Setup_0.6.2_preview1.exe'
