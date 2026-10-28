import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
test('every visible tool has bilingual guidance and only supported input extensions',()=>{
 const ops=[...html.matchAll(/data-op="([^"]+)"/g)].map(m=>m[1]);
 assert.deepEqual(ops,['formguard-audit','ocr-images','ocr-pdf']);
 for(const op of ops)assert.ok(html.includes("'"+op+"':["),'missing guidance '+op);
 for(const id of ['beginner-guide','beginner-heading','beginner-tip','file-warning'])assert.ok(html.includes('id="'+id+'"'));
 assert.match(html,/if\(op==='formguard-audit'\)return \['\.docx','\.xlsx','\.pptx'\]/);
 assert.match(html,/state\.op==='formguard-audit'&&state\.files\.length>8/);
 assert.match(html,/!!fileTypeError\(\)/);
 assert.match(html,/role="alert" hidden/);
});
