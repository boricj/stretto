package main

import (
	"encoding/binary"
	"io"
	"os"
	"strings"

	"github.com/stretto-editor/gocui"
)

var currentFile string

var fileMode = "file"
var editMode = "edit"
var cmdMode = "cmd"

func initModes(g *gocui.Gui) {
	g.SetMode(cmdMode)
	g.SetMode(fileMode)
	g.SetMode(editMode)
}

func initKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(fileMode, "", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "", gocui.KeyCtrlT, gocui.ModNone, currTopViewHandler("cmdline")); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "cmdline", gocui.KeyCtrlT, gocui.ModNone, currTopViewHandler("main")); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "", gocui.KeyHome, gocui.ModNone, cursorHome); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "", gocui.KeyEnd, gocui.ModNone, cursorEnd); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "", gocui.KeyPgup, gocui.ModNone, goPgUp); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "", gocui.KeyPgdn, gocui.ModNone, goPgDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "main", gocui.KeyCtrlS, gocui.ModNone, saveMain); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "main", gocui.KeyTab, gocui.ModNone, switchModeTo(editMode)); err != nil {
		return err
	}
	if err := g.SetKeybinding(fileMode, "main", gocui.KeyCtrlF, gocui.ModNone, search); err != nil {
		return err
	}
	if err := g.SetKeybinding(editMode, "main", gocui.KeyTab, gocui.ModNone, switchModeTo(fileMode)); err != nil {
		return err
	}
	if err := g.SetKeybinding(editMode, "", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	g.SetCurrentMode(fileMode)

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func switchModeTo(name string) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		if err := g.SetCurrentMode(name); err != nil {
			return err
		}
		if v.Name() == "cmdline" {
			v.SetCursor(0, 0)
			v.Clear()
		}
		if name == "cmd" {
			if err := g.SetCurrentView(name); err != nil {
				return err
			}
		}
		return nil
	}
}

func readCmd(g *gocui.Gui, v *gocui.View) error {

	return nil
}

func currTopViewHandler(name string) gocui.KeybindingHandler {
	return func(g *gocui.Gui, v *gocui.View) error {
		if err := g.SetCurrentView(name); err != nil {
			return err
		}
		if _, err := g.SetViewOnTop(name); err != nil {
			return err
		}
		return nil
	}
}

func cursorHome(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, _ := v.Cursor()
		v.MoveCursor(-cx, 0, true)
	}
	return nil
}

func cursorEnd(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		l, _ := v.Line(cy)
		v.MoveCursor(len(l)-cx, 0, true)
	}
	return nil
}

func goPgUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		_, y := v.Size()
		yOffset := oy - y
		if yOffset < 0 {
			v.SetOrigin(ox, 0)
			if oy == 0 {
				_, cy := v.Cursor()
				v.MoveCursor(0, -cy, false)
			}
		} else {
			v.SetOrigin(ox, yOffset)
			v.MoveCursor(0, -yOffset, false)
		}
	}
	return nil
}

func goPgDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		_, y := v.Size()
		err := v.SetOrigin(ox, oy+y)
		if err != nil {
			return err
		}
		v.MoveCursor(0, 0, false)
	}
	return nil
}

func saveMain(g *gocui.Gui, v *gocui.View) error {
	if currentFile == "" {
		return nil
	}
	f, err := os.OpenFile(currentFile, os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	p := make([]byte, 5)
	v.Rewind()
	var size int64 = -1
	for {
		n, err := v.Read(p)
		size += int64(n)
		if n > 0 {
			if _, er := f.Write(p[:n]); err != nil {
				return er
			}
		}
		if err == io.EOF {
			f.Truncate(size * int64(binary.Size(p[0])))
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func search(g *gocui.Gui, v *gocui.View) error {
	var s string
	var err error
	var sameline = 1

	x, y := v.Cursor()

	for i := 0; err == nil; i++ {
		s, err = v.Line(y + i)
		if err == nil {
			// size of line is long enough to move the cursor
			if x < len(s)-1 {
				indice := strings.Index(s[x+sameline:], "deux") // string will be taken into parameter after refactoring structure

				// existing element on this line
				if indice >= 0 {
					if sameline == 0 {
						x, y = v.Cursor()
						v.MoveCursor(indice+sameline-x, i, false)
					} else {
						v.MoveCursor(indice+sameline, i, false)
					}
					return nil
				}
			}
			x = 0
			sameline = 0
		}
	}
	return nil
}
