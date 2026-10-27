import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';
const portable=fs.readFileSync('cmd/opendesk/main.go','utf8');
const native=fs.readFileSync('cmd/opendesk/browser_windows.go','utf8');
test('Windows Portable launches localhost via native ShellExecute and reports errors via native MessageBox',()=>{
 assert.doesNotMatch(portable,/powershell\.exe|rundll32\.exe|Add-Type -AssemblyName/);
 assert.match(native,/ShellExecuteW/);
 assert.match(native,/MessageBoxW/);
 assert.match(native,/refusing to launch a non-local URL/);
 assert.doesNotMatch(native,/os\/exec|EncodedCommand|ExecutionPolicy/);
});
