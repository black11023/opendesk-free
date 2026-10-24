// Launches the actual Windows GUI executable, exercises local HTTP routes, and shuts down.
// This is an end-to-end runtime test (not a mocked HTTP server).
import assert from 'node:assert/strict';
import test from 'node:test';
import {spawn} from 'node:child_process';
import net from 'node:net';
import {setTimeout as delay} from 'node:timers/promises';
import path from 'node:path';
const executable=path.resolve('cmd/setup/payload/OpenDeskFree.exe');

async function findAvailablePort(){
 const s=net.createServer();
 await new Promise((resolve,reject)=>{s.once('error',reject);s.listen(0,'127.0.0.1',resolve)});
 const p=s.address().port;
 await new Promise(resolve=>s.close(resolve));
 return p;
}
async function awaitReady(url,child){
 for(let n=0;n<180;n++){
  if(child.exitCode!==null)throw Error('OpenDesk process exited early: '+child.exitCode);
  try {const response=await fetch(url+'/health',{signal:AbortSignal.timeout(1000)});if(response.ok)return await response.text()}catch(_){}
  await delay(125);
 }
 throw Error('OpenDesk never opened its localhost HTTP listener');
}
test('actual Windows executable: localhost, CSRF, processing, shutdown', {timeout:45000}, async()=>{
 const port=await findAvailablePort(),url='http://127.0.0.1:'+port;
 const child=spawn(executable,['--port='+port,'--no-browser'],{windowsHide:true,stdio:'ignore'});
 try{
  const health=await awaitReady(url,child);
  assert.match(health,/OpenDesk Free .* OK/);
  const page=await fetch(url+'/');assert.equal(page.status,200);
  const html=await page.text();assert.match(html,/<html lang="en">/);
  const token=html.match(/const TOKEN='([a-f0-9]{64})'/)?.[1];
  assert.ok(token,'missing per-run CSRF token');
  const forbidden=await fetch(url+'/office-status');assert.equal(forbidden.status,403);
  const origin=await fetch(url+'/ocr-status',{headers:{'X-OpenDesk-Token':token,Origin:'https://other.example'}});assert.equal(origin.status,403);
  const form=new FormData();form.set('operation','join-text');
  form.append('files',new Blob(['Hello world'],{type:'text/plain'}),'hello.txt');
  form.append('files',new Blob(['Vietnam'],{type:'text/plain'}),'vietnam.txt');
  const noAuth=await fetch(url+'/process',{method:'POST',body:form});assert.equal(noAuth.status,403);
  const ok=await fetch(url+'/process',{method:'POST',body:form,headers:{'X-OpenDesk-Token':token,'X-OpenDesk-Lang':'en'}});
  assert.equal(ok.status,200);
  assert.match(ok.headers.get('content-type')||'',/text\/plain/);
  assert.match(await ok.text(),/Hello world[\s\S]*Vietnam/);
  const malformed=new FormData();malformed.set('operation','pdf-split');malformed.append('files',new Blob(['corrupt'],{type:'application/pdf'}),'invalid.pdf');malformed.set('pages','1');
  const invalid=await fetch(url+'/process',{method:'POST',headers:{'X-OpenDesk-Token':token},body:malformed});
  assert.equal(invalid.status,422);
  const quit=await fetch(url+'/quit',{method:'POST',headers:{'X-OpenDesk-Token':token}});assert.equal(quit.status,200);
 }finally{
  child.kill();
 }
});
