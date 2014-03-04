# Boulevard Of Broken Links
## What?

Fetch a URL, parse all hyperlinks from the returned HTML and check each hyperlinks status. Collect all broken links (http status >= 400) and send an email to owner about it. Triggered via webhook.

Example email:

```
http://freenerd.de/: 404 http://www.freenerd.de/hackhpi/
http://freenerd.de/: 503 http://takesquestions.com/
http://freenerd.de/: 403 http://bcn.musichackday.org/2012/
```

## Installation

* Install `go`
* `go get github.com/freenerd/boulevardofbrokenlinks`

## Run

```
cd "$GOPATH/src/github.com/freenerd/boulevardofbrokenlinks"
CHECK_URL="http://www.freenerd.de" go run web.go handler.go
curl "localhost:8080/trigger"
```

Emails will only be sent, if sendgrid environment is configured. Otherwise output to STDOUT.

## Deploy on heroku

```
heroku apps:create -b https://github.com/kr/heroku-buildpack-go.git
heroku addons:add sendgrid:starter
heroku config:set EMAIL_RECIPIENT=recipient@example.com
heroku config:set CHECK_URL=http://www.freenerd.de
git push heroku master
```

The scans are triggered via a webhook. To e.g. set it up with a github repo, go to your repo settings -> webhooks and enter the url `herokuapp-url/trigger` where you find out `herokuapp-url` via `heroku info | grep "Web URL"`.

## TODO

* Test suite
* Split web.go into more modules
* Better inline documentation
* Homepage
* Connect with Github (because this is only supposed to be for jekyll blogs)
* Token for github callback (or better: automatic setup after connection)
* WaitGroups to figure out, when a site has been fully crawled
* Do some buffering, if people deploy often, they should only get an email every X minutes
* unsubscribe link in the emails

## Caveats

- Since everything is in memory, a deploy kills currently ongoing scans

## Why the name?

```
  I walk a lonely URL
  The only one that I have ever known
  Don't know where it goes
  But it's only me and I walk alone
```

https://www.youtube.com/watch?v=tijW_SrCoxs
