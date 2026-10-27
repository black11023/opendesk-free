import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';

const html=fs.readFileSync('cmd/opendesk/web/index.html','utf8');
test('every implemented operation gets concise offline first-use guidance',()=>{
 const ops=[...html.matchAll(/data-op="([^"]+)"/g)].map(m=>m[1]);
 const guideSection=html.match(/const beginnerTips=(\{[\s\S]*?\});\s*function acceptedTypes/);
 assert.ok(guideSection,'guide dictionary missing');
 for(const op of ops)assert.ok(guideSection[1].includes("'"+op+"':"),'missing guide: '+op);
 for(const id of ['beginner-guide','beginner-heading','beginner-tip','file-warning'])assert.ok(html.includes('id="'+id+'"'));
 assert.match(html,/function renderBeginnerGuide\(\)/);
 assert.match(html,/function refresh\(\)\{renderBeginnerGuide\(\)/);
});
test('file extension warning is accessible, blocks unsuitable input but not audits',()=>{
 assert.match(html,/role="alert" hidden/);
 assert.match(html,/state\.processing\|\|!!fileTypeError\(\)/);
 assert.match(html,/state\.op==='pdf-audit'\|\|state\.op==='office-pdf-batch'/);
 assert.match(html,/acceptedTypes\(state\.op\)\.join\(','\)/);
 assert.match(html,/if\(op==='office-pdf'\|\|op==='office-pdf-batch'\)return 'office'/);
});
