import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
const main=fs.readFileSync('cmd/opendesk/main.go','utf8');
test('focused workbench has FormGuard plus translator and two OCR tools',()=>{
 const ops=[...html.matchAll(/<button class="operation(?: active)?" data-op="([^"]+)"/g)].map(x=>x[1]);
 assert.deepEqual(ops,['formguard-audit','translate-docx','ocr-images','ocr-pdf']);
 for(const op of ['pdf-merge','pdf-split','pdf-rotate','images-pdf','organize','csv-xlsx','docx-txt','office-pdf']){
  assert.ok(!ops.includes(op),'generic tool still in main menu: '+op);
 }
 assert.doesNotMatch(html,/id="community-hub"|src="\/catalog\.js"|id="tool-search"/);
 assert.doesNotMatch(main,/web\/catalog\.js|GET \/catalog\.js/);
});
test('translation roadmap must never be misrepresented as a working translator',()=>{
 assert.match(html,/data-en="Multilingual document translation — limited DOCX beta"/);
 assert.match(html,/Full-format translation and layout preservation are NOT ready/);
 assert.match(html,/data-en="FormGuard — Document preflight"/);
 assert.match(main,/preservetranslate\.Preflight\(r\.Context\(\),files\)/);
});
