import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';
const workflow=fs.readFileSync('.github/workflows/build-windows.yml','utf8');
const gate=fs.readFileSync('scripts/defender-release-scan.ps1','utf8');
test('latest public Windows release requires antivirus scanning of both executables',()=>{
 assert.match(workflow,/Antivirus security gate \(both exact built EXEs\)/);
 assert.match(workflow,/defender-release-scan\.ps1/);
 assert.match(workflow,/OpenDeskFree_Portable_0\.6\.5_preview1\.exe/);
 assert.match(workflow,/OpenDeskFree_Setup_0\.6\.5_preview1\.exe/);
 assert.match(workflow,/AV-Evidence/);
});
test('antivirus missing, disabled or detection fails closed without exclusions',()=>{
 assert.match(gate,/Get-MpComputerStatus/);
 assert.match(gate,/Start-MpScan -ScanType CustomScan -ScanPath/);
 assert.match(gate,/Get-MpThreatDetection/);
 assert.match(gate,/AMRunningMode -ne 'Normal'/);
 assert.match(gate,/PASS_EICAR_DETECTED/);
 assert.match(gate,/Get-MpThreatDetection/);
 assert.match(gate,/INCONCLUSIVE: Defender did not detect the benign EICAR positive control/);
 assert.match(gate,/Get-FileHash/);
 assert.match(gate,/DETECTION_OR_QUARANTINE/);
 assert.match(gate,/status = 'BLOCKED'/);
 assert.doesNotMatch(gate,/Set-MpPreference|Add-MpPreference|DisableRealtimeMonitoring|ExclusionPath|ExecutionPolicy Bypass/);
});
