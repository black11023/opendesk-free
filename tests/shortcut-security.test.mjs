import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';
const installer=fs.readFileSync('cmd/setup/main.go','utf8');
const shortcut=fs.readFileSync('cmd/setup/shortcut_windows.go','utf8');
test('installation creates start menu shortcut through native COM not an encoded shell',()=>{
 assert.match(shortcut,/CoCreateInstance/);
 assert.match(shortcut,/IPersistFile\.Save/);
 assert.doesNotMatch(shortcut,/powershell|WScript|exec\.Command/);
 assert.match(installer,/createNativeShortcut\(app, shortcutPath\(\)\)/);
});
