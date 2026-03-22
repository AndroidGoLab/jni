// Command gio demonstrates using this library's typed JNI wrappers
// from within a Gio UI application. It reads Android build information
// via the build package and displays it as text labels.
package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		if err := run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run() error {
	th := material.NewTheme()
	th.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	info := deviceInfo()

	var w app.Window
	w.Option(app.Title("JNI + Gio"))

	var ops op.Ops
	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
					label(th, "JNI Typed Wrappers + Gio"),
					label(th, ""),
					label(th, info),
				)
			})
			e.Frame(gtx.Ops)
		}
	}
}

func label(th *material.Theme, txt string) layout.FlexChild {
	return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		l := material.Body1(th, txt)
		l.TextSize = unit.Sp(18)
		return l.Layout(gtx)
	})
}
