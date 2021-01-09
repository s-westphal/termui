// Copyright 2017 Zack Guo <zack.y.guo@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT license that can
// be found in the LICENSE file.

package widgets

import (
	"fmt"
	"image"
	"math"

	. "github.com/s-westphal/termui/v3"
)

// Plot has two modes: line(default) and scatter.
// Plot also has two marker types: braille(default) and dot.
// A single braille character is a 2x4 grid of dots, so using braille
// gives 2x X resolution and 4x Y resolution over dot mode.
type Plot struct {
	Block

	Data       [][]float64
	DataLabels []string
	MaxVal     float64
	MinVal     float64
	XMaxVal    float64
	XMinVal    float64

	LineColors []Color
	AxesColor  Color // TODO
	ShowAxes   bool

	Marker          PlotMarker
	DotMarkerRune   rune
	PlotType        PlotType
	HorizontalScale int
	DrawDirection   DrawDirection // TODO
}

const (
	xAxisLabelsHeight = 1
	yAxisLabelsWidth  = 4
	xAxisLabelsGap    = 2
	yAxisLabelsGap    = 1
)

type PlotType uint

const (
	LineChart PlotType = iota
	ScatterPlot
)

type PlotMarker uint

const (
	MarkerBraille PlotMarker = iota
	MarkerDot
)

type DrawDirection uint

const (
	DrawLeft DrawDirection = iota
	DrawRight
)

func NewPlot() *Plot {
	return &Plot{
		Block:           *NewBlock(),
		LineColors:      Theme.Plot.Lines,
		AxesColor:       Theme.Plot.Axes,
		Marker:          MarkerBraille,
		DotMarkerRune:   DOT,
		Data:            [][]float64{},
		HorizontalScale: 1,
		DrawDirection:   DrawRight,
		ShowAxes:        true,
		PlotType:        LineChart,
		MinVal:          math.Inf(1),
		MaxVal:          math.Inf(-1),
		XMinVal:         math.Inf(1),
		XMaxVal:         math.Inf(-1),
	}
}

func (self *Plot) renderBraille(buf *Buffer, drawArea image.Rectangle, minVal float64, maxVal float64) {
	canvas := NewCanvas()
	canvas.Rectangle = drawArea
	xDx := MaxFloat64(1, self.XMaxVal-self.XMinVal)

	switch self.PlotType {
	case ScatterPlot:
		for i, x := range self.Data[0] {
			y := self.Data[1][i]
			height := int((y - minVal) / MaxFloat64(1, maxVal-minVal) * float64(drawArea.Dy()-1))
			canvas.SetPoint(
				image.Pt(
					(drawArea.Min.X+int((x-self.XMinVal)*float64(self.HorizontalScale*(drawArea.Dx()-1))/xDx))*2,
					(drawArea.Max.Y-height-1)*4,
				),
				SelectColor(self.LineColors, 0),
			)

		}
	case LineChart:
		for i, line := range self.Data {
			previousHeight := int(((line[1] - minVal) / MaxFloat64(1, maxVal-minVal)) * float64(drawArea.Dy()-1))
			for j, val := range line[1:] {
				height := int((val - minVal) / MaxFloat64(1, maxVal-minVal) * float64(drawArea.Dy()-1))
				canvas.SetLine(
					image.Pt(
						(drawArea.Min.X+(j*self.HorizontalScale))*2,
						(drawArea.Max.Y-previousHeight-1)*4,
					),
					image.Pt(
						(drawArea.Min.X+((j+1)*self.HorizontalScale))*2,
						(drawArea.Max.Y-height-1)*4,
					),
					SelectColor(self.LineColors, i),
				)
				previousHeight = height
			}
		}
	}

	canvas.Draw(buf)
}

func (self *Plot) renderDot(buf *Buffer, drawArea image.Rectangle, minVal float64, maxVal float64) {
	xDx := MaxFloat64(1, self.XMaxVal-self.XMinVal)
	switch self.PlotType {
	case ScatterPlot:
		for i, x := range self.Data[0] {
			y := self.Data[1][i]
			height := int((y - minVal) / MaxFloat64(1, maxVal-minVal) * float64(drawArea.Dy()-1))
			point := image.Pt(drawArea.Min.X+int((x-self.XMinVal)*float64(self.HorizontalScale*(drawArea.Dx()-1))/xDx), drawArea.Max.Y-1-height)
			if point.In(drawArea) {
				buf.SetCell(
					NewCell(self.DotMarkerRune, NewStyle(SelectColor(self.LineColors, 0))),
					point,
				)

			}
		}
	case LineChart:
		for i, line := range self.Data {
			for j := 0; j < len(line) && j*self.HorizontalScale < drawArea.Dx(); j++ {
				val := line[j]
				height := int((val - minVal) / MaxFloat64(1, maxVal-minVal) * float64(drawArea.Dy()-1))
				buf.SetCell(
					NewCell(self.DotMarkerRune, NewStyle(SelectColor(self.LineColors, i))),
					image.Pt(drawArea.Min.X+(j*self.HorizontalScale), drawArea.Max.Y-1-height),
				)
			}
		}
	}
}

