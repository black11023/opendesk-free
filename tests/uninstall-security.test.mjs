import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
const installer=fs.readFileSync('cmd/setup/main.go','utf8');
test('uninstall cleanup does not use encoded PowerShell or policy bypass',()=>{
 assert.doesNotMatch(installer,/EncodedCommand|ExecutionPolicy|encoding\/base64|unicode\/utf16/);
 assert.match(installer,/cmd:=exec\.Command\("powershell\.exe","-NoProfile","-NonInteractive","-Command",script\)/);
 assert.match(installer,/Remove-Item -LiteralPath/);
});
