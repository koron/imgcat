package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
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

func args2drawdata(a *Args) ([]DrawData, image.Rectangle) {
	dd := make([]DrawData, 0, len(a.Inputs))
	m := a.Margin
	unit_w, unit_h := a.W+a.Gap, a.H+a.Gap
	max_w, max_h := 0, 0
	for i, s := range a.Inputs {
		var x, y int
		switch a.Layout {
		case Vertical:
			if a.Wrap == 0 {
				x, y = m, i*unit_h+m
			} else {
				x = (i/a.Wrap)*unit_w + m
				y = (i%a.Wrap)*unit_h + m
			}
		case Horizontal:
			if a.Wrap == 0 {
				x, y = i*unit_w+m, m
			} else {
				x = (i%a.Wrap)*unit_w + m
				y = (i/a.Wrap)*unit_h + m
			}
		}
		right, bottom := x+a.W, y+a.H
		dd = append(dd, DrawData{
			File:     s,
			SrcPoint: image.Pt(a.X, a.Y),
			DstRect:  image.Rect(x, y, right, bottom),
		})
		most_right, most_bottom := right + m, bottom + m
		if most_right > max_w {
			max_w = most_right
		}
		if most_bottom > max_h {
			max_h = most_bottom
		}
	}
	return dd, image.Rect(0, 0, max_w, max_h)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func writeFile(file string, m image.Image) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, m)
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

	dd, sz := args2drawdata(args)
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
