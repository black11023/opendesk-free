import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
const ctx={window:{}};
vm.runInNewContext(fs.readFileSync('cmd/opendesk/web/catalog.js','utf8'),ctx);
const catalog=ctx.window.OpenDeskCatalog;
const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
test('curated directory has unique HTTPS-only external links',()=>{
 assert.ok(catalog.length>=40,'expected at least 40 projects');
 assert.equal(new Set(catalog.map(x=>x.name.toLowerCase())).size,catalog.length);
 assert.ok(new Set(catalog.map(x=>x.category)).size>=10);
 for(const x of catalog){
  assert.equal(x.status,'external');
  assert.match(x.url,/^https:\/\//);
  assert.ok(x.name&&x.aboutEn&&x.aboutVi&&x.category);
 }
});
test('directory never implies bundled integrations',()=>{
 assert.match(html,/id="community-hub"/);
 assert.match(html,/id="hub-search"/);
 assert.match(html,/id="hub-tabs"/);
 assert.match(html,/id="hub-grid"/);
 assert.match(html,/window\.OpenDeskCatalog/);
 assert.match(html,/rec\.status!=='external'/);
 assert.match(html,/noopener noreferrer/);
 assert.match(html,/Ctrl\+K/);
});
test('native image PDF is offered in both locales',()=>{
 assert.match(html,/data-op="images-pdf"/);
 assert.match(html,/data-en="Images → PDF"/);
 assert.match(html,/data-vi="Ảnh → PDF"/);
 assert.match(html,/if\(op==='ocr-images'\|\|op==='images-pdf'\)return \['\.png','\.jpg','\.jpeg'\]/);
});
