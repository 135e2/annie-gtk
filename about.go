package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const comment = "annie-gtk is an minimal GTK-3 wrapper for iawia002/annie (a command-line video downloader), licensed under GPLv3.\nCredits:\n\t- iawia002/annie, MIT license;\n\t- The GTK Project, LGPLv2.1+;\n\t- gotk3/gotk3, ISC license;\n\t- fanaticscripter/annie-mingui, MIT license"

func About(path string) *gtk.AboutDialog {
	about, _ := gtk.AboutDialogNew()
	about.SetModal(true)
	about.SetProgramName("annie-gtk")
	about.SetVersion(VERSION)
	about.SetTitle("About Page")
	about.SetCopyright("Copyright (c) 2021 135e2 <135e2@135e2.tk>")
	about.SetLicenseType(gtk.LICENSE_GPL_3_0)
	about.SetComments(comment)
	logo, err := gdk.PixbufNewFromFile(path + "logo.png")
	errorCheck(err)
	about.SetLogo(logo)
	about.SetWebsite("https://github.com/135e2/annie-gtk")
	about.SetWebsiteLabel("Source Code")
	about.SetAuthors([]string{
		"135e2 <135e2@135e2.tk>",
		"iawia002",
		"Zhiming Wang",
	})
	return about
}
