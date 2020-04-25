# Golang mailer service
This service is made for sending email via REST api
### Installation
##### Docker:
1. Build the container: `docker build -t dslmailer .`
2. Put config into `cfg` folder and rename it to cfg.json or just change the name in docker file
3. Run it: `docker run -p 443:443 -v $(pwd)/cfg:/cfg --name dslmailer --rm dslmailer:latest`
##### Linux/osx:
1. Make all in one time: `make all` -- this pull deps, create certificates and build sources

Use `make deploy` to not to generate 

#### Usage:
From command line: `bin/dslmailer <command> <path to config file>`
Command could be:
 - `start` -- starts service with config file
 - `config` -- creates blank config file

#### Api:
There are three endpoints:
1. `/` -- just to ping application
1. `/send` -- send email.
- Params:
   1. `to` -- email address to send mail to
   1. `subject` -- subject of email
   1. `body` -- body in **base64**
   1. `secret` -- just application password. You must use it. It is defined in `cfg.json`
- Returns (json with params):
   1. `code` -- `0 = OK`, `1 = ERROR`
   1. `message` -- error/ok message
   1. `error` -- empty if status is OK
   1. `uid` -- unique id of task(should be passed to `/status`)
1. `/status` -- get status of some email.(**Could be called once for one uid, after that id is removed**)
 - Params:
   1. `uid` -- unique id of task
   1. `secret` -- just application password. You must use it.
 - Returns: same as `/send`