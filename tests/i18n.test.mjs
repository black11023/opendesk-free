import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import vm from 'node:vm';

const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
const match=html.match(/const t=(\{.*?\});(?=\s*(?:const toolFilterCategories|function render))/s);
if(!match)throw new Error('UI translation dictionary missing');
const translations=vm.runInNewContext('('+match[1]+')');
test('English and Vietnamese include the same keys',()=>{
 const en=Object.keys(translations.en).sort();
 const vi=Object.keys(translations.vi).sort();
 assert.deepEqual(en,vi);
 for(const lang of ['en','vi'])for(const key of en)
  assert.ok(translations[lang][key].trim(),lang+':'+key);
});
test('locale selector exists, English is the default, and selection is saved',()=>{
 assert.match(html,/<html lang="en">/);
 assert.match(html,/<select class="langbtn" id="lang"/);
 assert.match(html,/<option value="en">English<\/option>/);
 assert.match(html,/<option value="vi">Tiếng Việt<\/option>/);
 assert.match(html,/localStorage\.getItem\('opendesk-language'\)/);
 assert.match(html,/localStorage\.setItem\('opendesk-language'/);
 assert.match(html,/document\.cookie/);
 assert.match(html,/opendesk-free-language/);
 assert.match(html,/\?'vi':'en'/);
});
test('every inline operation label supplies both languages',()=>{
 const tagged=[...html.matchAll(/data-vi="([^"]*)"\s+data-en="([^"]*)"/g)];
 assert.ok(tagged.length>=16,'expected operation labels, descriptions and PDF tools');
 for(const m of tagged)assert.ok(m[1]&&m[2]);
 for(const op of ['pdf-merge','pdf-split','pdf-rotate']) assert.ok(html.includes('data-op="'+op+'"'));
});

test('PDF page-range controls and server form fields are wired in both languages',()=>{
 for(const id of ['pdf-options','pdf-heading','pdf-controls'])assert.ok(html.includes('id="'+id+'"'));
 for(const prop of ['pdfRange','pdfSplitHelp','pdfMergeHelp','pdfCombined','pdfSeparate','pdfMoveUp','pdfMoveDown']){
  assert.ok(prop in translations.vi&&prop in translations.en,prop);
 }
 assert.match(html,/form\.append\('pages',state\.pageText\)/);
 assert.match(html,/form\.append\('page_ranges',state\.ranges\.get\(f\)\|\|''\)/);
 assert.match(html,/state\.files\[i\].*state\.files\[next\]/);
 assert.match(html,/'pdf-extract'/);
});

test('Office conversion UI requires detected local LibreOffice',()=>{
 assert.match(html,/data-op="office-pdf"/);
 assert.match(html,/data-en="Office → PDF"/);
 assert.match(html,/data-vi="Văn phòng → PDF"/);
 assert.match(html,/id="office-options"/);
 assert.match(html,/fetch\('\/office-status'/);
 assert.match(html,/state\.officeStatus\?\.installed/);
});

test('professional tool search, categories and batch-PDF interface are wired',()=>{
 for(const id of ['tool-search','workbench-tabs','tool-counter','workbench-privacy'])assert.match(html,new RegExp('id="'+id+'"'));
 assert.match(html,/function renderToolFilters\(\)/);
 assert.match(html,/data-op="pdf-rotate"/);
 assert.match(html,/Rotate PDFs 90° \(batch\)/);
 assert.match(html,/state\.processing/);
 assert.match(html,/aria-busy/);
});

test('read-only PDF batch audit is labeled in both languages and permits multiple sources',()=>{
 assert.match(html,/data-op="pdf-audit"/);
 assert.match(html,/data-en="Audit PDF files → CSV"/);
 assert.match(html,/data-vi="Kiểm tra PDF hàng loạt → CSV"/);
 assert.match(html,/state\.op!=='pdf-audit'/);
});

test('resilient Office batch controls are bilingual and guarded by local converter',()=>{
 assert.match(html,/data-op="office-pdf-batch"/);
 assert.match(html,/data-en="Office batch → PDF \+ report"/);
 assert.match(html,/data-vi="Office hàng loạt → PDF \+ báo cáo"/);
 assert.match(html,/state\.op==='office-pdf-batch'/);
});