func (self *Plot) plotAxes(buf *Buffer, minVal, maxVal float64) {
	// draw origin cell
	buf.SetCell(
		NewCell(BOTTOM_LEFT, NewStyle(ColorWhite)),
		image.Pt(self.Inner.Min.X+yAxisLabelsWidth, self.Inner.Max.Y-xAxisLabelsHeight-1),
	)
	// draw x axis line
	for i := yAxisLabelsWidth + 1; i < self.Inner.Dx(); i++ {
		buf.SetCell(
			NewCell(HORIZONTAL_DASH, NewStyle(ColorWhite)),
			image.Pt(i+self.Inner.Min.X, self.Inner.Max.Y-xAxisLabelsHeight-1),
		)
	}
	// draw y axis line
	for i := 0; i < self.Inner.Dy()-xAxisLabelsHeight-1; i++ {
		buf.SetCell(
			NewCell(VERTICAL_DASH, NewStyle(ColorWhite)),
			image.Pt(self.Inner.Min.X+yAxisLabelsWidth, i+self.Inner.Min.Y),
		)
	}
	// draw y axis labels
	verticalScale := (maxVal - minVal) / float64(self.Inner.Dy()-xAxisLabelsHeight-1)
	for i := 0; i*(yAxisLabelsGap+1) < self.Inner.Dy()-1; i++ {
		buf.SetString(
			fmt.Sprintf("%.2f", float64(i)*verticalScale*(yAxisLabelsGap+1)+minVal),
			NewStyle(ColorWhite),
			image.Pt(self.Inner.Min.X, self.Inner.Max.Y-(i*(yAxisLabelsGap+1))-2),
		)
	}
	switch self.PlotType {
	case ScatterPlot:
		for _, x := range self.Data[0] {
			self.XMinVal = MinFloat64(self.XMinVal, x)
			self.XMaxVal = MaxFloat64(self.XMaxVal, x)
		}

		for x := self.Inner.Min.X + yAxisLabelsWidth; x < self.Inner.Max.X-1; {
			index := (x - (self.Inner.Min.X + yAxisLabelsWidth)) / (self.HorizontalScale)
			label := fmt.Sprintf("%.02f", self.XMinVal+(float64(index)*(self.XMaxVal-self.XMinVal)/float64(self.Inner.Dx()-yAxisLabelsWidth-1)))
			if len(self.DataLabels) > index {
				label = fmt.Sprintf(
					"%s",
					self.DataLabels[index],
				)
			}
			buf.SetString(
				label,
				NewStyle(ColorWhite),
				image.Pt(x, self.Inner.Max.Y-1),
			)
			x += (len(label) + xAxisLabelsGap) * self.HorizontalScale
		}
	case LineChart:
		// draw x axis labels
		// draw first label or 0
		firstLabel := "0"
		if len(self.DataLabels) > 0 {
			firstLabel = self.DataLabels[0]
		}
		buf.SetString(
			firstLabel,
			NewStyle(ColorWhite),
			image.Pt(self.Inner.Min.X+yAxisLabelsWidth, self.Inner.Max.Y-1),
		)
		// draw rest
		for x := self.Inner.Min.X + yAxisLabelsWidth + (xAxisLabelsGap+len(firstLabel)-1)*self.HorizontalScale + 1; x < self.Inner.Max.X-1; {
			index := int((x-(self.Inner.Min.X+yAxisLabelsWidth)-1)/(self.HorizontalScale) + 1)
			label := fmt.Sprintf("%d", index)
			if len(self.DataLabels) > index {
				label = fmt.Sprintf(
					"%s",
					self.DataLabels[index],
				)
			}
			buf.SetString(
				label,
				NewStyle(ColorWhite),
				image.Pt(x, self.Inner.Max.Y-1),
			)
			x += (len(label) + xAxisLabelsGap) * self.HorizontalScale
		}
	}
}

func (self *Plot) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	currentMaxVal, _ := GetMaxFloat64From2dSlice(self.Data)
	self.MaxVal = MaxFloat64(self.MaxVal, currentMaxVal)

	currentMinVal, _ := GetMinFloat64From2dSlice(self.Data)
	self.MinVal = MinFloat64(currentMinVal, self.MinVal)

	if self.ShowAxes {
		self.plotAxes(buf, self.MinVal, self.MaxVal)
	}

	drawArea := self.Inner
	if self.ShowAxes {
		drawArea = image.Rect(
			self.Inner.Min.X+yAxisLabelsWidth+1, self.Inner.Min.Y,
			self.Inner.Max.X, self.Inner.Max.Y-xAxisLabelsHeight-1,
		)
	}

	switch self.Marker {
	case MarkerBraille:
		self.renderBraille(buf, drawArea, self.MinVal, self.MaxVal)
	case MarkerDot:
		self.renderDot(buf, drawArea, self.MinVal, self.MaxVal)
	}
}
