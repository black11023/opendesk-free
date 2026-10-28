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

test('multiple runs for exact SHA use newest one; older success cannot mask newer failure',()=>{
 assert.match(release,/sort_by\(\.created_at, \.run_number, \.id\) \| last \/\/ null/);
 assert.match(release,/\$state" != completed/);
 assert.match(release,/\$status" != success/);
 assert.match(release,/\$count" -lt 1/);
 assert.doesNotMatch(release,/\$count" -ne 1/);
});

test('screenshot refresh never bypasses release CI',()=>{
 const screenshot=fs.readFileSync('.github/workflows/refresh-readme-preview.yml','utf8');
 assert.match(screenshot,/git commit -m "docs\(screenshot\): refresh verified English Chromium image"/);
 assert.doesNotMatch(screenshot,/git commit -m[^\n]*\[skip ci\]/);
 assert.match(screenshot,/git diff --cached --quiet/);
});

test('screenshot workflow never directly advances main from workflow_run',()=>{
 const flow=fs.readFileSync('.github/workflows/refresh-readme-preview.yml','utf8');
 assert.doesNotMatch(flow,/git push origin HEAD:main/);
 assert.match(flow,/gh pr create -R "\$REPO" --base main/);
 assert.match(flow,/git diff --cached --quiet/);
});
