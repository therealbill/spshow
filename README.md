# StatusPage.io Show Tool

This tool was written for use via
[https://github.com/matryer/bitbar#writing-plugins](BitBar). It will
show the status from your statuspage.io page.

# Installation
First, decide if you are going to use it in script mode or via the
binary. This document will assume the use of the binary. If you intend
to run as a script, you should already know how. ;) If you already have
BitBar with plugins installed, put it in your plugin directory.

Personally, I like to have a `~/BitBar` directory with a pair of
subdirectories: `Available` and `Enabled`. I configure my plugins
directory to be `Enabled`, so I can put the plugins I want to use in
`Available` and symlink in `Enabled` but use refresh settings. For
example: if you want this plugin to refresh once every five minutes, put
it in `Available`, then in `Enabled` execute:

`ln -s ../Available/spshow spshow.5m.bin`

# Configuration
This app needs configuration. To keep it simple, copy the sample.config
from the repo to `~/.pdshow` and edit it to put in your PageID and
auth token. If you want to change the "icons", which are actuall emoji,
feel free to edit them in - but be sure they are "github supported"
emoji.

If you look at the sections you will see one for "scheduled" and one for
"resolved". These sections control what the app does about scheduled and
resolved incidents in statuspage. For example, if you don't care to see
scheduled incidents, set `enabled=false` and they will not be shown.

If you have multiple accounts to monitor, make a file for each, and
instead of calling the binary, use a shell script which passes `-c
`~/.OTHERCONFIGFILE` or the full path for each and you should see them
both show. Also: My condolences on monitoring more than one.

# TODO
* The Statuspage client I am using is ... less than complete and somewhat
annoying. Thus I need to fix it as well as add some missing calls such
as the page info API.
* more customization options
 
