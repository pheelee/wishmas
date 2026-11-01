// @test CON-001-AC5 CON-001-AC6: resolve the reference with the real Compose CLI.
// WISHMAS_COMPOSE_BIN optionally selects a standalone official Compose binary.
const {execFileSync}=require('node:child_process');
const assert=require('node:assert/strict');
const {resolve}=require('node:path');
const command=process.env.WISHMAS_COMPOSE_BIN||'docker';
const prefix=process.env.WISHMAS_COMPOSE_BIN?[]:['compose'];
// @test REL-001-AC2 CON-001-AC6: compare the CLI's own normalized false form.
// Compose 2 serializes false as {}; newer releases preserve explicit false.
// Probe both values so a serializer that cannot distinguish them fails closed.
function normalizedBind(createHostPath){
 const fixture={services:{probe:{image:'scratch',volumes:[{type:'bind',source:'./data',target:'/data',bind:{create_host_path:createHostPath}}]}}};
 const result=execFileSync(command,[...prefix,'--env-file','/dev/null','-f','-','config','--format','json'],{encoding:'utf8',input:JSON.stringify(fixture)});
 return JSON.parse(result).services.probe.volumes[0].bind;
}
const noHostCreation=normalizedBind(false);
assert.notDeepEqual(noHostCreation,normalizedBind(true),'Compose must distinguish disabled host-directory creation');
for(const [uid,gid,expected]of [['','','1000:1000'],['1234','','1234:1000'],['','5678','1000:5678'],['1234','5678','1234:5678']]){
 const result=execFileSync(command,[...prefix,'--env-file','/dev/null','-f','docker-compose.yml','config','--format','json'],{encoding:'utf8',env:{...process.env,WISHMAS_UID:uid,WISHMAS_GID:gid,WISHMAS_TRUSTED_PROXIES:'192.0.2.2/32'}});
 const service=JSON.parse(result).services.wishmas;
 assert.equal(service.user,expected);assert.ok(!service.ports?.length,'Do not publish the backend port');
 assert.deepEqual(service.expose,['8080']);assert.equal(service.environment.WISHMAS_TRUSTED_PROXIES,'192.0.2.2/32');
 assert.equal(service.volumes.length,1);const volume=service.volumes[0];
 assert.equal(volume.type,'bind');assert.equal(volume.source,resolve('data'));assert.equal(volume.target,'/data');assert.deepEqual(volume.bind,noHostCreation,'Host-directory creation must remain disabled');
}
console.log('PASS reference Compose: default and independently overridden UID/GID, persistent bind mount, no automatic host-directory creation, private backend and explicit proxy input.');
