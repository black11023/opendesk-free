import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';
const w=fs.readFileSync('.github/workflows/build-windows.yml','utf8');
test('Windows release requires real ClamAV with verified EICAR positive-control',()=>{
 assert.match(w,/malware-audit:/);
 assert.match(w,/needs: windows-release/);
 assert.match(w,/sudo freshclam/);
 assert.match(w,/clamscan --no-summary/);
 assert.match(w,/eicar\.com/);
 assert.match(w,/grep -q 'FOUND'/);
 assert.match(w,/sha256sum --check SHA256SUMS\.txt/);
 assert.match(w,/OpenDeskFree_Portable_0\.7\.0_preview1\.exe/);
 assert.match(w,/OpenDeskFree_Setup_0\.7\.0_preview1\.exe/);
 assert.match(w,/Upload audit logs even when malware scan blocks release/);
 assert.doesNotMatch(w,/-s -w -H=windowsgui/);
});
