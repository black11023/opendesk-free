
import assert from 'node:assert/strict';
import {spawn} from 'node:child_process';
import {readFile,writeFile,mkdir} from 'node:fs/promises';
import {setTimeout as sleep} from 'node:timers/promises';
import {chromium} from 'playwright';
import net from 'node:net';
import path from 'node:path';

// Real Chromium clicks and downloads against the compiled OpenDesk binary.
// Screenshots are generated CI artifacts, not screenshots of a consumer Windows PC.
const executable=process.env.OPEN_DESK_EXECUTABLE;
assert.ok(executable,'OPEN_DESK_EXECUTABLE must point to the compiled local binary');
const artifactDir=path.resolve('artifacts/browser');
await mkdir(artifactDir,{recursive:true});
const source=await readFile('tests/fixtures/pdf_two_pages.pdf');
async function freePort(){
 const srv=net.createServer();
 await new Promise((resolve,reject)=>{srv.once('error',reject);srv.listen(0,'127.0.0.1',resolve)});
 const p=srv.address().port;await new Promise(resolve=>srv.close(resolve));return p;
}
async function ready(url,proc){
 for(let i=0;i<160;i++){
  if(proc.exitCode!==null)throw Error('OpenDesk exited '+proc.exitCode);
  try{const response=await fetch(url+'/health',{signal:AbortSignal.timeout(700)});if(response.ok)return}
  catch(_){}
  await sleep(150);
 }
 throw Error('No OpenDesk listener after 24 seconds');
}
async function rssMiB(pid){
 if(process.platform!=='linux')return null;
 try{
  const status=await readFile('/proc/'+pid+'/status','utf8');
  const match=status.match(/^VmRSS:\s+(\d+)\s+kB/m);
  return match?Number(match[1])/1024:null;
 }catch(_){return null}
}
const port=await freePort(),url='http://127.0.0.1:'+port;
const app=spawn(executable,['--port='+port,'--no-browser'],{stdio:['ignore','pipe','pipe']});
let diagnostics='';
for(const stream of [app.stdout,app.stderr])stream.on('data',x=>{diagnostics=(diagnostics+String(x)).slice(-14000)});
let browser;
let screenshot='Not captured';
try{
 await ready(url,app);
 browser=await chromium.launch({headless:true,args:['--no-sandbox']});
 const page=await browser.newPage({acceptDownloads:true,viewport:{width:1365,height:900},locale:'en-US'});
 const errors=[];
 page.on('pageerror',e=>errors.push(e.message));
 const response=await page.goto(url+'/',{waitUntil:'networkidle'});
 assert.equal(response.status(),200);
 assert.equal(await page.locator('#lang').inputValue(),'en','new users should start in English');
 await page.locator('#tool-search').fill('rotate');
 assert.equal(await page.locator('.operation:visible').count(),1,'search should narrow operations');
 assert.equal(await page.locator('.operation[data-op="pdf-rotate"]').isVisible(),true);
 await page.locator('#tool-search').fill('');
 await page.locator('#workbench-tabs button').filter({hasText:'PDF'}).first().click();
 assert.equal(await page.locator('.operation[data-op="pdf-merge"]').isVisible(),true);
 assert.equal(await page.locator('.operation[data-op="office-pdf"]').isVisible(),false);
 await page.locator('#workbench-tabs button').first().click();
 await page.screenshot({path:path.join(artifactDir,'english-workbench.png'),fullPage:true});
 await page.locator('#lang').selectOption('vi');
 assert.match(await page.locator('#headline').innerText(),/Không mất phí/);
 await page.locator('#tool-search').fill('xoay');
 assert.equal(await page.locator('.operation[data-op="pdf-rotate"]').isVisible(),true,'Vietnamese search');
 await page.screenshot({path:path.join(artifactDir,'vietnamese-workbench.png'),fullPage:true});
 await page.locator('#lang').selectOption('en');
 await page.locator('#tool-search').fill('');
 await page.locator('.operation[data-op="pdf-split"]').click();
 await page.locator('#files').setInputFiles({name:'two-pages.pdf',mimeType:'application/pdf',buffer:source});
 await page.locator('#pdf-pages').fill('2');
 assert.equal(await page.locator('#process').isEnabled(),true,'valid PDF selection must be actionable');
 const selectionDownload=page.waitForEvent('download',{timeout:45000});
 await page.locator('#process').click();
 const download=await selectionDownload;
 assert.equal(download.suggestedFilename(),'OpenDesk_Selected_Pages.pdf');
 const pageFile=await readFile(await download.path());
 assert.ok(pageFile.subarray(0,5).toString()==='%PDF-','selected-page result not PDF');
 assert.match(await page.locator('#status').innerText(),/Result created/);
 await page.locator('.operation[data-op="pdf-rotate"]').click();
 await page.locator('#files').setInputFiles([
  {name:'first.pdf',mimeType:'application/pdf',buffer:source},
  {name:'second.pdf',mimeType:'application/pdf',buffer:source}
 ]);
 const batchDownload=page.waitForEvent('download',{timeout:45000});
 await page.locator('#process').click();
 const batch=await batchDownload;
 assert.equal(batch.suggestedFilename(),'OpenDesk_Rotated_PDFs.zip');
 const zipped=await readFile(await batch.path());
 assert.equal(zipped.subarray(0,2).toString(),'PK','batch result not ZIP');
 assert.deepEqual(errors,[],'browser console script errors');
 const rss=await rssMiB(app.pid);
 const report={browser:'Chromium Linux',pdfTestBytes:source.length,batchZipBytes:zipped.length,appRssMiB:rss,untrustedOSLimitMiB:512,screenshots:['english-workbench.png','vietnamese-workbench.png']};
 if(rss!==null)assert.ok(rss<512,'unusually high RSS for tiny real PDF fixtures: '+rss.toFixed(1)+' MiB');
 await writeFile(path.join(artifactDir,'browser-memory.json'),JSON.stringify(report,null,2));
 console.log(JSON.stringify(report,null,2));
 screenshot='Captured';
 await browser.close();browser=null;
 const token=(await (await fetch(url)).text()).match(/const TOKEN='([a-f0-9]{64})'/)?.[1];
 if(token)await fetch(url+'/quit',{method:'POST',headers:{'X-OpenDesk-Token':token}}).catch(()=>{});
}catch(e){
 console.error('OpenDesk browser acceptance failed. App log:',diagnostics);
 console.error('Screenshots:',screenshot);
 throw e;
}finally{if(browser)await browser.close().catch(()=>{});app.kill()}
