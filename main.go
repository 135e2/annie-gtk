package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appId = "com.experimental.annie-gtk"

var (
	DestFolder string
	URL        string
)

type outputBuffer struct {
	reader   *bufio.Reader
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

		ProgBarobj, err := builder.GetObject("progbar")
		progbar := ProgBarobj.(*gtk.ProgressBar)

		Aboutdialogobj, err := builder.GetObject("aboutdialog")
		aboutdialog := Aboutdialogobj.(*gtk.AboutDialog)
		aboutdialog.SetVersion(version)
		aboutdialog.SetTitle("About Page")

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
			if checkURL(URL) {
				AddText(textview, "Download started")
				progbar.SetFraction(0) // Reset ProgressBar
				AddText(textview, fmt.Sprintf("annie-gtk is now downloading %s => %s", URL, DestFolder))
				// TODO: Download progress
				output := &outputBuffer{
					reader:   nil,
					textview: textview,
				}
				savedStdout := os.Stdout
				r, w, _ := os.Pipe()
				output.attachReader(r, textview)
				os.Stdout = w

				defaultDownloader, data, err := setupDownloader(nil, URL)
				if err != nil {
					AddText(textview, "annie-backend got error while setting up downloader: "+err.Error())
				}
				err, Site, Title, Type, Size, FileNameLength, stream := GetInfo(defaultDownloader, data)
				if err != nil {
					AddText(textview, "annie-backend got error: "+err.Error())
				}

				go func() {
					var savedSize int64
					for {
						if len(stream.Parts) == 1 {
							savedSize, err = GetSize(defaultDownloader, data, Title, FileNameLength, stream.Parts[0])
							if err != nil {
								AddText(textview, "GetSize error:"+err.Error())
							}
						} else {
							savedSize = 0
							for index, part := range stream.Parts {
								partFileName := fmt.Sprintf("%s[%d]", Title, index)
								partSize, err := GetSize(defaultDownloader, data, partFileName, FileNameLength, part)
								if err != nil {
									AddText(textview, "GetSize error (multi parts):"+err.Error())
								}
								// AddText(textview, fmt.Sprintf("partSize[%d]: %.2f MiB", index, float64(partSize)/(1024*1024)))
								savedSize += partSize
							}
						}

						if savedSize < Size {
							// AddText(textview, fmt.Sprintf("Downloaded %.2f MiB/%.2f MiB", float64(savedSize)/(1024*1024), float64(Size)/(1024*1024)))
							progbar.SetFraction(float64(savedSize) / float64(Size))
							progbar.SetText(fmt.Sprintf("Downloaded %.2f MiB/%.2f MiB", float64(savedSize)/(1024*1024), float64(Size)/(1024*1024)))
							time.Sleep(500 * time.Millisecond)
						}
						line, err := output.reader.ReadString('\n')
						if err == nil || err == io.EOF {
							if line != "" {
								// AddText(textview, "DEBUG: "+line)
							}
							if err == io.EOF {
								break
							}
						} else {
							AddText(textview, err.Error())
							break
						}
						// fmt.Fprint(savedStdout, line)
					}
					AddText(textview, "Download completed")
					progbar.SetText("Download completed")
					progbar.SetFraction(1)
				}()

				go func() {
					AddText(textview, "Downloading from: "+Site)
					AddText(textview, "File title: "+Title)
					AddText(textview, "File type: "+Type)
					AddText(textview, "File size: "+fmt.Sprintf("%.2f MiB (%d Bytes)\n", float64(Size)/(1024*1024), Size))
					if err := Download(defaultDownloader, data); err != nil {
						AddText(textview, "Got error while downloading: "+err.Error())
					}

					err := w.Close()
					if err != nil {
						AddText(textview, "Got error while closing the output: "+err.Error())
					}
					os.Stdout = savedStdout
				}()
			} else {
				AddText(textview, "You typed something. but not valid URL!")
			}
		})

		menuitem1.Connect("select", func(menuitem1 *gtk.MenuItem) {
			about := About()
			about.SetTransientFor(win)
			about.Show()
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
	// b.reader = bufio.NewReaderSize(r, bufio.MaxScanTokenSize)
	b.reader = bufio.NewReader(r)
	b.textview = textview
}

func isWindow(obj glib.IObject) (*gtk.Window, error) {
	// Make type assertion (as per gtk.go).
	if win, ok := obj.(*gtk.Window); ok {
		return win, nil
	}
	return nil, errors.New("not a *gtk.Window")
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
	text = time.Now().Format("15:04:05 ") + text + "\n"
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

func checkURL(URL string) bool {
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return false
	}
	return true
}
