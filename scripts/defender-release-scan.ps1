#Requires -Version 5.1
<#
Windows-only release security gate. It does not modify Defender exclusions,
disable protection, or execute the candidate EXE.
Any missing/disabled antivirus, stale signatures, scan error or detection
must FAIL. PASS is limited to this Defender signature scan, not a claim
that Chrome Safe Browsing or all AV vendors will approve the download.
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)][string[]]$Paths,
    [Parameter(Mandatory = $true)][string]$ReportPath
)
$ErrorActionPreference = 'Stop'
$report = [ordered]@{
    gate = 'Microsoft Defender Antivirus release scan'
    status = 'BLOCKED'
    scanned_at_utc = (Get-Date).ToUniversalTime().ToString('o')
    engine = $null
    signatures = $null
    files = @()
    reason = $null
}
function Save-Report {
    $dir = Split-Path -Parent $ReportPath
    if ($dir) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    $report | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $ReportPath -Encoding UTF8
}
try {
    if ($Paths.Count -ne 2) { throw 'Expected exactly the built Portable and Setup binaries' }
    $targets = @()
    foreach ($path in $Paths) {
        $item = Get-Item -LiteralPath $path -ErrorAction Stop
        if ($item.Extension -ne '.exe' -or $item.Length -lt 100000) {
            throw "Invalid executable artifact: $path"
        }
        $targets += $item
    }
    $status = Get-MpComputerStatus -ErrorAction Stop
    if (-not $status) { throw 'Defender status is unavailable' }
    $report.engine = [string]$status.AMEngineVersion
    $report.signatures = [string]$status.AntivirusSignatureVersion
    if ($status.AMRunningMode -ne 'Normal' -or
        $status.AntivirusEnabled -ne $true -or
        $status.RealTimeProtectionEnabled -ne $true) {
        throw "INCONCLUSIVE: Defender is not active (mode=$($status.AMRunningMode), antivirus=$($status.AntivirusEnabled), realtime=$($status.RealTimeProtectionEnabled))"
    }
    if ($status.DefenderSignaturesOutOfDate -eq $true) {
        throw 'INCONCLUSIVE: Defender signatures are outdated'
    }
    $startTime = Get-Date
    foreach ($item in $targets) {
        $before = (Get-FileHash -LiteralPath $item.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
        $entry = [ordered]@{ file = $item.Name; bytes = $item.Length; sha256 = $before; scan = 'PENDING' }
        $report.files += $entry
        Start-MpScan -ScanType CustomScan -ScanPath $item.FullName -ErrorAction Stop
        if (-not (Test-Path -LiteralPath $item.FullName -PathType Leaf)) {
            throw "DETECTION_OR_QUARANTINE: file missing after antivirus scan: $($item.Name)"
        }
        $after = (Get-FileHash -LiteralPath $item.FullName -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($after -ne $before) { throw "File hash changed after Defender scan: $($item.Name)" }
        $entry.scan = 'COMPLETED'
    }
    $alerts = @(Get-MpThreatDetection -ErrorAction Stop | Where-Object {
        $_.InitialDetectionTime -and $_.InitialDetectionTime -ge $startTime
    })
    if ($alerts.Count -gt 0) {
        throw "DETECTED: Defender registered $($alerts.Count) new threat event(s) during the scans"
    }
    $report.status = 'PASS'
    $report.reason = 'Both artifacts scanned by active Microsoft Defender; no new detection seen'
    Save-Report
    Write-Host "Microsoft Defender release gate PASS, engine $($report.engine), signatures $($report.signatures)"
    exit 0
}
catch {
    $report.status = 'BLOCKED'
    $report.reason = [string]$_.Exception.Message
    Save-Report
    Write-Error "Antivirus release gate BLOCKED: $($report.reason)"
    exit 1
}
