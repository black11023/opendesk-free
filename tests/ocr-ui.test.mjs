import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
const backend=fs.readFileSync('cmd/opendesk/main.go','utf8');
test('bilingual optional OCR UI has status, languages, output modes',()=>{
 for(const op of ['ocr-images','ocr-options','ocr-status','ocr-language','ocr-format','ocr-retry'])assert.ok(html.includes(op));
 assert.match(html,/data-en="OCR scanned images"/);
 assert.match(html,/data-vi="Nhận dạng chữ từ ảnh"/);
 for(const lang of ['eng','vie','eng+vie']) assert.ok(html.includes('value="'+lang+'"'));
 assert.match(html,/value="pdf"/);
 assert.match(html,/value="txt"/);
 assert.match(html,/Tesseract/);
});
test('process sends fixed OCR arguments through authenticated local route',()=>{
 assert.match(html,/form.append\('ocr_lang',state.ocrLang\)/);
 assert.match(html,/form.append\('ocr_format',state.ocrFormat\)/);
 assert.match(html,/fetch\('\/ocr-status'/);
 assert.match(backend,/GET \/ocr-status/);
 assert.match(backend,/if !auth\(w,r\)/);
 assert.match(backend,/ocr.RunImages/);
});
test('OCR is disabled when language pack or engine missing',()=>{
 assert.match(html,/state\.op==='ocr-images'\|\|state\.op==='ocr-pdf'/);
 assert.match(html,/!ocrReady\(\)/);
 assert.match(html,/function ocrReady\(\)/);
 assert.match(html,/state.ocrStatus=null/);
});

test('PDF OCR requires both local OCR engines with page limits and truthful UI',()=>{
 assert.match(html,/data-op="ocr-pdf"/);
 assert.match(html,/data-en="OCR scanned PDF"/);
 assert.match(html,/data-vi="Nhận dạng chữ trong PDF scan"/);
 assert.match(html,/fetch\('\/pdf-ocr-status'/);
 assert.match(html,/state\.pdfOcrStatus\?\.installed/);
 assert.match(html,/form\.append\('ocr_lang',state\.ocrLang\)/);
 assert.match(backend,/GET \/pdf-ocr-status/);
 assert.match(backend,/ocr\.RunPDF/);
});
