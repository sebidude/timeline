# timeline

This is a simple cli app that logs the text you type in it. 
It can be used for serveral things:
* Note down things during an incident and use the timeline for a postmortem
* Track what happens over the day
* Meeting notes
* Protocols for something

The buffer is scrollable and will jump to the last entry when you add one. Use the arrow keys to scroll. 

```
go install github.com/sebidude/timeline
$GOPATH/bin/timeline my-timeline.txt
```
If you provide an existing file, the content will be loaded and you can append.

