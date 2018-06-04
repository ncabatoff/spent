# spent

Spent is inspired by [nf's spent](https://gist.github.com/nf/7d378cb9417144caf272),
but written for Linux and with my own spin on it. 

It keeps track of when the user is idle or not, and when not idle, what window is focused.
A CSV report is continuously emitted to stdout.

## Details 

When transitioning to idle or the focused window's title changes, write what
we've been doing up to that point. That means either the previous focused
window title, or 'idle' if we were idle long enough or screensaver was active.

CSV base output fields:
- endtime
- elapsed seconds
- full window title

Additional fields that may be included, as detailed below:
- app type
- app context
- app detail

### browser

The app type `browser` is used if the window title ends in " - Chromium".
In that case app context is the website open, and app detail is the path part of the URL.

To get app context and app detail you must install the [Url in title](https://chrome.google.com/webstore/detail/url-in-title/ignpacbgnbnkaiooknalneoeladjnfgb) Chrome extension, 
and configure it with format 

```
{title} - {protocol}://{hostname}{port}/{path}
```

### terminal

The app type `terminal` is used if the window title ends in " - xterm".
In that case app context is the user@host, and app detail is the current working directory.

On Ubuntu 16.04 the default PS1 in .bashrc is:

```
PS1="\[\e]0;${debian_chroot:+($debian_chroot)}\u@\h: \w\a\]$PS1"
```

Extend it like this to make sure title ends in " - xterm":

```
PS1="\[\e]0;${debian_chroot:+($debian_chroot)}\u@\h: \w - xterm\a\]$PS1"
```

### editor

The app type `editor` is used if the window title ends in " - Visual Studio Code".
In that case app context is the root directory open and app detail is the file being edited.

To get app context and app detail you must set

```
"window.title": "${activeEditorMedium}${separator}${rootPath}${separator}${appName}"
```

## Dependencies
- xprintidle, from the Debian/Ubuntu package of the same name, or see https://github.com/g0hl1n/xprintidle or https://github.com/lucianposton/xprintidle
- [xdotool](https://www.semicomplete.com/projects/xdotool/)
- gnome-screensaver-command
