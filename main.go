package main

import (
	"bufio"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gotk3/gotk3/gtk"
)

const appId = "com.github-135e2.annie-gtk"

var (
	DestFolder string
	URL        string
)

type outputBuffer struct {
	reader   *bufio.Reader
	textview *gtk.TextView
}

func main() {
	app, err := gtk.ApplicationNew(appId, glib.APPLICATION_FLAGS_NONE)

	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	app.Connect("activate", func() {
		onActivate(app)
	})

	app.Run(os.Args)
}

func onActivate(application *gtk.Application) {
	// Create a new application window.
	win := setupWindow("annie-gtk", application)
	mainBox := setupBox(gtk.ORIENTATION_VERTICAL)
	win.Add(mainBox)

	// Setup menuBar
	menuBar := setupMenuBar()
	menuItem1 := setupMenuItem()
	menuBar.Append(menuItem1)
	mainBox.PackStart(menuBar, false, true, 0)

	// Setup box1
	box1 := setupBox(gtk.ORIENTATION_HORIZONTAL)
	label1 := setupLabel("视频链接")
	box1.PackStart(label1, false, true, 0)
	entry1, err := gtk.EntryNew()
	errorCheck(err)
	box1.PackStart(entry1, true, true, 0)
	mainBox.PackStart(box1, false, true, 0)

	// Setup box2
	box2 := setupBox(gtk.ORIENTATION_HORIZONTAL)
	label2 := setupLabel("目标文件夹")
	box2.PackStart(label2, false, true, 0)
	fileButton, err := gtk.FileChooserButtonNew("选择文件", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
	box2.PackStart(fileButton, false, true, 0)
	startButton, err := gtk.ButtonNewWithLabel("开始下载")
	errorCheck(err)
	startButton.SetMarginStart(256)
	startButton.SetMarginEnd(256) // Set Margin to 256
	box2.PackStart(startButton, true, true, 0)
	mainBox.PackStart(box2, false, true, 0)

	// Setup box3
	box3 := setupBox(gtk.ORIENTATION_HORIZONTAL)
	label3 := setupLabel("下载进度")
	box3.PackStart(label3, false, true, 0)
	progressBar, err := gtk.ProgressBarNew()
	progressBar.SetShowText(true)
	box3.PackStart(progressBar, true, true, 0)
	mainBox.PackStart(box3, false, true, 0)

	// Setup scrolled window
	scrolledWindow, err := gtk.ScrolledWindowNew(nil, nil)
	errorCheck(err)
	textview := setupTextView()
	scrolledWindow.Add(textview)
	mainBox.PackStart(scrolledWindow, true, true, 0)

	// Get exec path
	ex, err := os.Executable()
	errorCheck(err)
	exPath := filepath.Dir(ex) + string(os.PathSeparator)

	// Deal with signals
	fileButton.Connect("current-folder-changed", func() {
		DestFolder, err = fileButton.GetCurrentFolder()
		errorCheck(err)
	})

	startButton.Connect("clicked", func() {
		URL, err = entry1.GetText()
		errorCheck(err)
		DestFolder, err = fileButton.GetCurrentFolder()
		errorCheck(err)
		if checkURL(URL) {
			AddText(textview, "Download started")
			progressBar.SetFraction(0) // Reset ProgressBar
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
						progressBar.SetFraction(float64(savedSize) / float64(Size))
						progressBar.SetText(fmt.Sprintf("Downloaded %.2f MiB/%.2f MiB", float64(savedSize)/(1024*1024), float64(Size)/(1024*1024)))
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
				progressBar.SetText("Download completed")
				progressBar.SetFraction(1)
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

	menuItem1.Connect("select", func(menuitem1 *gtk.MenuItem) {
		about := About(exPath)
		about.SetTransientFor(win)
		about.Show()
	})

	// Launch the application
	win.ShowAll()
}

// Modified from fanaticscripter/annie-mingui

func (b *outputBuffer) attachReader(r io.Reader, textview *gtk.TextView) {
	// b.reader = bufio.NewReaderSize(r, bufio.MaxScanTokenSize)
	b.reader = bufio.NewReader(r)
	b.textview = textview
}

func setupWindow(title string, application *gtk.Application) *gtk.ApplicationWindow {
	win, err := gtk.ApplicationWindowNew(application)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}

	win.SetTitle(title)
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.SetDefaultSize(800, 600)
	return win
}

func setupBox(orient gtk.Orientation) *gtk.Box {
	box, err := gtk.BoxNew(orient, 2)
	if err != nil {
		log.Fatal("Unable to create box:", err)
	}
	return box
}

func setupLabel(text string) *gtk.Label {
	label, err := gtk.LabelNew(text)
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	return label
}

func setupMenuBar() *gtk.MenuBar {
	menubar, err := gtk.MenuBarNew()
	if err != nil {
		log.Fatal("Unable to create MenuBar:", err)
	}
	return menubar
}

func setupMenuItem() *gtk.MenuItem {
	menuitem, err := gtk.MenuItemNewWithLabel("关于")
	if err != nil {
		log.Fatal("Unable to create MenuItem:", err)
	}
	return menuitem
}

func setupTextView() *gtk.TextView {
	tv, err := gtk.TextViewNew()
	if err != nil {
		log.Fatal("Unable to create TextView:", err)
	}
	return tv
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
