// End-to-end Windows Setup acceptance: install, upgrade, launch, uninstall.
// Runs only on an ephemeral GitHub-hosted Windows runner. Uses isolated app
// data directories; registration is confined to HKCU of the runner user.
import assert from 'node:assert/strict';
import test from 'node:test';
import {spawn,spawnSync} from 'node:child_process';
import {mkdtemp,mkdir,writeFile,readFile,stat,rm,access} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import path from 'node:path';
import net from 'node:net';
import {setTimeout as delay} from 'node:timers/promises';
const setup=path.resolve('dist/OpenDeskFree_Setup_0.6.5_preview1.exe');
const portable=path.resolve('cmd/setup/payload/OpenDeskFree.exe');
const regKey='HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\OpenDeskFree';
const missing=async p=>{try{await access(p);return false}catch(e){if(e.code==='ENOENT')return true;throw e}};
function execWithEnv(exe,args,env){
 const r=spawnSync(exe,args,{env,encoding:'utf8',windowsHide:true,timeout:30000});
 if(r.error)throw r.error;
 assert.equal(r.status,0,exe+' '+args.join(' ')+' failed with '+r.status+' '+r.stderr);
 return r;
}
async function port(){
 const s=net.createServer();
 await new Promise((yes,no)=>{s.once('error',no);s.listen(0,'127.0.0.1',yes)});
 const n=s.address().port;await new Promise(resolve=>s.close(resolve));return n;
}
async function poll(testCondition,limitMs=12000){
 const stop=Date.now()+limitMs;
 while(Date.now()<stop){if(await testCondition())return;await delay(150)}
 throw Error('condition not reached in '+limitMs+'ms');
}
test('actual Setup EXE: install, upgrade, preserve data, run, uninstall',{timeout:80000},async()=>{
 assert.equal(process.platform,'win32','Windows-only acceptance');
 assert.equal(process.env.GITHUB_ACTIONS,'true','must use isolated CI Windows runner');
 const root=await mkdtemp(path.join(tmpdir(),'OpenDesk Setup Acceptance '));
 const local=path.join(root,'Local AppData');const roaming=path.join(root,'Roaming');
 const outputs=path.join(root,'User Documents');
 await Promise.all([mkdir(local,{recursive:true}),mkdir(roaming,{recursive:true}),mkdir(outputs,{recursive:true})]);
 const env={...process.env,LOCALAPPDATA:local,APPDATA:roaming};
 const installed=path.join(local,'Programs','OpenDeskFree');
 const app=path.join(installed,'OpenDeskFree.exe');
 const uninstaller=path.join(installed,'OpenDeskFree_Uninstall.exe');
 const shortcut=path.join(roaming,'Microsoft','Windows','Start Menu','Programs','OpenDesk Free.lnk');
 const output=path.join(outputs,'valuable-user-document.txt');
 const retained=path.join(installed,'MY_NOTES_DO_NOT_REMOVE.txt');
 let child;
 try{
  await writeFile(output,'user document must survive','utf8');
  execWithEnv(setup,['--quiet-install'],env);
  for(const file of [app,uninstaller,shortcut])assert.equal(await missing(file),false,'missing: '+file);
  const info=spawnSync('reg.exe',['query',regKey,'/v','DisplayVersion'],{encoding:'utf8',windowsHide:true});
  assert.equal(info.status,0,'per-user Installed Apps registration missing');
  assert.match(info.stdout,/0\.6\.5-preview\.1/);
  await writeFile(retained,'keep my data','utf8');
  execWithEnv(setup,['--quiet-install'],env); // actual installer upgrade over existing exe
  assert.equal((await readFile(retained,'utf8')),'keep my data');
  const original=await readFile(portable),after=await readFile(app);
  assert.deepEqual(after,original,'installed app differs from built EXE');
  const listen=await port(),url='http://127.0.0.1:'+listen;
  child=spawn(app,['--port='+listen,'--no-browser'],{env,windowsHide:true,stdio:'ignore'});
  await poll(async()=>{if(child.exitCode!==null)throw Error('installed EXE failed to launch');
    try{return (await fetch(url+'/health',{signal:AbortSignal.timeout(750)})).ok}catch(_){return false}
  });
  const page=await (await fetch(url+'/')).text();
  const token=page.match(/const TOKEN='([a-f0-9]{64})'/)?.[1];assert.ok(token);
  assert.equal((await fetch(url+'/quit',{method:'POST',headers:{'X-OpenDesk-Token':token}})).status,200);
  await poll(async()=>child.exitCode!==null);
  child=null;
  execWithEnv(uninstaller,['--quiet-uninstall'],env);
  await poll(()=>missing(uninstaller));
  assert.equal(await missing(app),true,'installed app was not deleted');
  assert.equal(await missing(shortcut),true,'Start Menu shortcut remains');
  const check=spawnSync('reg.exe',['query',regKey],{encoding:'utf8',windowsHide:true});
  assert.notEqual(check.status,0,'Windows Installed Apps entry remains');
  assert.equal(await readFile(retained,'utf8'),'keep my data');
  assert.equal(await readFile(output,'utf8'),'user document must survive');
 }finally{
  if(child){child.kill();await delay(400)}
  try{execWithEnv(uninstaller,['--quiet-uninstall'],env)}catch(_){}
  await rm(root,{recursive:true,force:true,maxRetries:10,retryDelay:200});
  spawnSync('reg.exe',['delete',regKey,'/f'],{windowsHide:true});
 }
});
