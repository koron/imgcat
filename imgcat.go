package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
)

type Layout int

const (
	Vertical Layout = iota
	Horizontal
)

type Args struct {
	X, Y, W, H int
	Inputs     []string
	Output     string
	Layout     Layout
	Wrap       int
	Gap        int
	Margin     int
	Help       bool
}

type DrawData struct {
	File     string
	SrcPoint image.Point
	DstRect  image.Rectangle
}

func (d DrawData) Draw(dst draw.Image) error {
	f, err := os.Open(d.File)
	if err != nil {
		return err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	draw.Draw(dst, d.DstRect, src, d.SrcPoint, draw.Src)
	return nil
}

func flag2args() *Args {
	a := Args{
		Layout: Vertical,
	}
	var layout int
	flag.IntVar(&a.X, "x", 0, "X position, default means left edge")
	flag.IntVar(&a.Y, "y", 0, "Y position, default means top edge")
	flag.IntVar(&a.W, "width", -1, "width (must)")
	flag.IntVar(&a.H, "height", -1, "height (must)")
	flag.IntVar(&layout, "layout", 0, "layout, horz:0(default) vert:1")
	flag.IntVar(&a.Wrap, "wrap", 0, "num of images to wrap (default 0:nowrap)")
	flag.IntVar(&a.Gap, "gap", 0, "gap width between images (default 0)")
	flag.IntVar(&a.Margin, "margin", 0, "margin width around images (default 0)")
	flag.StringVar(&a.Output, "output", "", "output filename (must)")
	flag.BoolVar(&a.Help, "h", false, "show help")
	flag.Parse()
	a.Inputs = flag.Args()
	a.Layout = Layout(layout)
	return &a
}

func calcPos(a *Args, i int) (nx, ny int) {
	switch a.Layout {
	case Vertical:
		if a.Wrap == 0 {
			ny = i
		} else {
			nx, ny = i/a.Wrap, i%a.Wrap
		}
	case Horizontal:
		if a.Wrap == 0 {
			nx = i
		} else {
			nx, ny = i%a.Wrap, i/a.Wrap
		}
	}
	return
}

func calcDrawData(a *Args) ([]DrawData, image.Rectangle) {
	dd := make([]DrawData, 0, len(a.Inputs))
	m, w, h := a.Margin, a.W+a.Gap, a.H+a.Gap
	max_x, max_y := 0, 0
	for i, f := range a.Inputs {
		nx, ny := calcPos(a, i)
		x0, y0 := m+nx*w, m+ny*h
		x1, y1 := x0+a.W, y0+a.H
		x2, y2 := x1+m, y1+m
		dd = append(dd, DrawData{
			File:     f,
			SrcPoint: image.Pt(a.X, a.Y),
			DstRect:  image.Rect(x0, y0, x1, y1),
		})
		if x2 > max_x {
			max_x = x2
		}
		if y2 > max_y {
			max_y = y2
		}
	}
	return dd, image.Rect(0, 0, max_x, max_y)
}

func usage() {
	fmt.Fprintf(os.Stderr, "USAGE:  %s [OPTIONS] {IMAGE FILES}\n\nOPTIONS:\n",
		os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func writeFile(file string, m image.Image) error {
	var encode func(f *os.File) error
	switch filepath.Ext(file) {
	case ".jpeg":
	case ".jpg":
		encode = func(f *os.File) error {
			return jpeg.Encode(f, m, &jpeg.Options{90})
		}
	case ".png":
		encode = func(f *os.File) error {
			return png.Encode(f, m)
		}
	default:
		return errors.New("unknown ext: " + file)
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return encode(f)
}

func main() {
	args := flag2args()
	if args.Help || len(args.Inputs) == 0 {
		usage()
	} else if args.Output == "" {
		fmt.Fprintf(os.Stderr, "required '-output' arg\n\n")
		usage()
	} else if args.W < 0 || args.H < 0 {
		fmt.Fprintf(os.Stderr, "required '-width' and '-height'\n\n")
		usage()
	} else if args.Wrap < 0 {
		fmt.Fprintf(os.Stderr, "'-wrap' must be 0 or greater\n\n")
		usage()
	} else if args.Gap < 0 {
		fmt.Fprintf(os.Stderr, "'-gap' must be 0 or greater\n\n")
		usage()
	} else if args.Margin < 0 {
		fmt.Fprintf(os.Stderr, "'-margin' must be 0 or greater\n\n")
		usage()
	}

	dd, sz := calcDrawData(args)
	dst := image.NewRGBA(sz)
	draw.Draw(dst, sz, &image.Uniform{color.White}, image.ZP, draw.Src)
	for _, d := range dd {
		err := d.Draw(dst)
		if err != nil {
			log.Fatalf("failed to draw: %s\n", err)
		}
	}
	err := writeFile(args.Output, dst)
	if err != nil {
		log.Fatalf("failed: %s\n", err)
	}
}
