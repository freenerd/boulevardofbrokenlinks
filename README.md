# Boulevard Of Broken Links
## What?

Fetch a URL, parse all hyperlinks from the returned HTML and check each hyperlinks status. Collect all broken links (http status >= 400) and send an email to owner about it.

## TODO

Homepage
Connect with Github (because this is only supposed to be for jekyll blogs)
Token for github callback (or better: automatic setup after connection)
WaitGroups to figure out, when a site has been fully crawled
Do some buffering, if people deploy often, they should only get an email every X minutes
unsubscribe link in the emails

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
