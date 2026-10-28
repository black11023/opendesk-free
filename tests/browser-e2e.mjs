import assert from 'node:assert/strict';
import {spawn} from 'node:child_process';
import {readFile,writeFile,mkdir} from 'node:fs/promises';
import {setTimeout as sleep} from 'node:timers/promises';
import {chromium} from 'playwright';
import net from 'node:net';
import path from 'node:path';
const binary=process.env.OPEN_DESK_EXECUTABLE;
assert.ok(binary,'OPEN_DESK_EXECUTABLE required');
const folder=path.resolve('artifacts/browser');await mkdir(folder,{recursive:true});
async function freePort(){const s=net.createServer();await new Promise((a,b)=>{s.once('error',b);s.listen(0,'127.0.0.1',a)});const n=s.address().port;await new Promise(r=>s.close(r));return n}
async function ready(base,proc){for(let i=0;i<150;i++){if(proc.exitCode!==null)throw Error('App exited');try{const r=await fetch(base+'/health',{signal:AbortSignal.timeout(700)});if(r.ok)return}catch(_){}await sleep(180)}throw Error('App did not start')}
async function rssMiB(pid){if(process.platform!=='linux')return null;try{const s=await readFile('/proc/'+pid+'/status','utf8');const m=s.match(/^VmRSS:\s+(\d+)\s+kB/m);return m?Number(m[1])/1024:null}catch(_){return null}}
const url='http://127.0.0.1:'+(await freePort());
const proc=spawn(binary,['--port='+url.split(':').pop(),'--no-browser'],{stdio:['ignore','pipe','pipe']});
let browser;let diagnostics='';
for(const stream of [proc.stdout,proc.stderr])stream.on('data',d=>diagnostics=(diagnostics+d).slice(-12000));
try{
 await ready(url,proc);
 browser=await chromium.launch({headless:true,args:['--no-sandbox']});
 const page=await browser.newPage({acceptDownloads:true,viewport:{width:1280,height:840}});
 const errors=[];page.on('pageerror',e=>errors.push(e.message));
 assert.equal((await page.goto(url+'/',{waitUntil:'networkidle'})).status(),200);
 assert.equal(await page.locator('#lang').inputValue(),'en');
 assert.equal(await page.locator('.operation').count(),5,'FormGuard, DOCX/XLSX translation and OCR visible');
 assert.equal(await page.locator('#community-hub').count(),0);
 assert.equal(await page.locator('.operation[data-op="formguard-audit"]').count(),1);
 assert.match(await page.locator('.suite-banner').innerText(),/NOT ready/);
 await page.screenshot({path:path.join(folder,'focused-en.png'),fullPage:true});
 await page.locator('#files').setInputFiles({name:'bad.pdf',mimeType:'application/pdf',buffer:await readFile('tests/fixtures/pdf_two_pages.pdf')});
 assert.equal(await page.locator('#process').isDisabled(),true);
 assert.match(await page.locator('#file-warning').innerText(),/Wrong file type/);
 const docx=await readFile('tests/fixtures/sample.docx');
 await page.locator('#files').setInputFiles({name:'source.docx',mimeType:'application/vnd.openxmlformats-officedocument.wordprocessingml.document',buffer:docx});
 assert.equal(await page.locator('#file-warning').isHidden(),true);
 assert.equal(await page.locator('#process').isEnabled(),true);
 const event=page.waitForEvent('download',{timeout:45000});await page.locator('#process').click();
 const download=await event;
 assert.equal(download.suggestedFilename(),'OpenDesk_FormGuard_Preflight.csv');
 const contents=(await readFile(await download.path())).toString('utf8');
 assert.match(contents,/Source file,Format,SHA-256 original/);
 assert.match(contents,/source\.docx,DOCX,[0-9a-f]{64}/);
 assert.match(contents,/(READY|REVIEW|BLOCKED)/);
 assert.match(await page.locator('#status').innerText(),/Result created/);
 await page.locator('#lang').selectOption('vi');
 assert.match(await page.locator('#headline').innerText(),/Tài liệu đa ngôn ngữ/);
 assert.match(await page.locator('#beginner-tip').innerText(),/KHÔNG phải chức năng dịch/);
 await page.screenshot({path:path.join(folder,'focused-vi.png'),fullPage:true});
 await page.locator('.operation[data-op="ocr-images"]').click();
 assert.equal(await page.locator('#ocr-options').isVisible(),true);
 await page.locator('.operation[data-op="translate-docx"]').click();
 assert.equal(await page.locator('#translate-options').isVisible(),true);
 await page.waitForFunction(()=>document.querySelector('#model-pack-details').textContent.includes('SHA-256:'),null,{timeout:12000});
 assert.equal(await page.locator('#model-pack-select option').count(),1);
 assert.match(await page.locator('#model-pack-details').innerText(),/SHA-256:/);
 assert.equal(await page.locator('#model-download').isDisabled(),true,'absent inference engine must block installation');
 assert.equal(await page.locator('#model-install-file').isDisabled(),true);
 assert.equal(await page.locator('#model-cancel').isHidden(),true);

 await page.locator('#translate-status').waitFor({state:'visible'});
 assert.equal(await page.locator('#process').isDisabled(),true,'translator must fail closed when model not installed');
 await page.locator('.operation[data-op="translate-xlsx"]').click();
 assert.equal(await page.locator('#translate-options').isVisible(),true,'XLSX uses shared offline model controls');
 assert.equal(await page.locator('#process').isDisabled(),true,'Excel translation must fail closed without local model');
 assert.match(await page.locator('#translate-heading').innerText(),/Excel/);
 assert.match(await page.locator('#translate-details').innerText(),/công thức|formulas/);

 await page.locator('#lang').selectOption('en');
 await page.locator('.operation[data-op="formguard-audit"]').click();
 await page.locator('#files').setInputFiles({name:'cancel.docx',mimeType:'application/vnd.openxmlformats-officedocument.wordprocessingml.document',buffer:docx});
 await page.route('**/process',route=>{/* wait for cancellation */});
 await page.locator('#process').click();
 await page.locator('#cancel-process').waitFor({state:'visible'});
 await page.locator('#cancel-process').click();
 await page.waitForFunction(()=>document.querySelector('#status').textContent.includes('cancelled'),null,{timeout:8000});
 assert.equal(await page.locator('#process').isEnabled(),true);
 await page.unroute('**/process');
 assert.deepEqual(errors,[]);
 const rss=await rssMiB(proc.pid);if(rss!==null)assert.ok(rss<512,'Excess RSS '+rss);
 await writeFile(path.join(folder,'focused-memory.json'),JSON.stringify({rssMiB:rss,visibleTools:5,docxBytes:docx.length},null,2));
 await browser.close();browser=null;
 const html=await (await fetch(url+'/')).text(),token=html.match(/const TOKEN='([a-f0-9]{64})'/)?.[1];
 if(token)await fetch(url+'/quit',{method:'POST',headers:{'X-OpenDesk-Token':token}}).catch(()=>{});
}catch(e){console.error('Real Chromium focus acceptance failed:',e,'runtime:',diagnostics);throw e}
finally{if(browser)await browser.close().catch(()=>{});proc.kill()}
