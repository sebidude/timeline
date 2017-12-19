package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

var (
	tlfilename, vbuf, buf string
)

type Timeline struct {
	title  string
	tlfile *os.File
	gocui  *gocui.Gui
}

func NewTimeline(tlfilename string) *Timeline {
	tlf, err := os.OpenFile(tlfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Panicln(err)
	}
	tl := &Timeline{
		tlfile: tlf,
		title:  tlf.Name(),
	}
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	g.FgColor = gocui.ColorBlack
	g.BgColor = gocui.ColorWhite
	tl.gocui = g

	return tl
}

func (tl *Timeline) ReadFile() []byte {
	content, err := ioutil.ReadFile(tl.title)
	if err != nil {
		log.Panicln(err)

	}
	return content
}

func (tl *Timeline) Close() {
	tl.gocui.Close()
	tl.tlfile.Close()
}

func (tl *Timeline) quit(g *gocui.Gui, v *gocui.View) error {
	vbuf = v.ViewBuffer()
	buf = v.Buffer()
	return gocui.ErrQuit
}

func (tl *Timeline) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 0, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Wrap = true
		v.Title = fmt.Sprintf(" Timeline: %s ", tl.title)
		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	if p, err := g.SetView("timeline", 0, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		p.BgColor = gocui.ColorWhite
		p.FgColor = gocui.ColorBlack
		p.Editable = false
		p.Wrap = true
		p.Frame = false

		p.Autoscroll = true
		p.Title = tl.title
		p.SelFgColor = gocui.ColorWhite
		p.SelBgColor = gocui.ColorBlack
		content := tl.ReadFile()
		tl.SetContent(content)
	}
	return nil
}

func (tl *Timeline) scrollUp(g *gocui.Gui, v *gocui.View) error {
	tl.scroll(-1)
	return nil
}

func (tl *Timeline) scrollDown(g *gocui.Gui, v *gocui.View) error {
	tl.scroll(1)
	return nil
}

func (tl *Timeline) scroll(dy int) error {
	// Grab the view that we want to scroll.
	tlv, _ := tl.gocui.View("timeline")

	// Get the size and position of the view.
	_, y := tlv.Size()
	ox, oy := tlv.Origin()

	// If we're at the bottom...
	if oy+dy > strings.Count(tlv.ViewBuffer(), "\n")-y-1 {
		// Set autoscroll to normal again.
		tlv.Autoscroll = true
	} else {
		// Set autoscroll to false and scroll.
		tlv.Autoscroll = false
		tlv.SetOrigin(ox, oy+dy)
	}
	return nil
}
func (tl *Timeline) SetContent(content []byte) {
	v, _ := tl.gocui.View("timeline")
	fmt.Fprintf(v, string(content))
}

func (tl *Timeline) logEntry(g *gocui.Gui, v *gocui.View) error {
	if len(v.Buffer()) == 0 {
		return nil
	}
	out, _ := g.View("timeline")
	out.Autoscroll = true
	t := time.Now()
	timestring := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	entry := fmt.Sprintf("%s: %s", timestring, v.Buffer())
	fmt.Fprintf(out, entry)
	_, err := tl.tlfile.WriteString(string(entry))
	if err != nil {
		log.Panicf("%#v\n", err)
		//fmt.Fprint(out, err)
	}
	v.Clear()
	v.SetCursor(0, 0)
	return nil
}

func main() {
	tlfilename := os.Getenv("TIMELINEFILE")
	if len(os.Args) > 1 {
		tlfilename = os.Args[1]
	}

	if len(tlfilename) < 1 {
		tlfilename = ".timeline"
	}

	timeline := NewTimeline(tlfilename)
	defer timeline.Close()

	timeline.gocui.Cursor = true
	timeline.gocui.Mouse = true
	timeline.gocui.ASCII = false

	timeline.gocui.SetManagerFunc(timeline.layout)

	if err := timeline.gocui.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone, timeline.logEntry); err != nil {
		log.Println(vbuf)
	}

	if err := timeline.gocui.SetKeybinding("main", gocui.KeyCtrlC, gocui.ModNone, timeline.quit); err != nil {
		log.Panicln(err)
	}
	if err := timeline.gocui.SetKeybinding("timeline", gocui.KeyCtrlC, gocui.ModNone, timeline.quit); err != nil {
		log.Panicln(err)
	}
	if err := timeline.gocui.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, timeline.scrollUp); err != nil {
		log.Panicln(err)
	}
	if err := timeline.gocui.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, timeline.scrollDown); err != nil {
		log.Panicln(err)
	}
	if err := timeline.gocui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

}
