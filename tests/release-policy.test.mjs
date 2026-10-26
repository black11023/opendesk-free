import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';
const windows=fs.readFileSync('.github/workflows/build-windows.yml','utf8');
const release=fs.readFileSync('.github/workflows/release-preview.yml','utf8');
test('Windows build never publishes before independent CI gates',()=>{
 assert.doesNotMatch(windows,/gh release create/);
 assert.match(release,/workflow_run:/);
 for(const name of ['Build and test Windows installer','Real Office and OCR document integration','Chromium UI and memory acceptance'])assert.ok(release.includes(name),name);
 assert.match(release,/status" != success/);
 assert.match(release,/head_sha==\$sha/);
 assert.match(release,/sha256sum --check SHA256SUMS.txt/);
 assert.match(release,/gh release create/);
 assert.match(release,/current" = "\$SOURCE_SHA"/);
});
