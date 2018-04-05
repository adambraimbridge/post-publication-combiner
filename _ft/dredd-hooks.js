var hooks = require('hooks');
var http = require('http');
var fs = require('fs');

const defaultFixtures = './_ft/ersatz-fixtures.yml';

function sleepFor( sleepDuration ){
    var now = new Date().getTime();
    while(new Date().getTime() < now + sleepDuration){ /* do nothing */ } 
}

hooks.beforeAll(function(t, done) {
   if(!fs.existsSync(defaultFixtures)){
      console.log('No fixtures found, skipping hook.');
      done();
      return;
   }

   var contents = fs.readFileSync(defaultFixtures, 'utf8');

   var options = {
      host: 'localhost',
      port: '9000',
      path: '/__configure',
      method: 'POST',
      headers: {
         'Content-Type': 'application/x-yaml'
      }
   };

   var req = http.request(options, function(res) {
      res.setEncoding('utf8');
   });

   req.write(contents);
   req.end();
   sleepFor(5000) // sleep in order to allow ersatz to load the sent config
   done();
});
