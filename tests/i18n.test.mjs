import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import vm from 'node:vm';
const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
const match=html.match(/const t=(\{.*?\});(?=\s*function render)/s);
if(!match)throw new Error('Bilingual UI dictionary missing');
const translation=vm.runInNewContext('('+match[1]+')');
test('English/Vietnamese dictionaries have identical non-empty keys',()=>{
 const en=Object.keys(translation.en).sort(),vi=Object.keys(translation.vi).sort();
 assert.deepEqual(en,vi);
 for(const lang of ['en','vi'])for(const key of en)assert.ok(translation[lang][key].trim(),lang+':'+key);
});
test('locale persists, readable controls retain two translations',()=>{
 assert.match(html,/<html lang="en">/);
 assert.match(html,/<select class="langbtn" id="lang"/);
 assert.match(html,/<option value="en">English<\/option>/);
 assert.match(html,/<option value="vi">Tiếng Việt<\/option>/);
 assert.match(html,/localStorage\.getItem\('opendesk-language'\)/);
 assert.match(html,/localStorage\.setItem\('opendesk-language'/);
 assert.match(html,/opendesk-free-language/);
 const all=[...html.matchAll(/data-(?:en|vi)="[^"]+"/g)];
 assert.ok(all.length>=14);
 for(const op of ['formguard-audit','translate-docx','translate-xlsx','ocr-images','ocr-pdf'])assert.match(html,new RegExp('data-op="'+op+'"'));
 assert.match(html,/data-vi="FormGuard — Kiểm tra biểu mẫu"/);
 assert.match(html,/data-en="FormGuard — Document preflight"/);
});
test('only available workbench functions are actionable; old PDF editor is not visible',()=>{
 assert.doesNotMatch(html,/data-op="(?:pdf-merge|pdf-split|pdf-rotate|images-pdf|office-pdf)"/);
 assert.match(html,/id="workbench-privacy"/);
 assert.match(html,/state\.processing/);
 assert.match(html,/aria-busy/);
});
test('cancellation is visible in both languages',()=>{
 assert.match(html,/id="cancel-process"/);
 assert.match(html,/data-en="Cancel processing"/);
 assert.match(html,/data-vi="Hủy xử lý"/);
 assert.match(html,/new AbortController\(\)/);
 assert.match(html,/signal:controller\.signal/);
 assert.match(html,/state\.activeRequest\.abort\(\)/);
});
