package main

import (
	"flag"
	"fmt"
        "math"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
)

type Layout int

const (
	Vertical Layout = iota
	Horizontal
	Tiling
)

type Args struct {
	X, Y, W, H int
	H_TILE_NUM int
	Inputs     []string
	Output     string
	Layout     Layout
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
	draw.Draw(dst, d.DstRect, src, d.SrcPoint, draw.Over)
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
	flag.IntVar(&layout, "layout", 0, "layout, horz:0 vert:1 tile:2")
	flag.IntVar(&a.H_TILE_NUM, "h_tile_num", -1, "number of horizon tile (must if and only if layout=Tiling")
	flag.StringVar(&a.Output, "output", "", "output filename (must)")
	flag.BoolVar(&a.Help, "h", false, "show help")
	flag.Parse()
	a.Inputs = flag.Args()
	a.Layout = Layout(layout)
	return &a
}

func args2drawdata(a *Args) []DrawData {
	dd := make([]DrawData, 0, len(a.Inputs))
	for i, s := range a.Inputs {
		var x, y int
		switch a.Layout {
		case Vertical:
			x, y = 0, i*a.H
		case Horizontal:
			x, y = i*a.W, 0
		case Tiling:
			x = (i % a.H_TILE_NUM) * a.W
			y = (i / a.H_TILE_NUM) * a.H
		}
		dd = append(dd, DrawData{
			File:     s,
			SrcPoint: image.Pt(a.X, a.Y),
			DstRect:  image.Rect(x, y, x+a.W, y+a.H),
		})
	}
	return dd
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func size(a *Args) image.Rectangle {
	var w, h int
	switch a.Layout {
	case Vertical:
		w = a.W
		h = a.H * len(a.Inputs)
	case Horizontal:
		w = a.W * len(a.Inputs)
		h = a.H
	case Tiling:
		if len(a.Inputs) < a.H_TILE_NUM {
			w = a.W
		} else {
			w = a.W * a.H_TILE_NUM
		}
		h = a.H * int(math.Ceil(float64(len(a.Inputs)) / float64(a.H_TILE_NUM)))
	}
	return image.Rect(0, 0, w, h)
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
		fmt.Fprintf(os.Stderr, "required '-width' and '-width'\n\n")
		usage()
	} else if args.Layout == 2 && args.H_TILE_NUM < 0 {
		fmt.Fprintf(os.Stderr, "required '-h_tile_num' when you chose tile layout.\n\n")
		usage()
	}

	dst := image.NewRGBA(size(args))
	for _, d := range args2drawdata(args) {
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
