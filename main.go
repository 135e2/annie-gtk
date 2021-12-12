package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	appId = "com.experimental.annie-gtk"
	about = `annie-gtk v0.1.0
Copyright (c) 2021 135e2 <135e2@135e2.tk>
annie-gtk is an minimal GTK-3 wrapper for iawia002/annie (a command-line video downloader), licensed under GPLv3.
Credits:
	- iawia002/annie, MIT license;
	- The GTK Project, LGPLv2.1+;
	- gotk3/gotk3, ISC license;
	- fanaticscripter/annie-mingui, MIT license

Project URL: https://github.com/135e2/annie-mingui
`
)

var (
	DestFolder string
	URL        string
)

type outputBuffer struct {
	reader   *bufio.Reader
	scanner  *bufio.Scanner
	textview *gtk.TextView
}

func main() {
	// Create a new application.
	application, err := gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)
	errorCheck(err)

	// Connect function to application activate event
	application.Connect("activate", func() {
		// Get the GtkBuilder UI definition in the glade file.
		builder, err := gtk.BuilderNewFromFile("annie-gtk.ui")
		errorCheck(err)

		Window1obj, err := builder.GetObject("window1")
		errorCheck(err)

		// Verify that the object is a pointer to a gtk.ApplicationWindow.
		win, err := isWindow(Window1obj)
		errorCheck(err)
		win.SetTitle("annie-gtk")

		Entry1obj, err := builder.GetObject("entry1")
		errorCheck(err)
		entry1 := Entry1obj.(*gtk.Entry)

		StartButtonobj, err := builder.GetObject("startbutton")
		startbutton := StartButtonobj.(*gtk.Button)

		FileButtonobj, err := builder.GetObject("filebutton")
		errorCheck(err)
		filebutton := FileButtonobj.(*gtk.FileChooserButton)
		filebutton.SetCurrentFolder("./")

		Textviewobj, err := builder.GetObject("textview")
		textview := Textviewobj.(*gtk.TextView)

		MenuItem1obj, err := builder.GetObject("menuitem1")
		menuitem1 := MenuItem1obj.(*gtk.MenuItem)

		dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, about)
		dialog.SetTitle("About Page")

		// Deal with signals
		filebutton.Connect("file-set", func() {
			DestFolder, err = filebutton.GetCurrentFolder()
			errorCheck(err)
		})

		startbutton.Connect("clicked", func() {
			URL, err = entry1.GetText()
			errorCheck(err)
			DestFolder, err = filebutton.GetCurrentFolder()
			errorCheck(err)
			if len(URL) > 0 {
				AddText(textview, time.Now().Format("15:04:05 ")+"Download started")
				AddText(textview, time.Now().Format("15:04:05 ")+fmt.Sprintf("annie-gtk is now downloading %s => %s", URL, DestFolder))
				// TODO: Download progress
				output := &outputBuffer{
					reader:   nil,
					scanner:  nil,
					textview: textview,
				}
				savedStdout := os.Stdout
				r, w, _ := os.Pipe()
				output.attachReader(r, textview)
				os.Stdout = w

				go func() {
					for {
						_, err := output.readLineAndUpdate()
						if err != nil {
							if err == io.EOF {
								break
							}
							glog.Fatal(err)
						}
						// fmt.Fprint(savedStdout, line)
					}
					AddText(textview, time.Now().Format("15:04:05 ")+"Download completed")
				}()

				go func() {
					if Download(nil, URL) != nil {
						fmt.Println(time.Now().Format("15:04:05 ") + "On network errors, e.g. HTTP 403, please retry a few times.")
					}
					w.Close()
					os.Stdout = savedStdout
				}()
			}
		})

		menuitem1.Connect("select", func() {
			dialog.Run()
			dialog.Destroy()
		})

		// Show the Window and all of its components.
		win.Show()
		application.AddWindow(win)
	})

	// Launch the application
	os.Exit(application.Run(os.Args))
}

// Modified from fanaticscripter/annie-mingui

func (b *outputBuffer) attachReader(r io.Reader, textview *gtk.TextView) {
	b.reader = bufio.NewReaderSize(r, 64*1024)
	b.scanner = bufio.NewScanner(b.reader)
	b.textview = textview
	re := regexp.MustCompile(`^[^\r\n]*(\r\n|\r|\n)`)
	b.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		token = re.Find(data)
		if token != nil {
			return len(token), token, nil
		}
		if !atEOF {
			return 0, nil, nil
		}
		return 0, data, bufio.ErrFinalToken
	})
}

func (b *outputBuffer) readLineAndUpdate() (fullLine string, err error) {
	if !b.scanner.Scan() {
		err = b.scanner.Err()
		if err != nil {
			return "", err
		}
		err = io.EOF
	}
	fullLine = b.scanner.Text()
	if len(fullLine) > 0 {
		AddText(b.textview, fullLine)
	}
	return
}

func isWindow(obj glib.IObject) (*gtk.Window, error) {
	// Make type assertion (as per gtk.go).
	if win, ok := obj.(*gtk.Window); ok {
		return win, nil
	}
	return nil, errors.New("not a *gtk.Window")
}

func isDialog(obj glib.IObject) (*gtk.Dialog, error) {
	// Make type assertion (as per gtk.go).
	if dialog, ok := obj.(*gtk.Dialog); ok {
		return dialog, nil
	}
	return nil, errors.New("not a *gtk.Dialog")
}

func GetBuffer(tv *gtk.TextView) *gtk.TextBuffer {
	buffer, err := tv.GetBuffer()
	if err != nil {
		log.Fatal("Unable to get buffer:", err)
	}
	return buffer
}

func AddText(tv *gtk.TextView, text string) {
	// Add \n at the end of the message
	text += "\n"
	buffer := GetBuffer(tv)
	endIter := buffer.GetEndIter()
	buffer.Insert(endIter, text)
}

func errorCheck(e error) {
	if e != nil {
		// panic for any errors.
		log.Panic(e)
	}
}
