import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import vm from 'node:vm';

const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
const match=html.match(/const t=(\{.*?\});\s*function render/s);
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
