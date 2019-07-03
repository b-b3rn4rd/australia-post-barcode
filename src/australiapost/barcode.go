package australiapost

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ajstarks/svgo"
	"github.com/pkg/errors"
)

var (
	characterSet = []byte{
		'0', '1', '2', '3', '4', '5', '6', '7',
		'8', '9', 'A', 'B', 'C', 'D', 'E', 'F',
		'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N',
		'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V',
		'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd',
		'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l',
		'm', 'n', 'o', 'p', 'q', 'r', 's', 't',
		'u', 'v', 'w', 'x', 'y', 'z', ' ', '#',
	}
	nEncodingTable = []string{
		"00", "01", "02", "10", "11",
		"12", "20", "21", "22", "30",
	}
	cEncodingTable = []string{
		"222", "300", "301", "302", "310", "311", "312", "320",
		"321", "322", "000", "001", "002", "010", "011", "012",
		"020", "021", "022", "100", "101", "102", "110", "111",
		"112", "120", "121", "122", "200", "201", "202", "210",
		"211", "212", "220", "221", "023", "030", "031", "032",
		"033", "103", "113", "123", "130", "131", "132", "133",
		"203", "213", "223", "230", "231", "232", "233", "303",
		"313", "323", "330", "331", "332", "333", "003", "013",
	}

	barValueTable = []string{
		"000", "001", "002", "003", "010", "011", "012", "013",
		"020", "021", "022", "023", "030", "031", "032", "033",
		"100", "101", "102", "103", "110", "111", "112", "113",
		"120", "121", "122", "123", "130", "131", "132", "133",
		"200", "201", "202", "203", "210", "211", "212", "213",
		"220", "221", "222", "223", "230", "231", "232", "233",
		"300", "301", "302", "303", "310", "311", "312", "313",
		"320", "321", "322", "323", "330", "331", "332", "333",
	}
)

const (
	barcodeTypeStandard           int64  = 11
	barcodeTypeTwo                int64  = 59
	barcodeTypeThree              int64  = 62
	barcodeDefaultFontSize        int    = 10
	barcodeDefaultPadding         int    = 6
	barcodeDefaultFontColor       string = "black"
	barcodeDefaultBackgroundColor string = "white"
	barcodeDefaultFont            string = "Courier"
)

type Barcode interface {
	Generate() error
}

type Logger interface {
	Printf(string, ...interface{})
}

type fourStateBarcode struct {
	input           string
	encoder         Encoder
	wr              io.Writer
	text            string
	logger          Logger
	padding         int
	fontSize        int
	barRatio        int
	backgroundColor string
	fontColor       string
}

type option func(b *fourStateBarcode)

func OptionPadding(padding int) option {
	return func(b *fourStateBarcode) {
		b.padding = padding
	}
}

func OptionLogger(logger Logger) option {
	return func(b *fourStateBarcode) {
		b.logger = logger
	}
}

func OptionRatio(ratio int) option {
	return func(b *fourStateBarcode) {
		b.barRatio = ratio
	}
}

func OptionFontSize(fontSize int) option {
	return func(b *fourStateBarcode) {
		b.fontSize = fontSize
	}
}

func OptionBackgroundColor(color string) option {
	return func(b *fourStateBarcode) {
		b.backgroundColor = color
	}
}

func OptionFontColor(color string) option {
	return func(b *fourStateBarcode) {
		b.fontColor = color
	}
}

func OptionalEncoder(encoder Encoder) option {
	return func(b *fourStateBarcode) {
		b.encoder = encoder
	}
}

func NewFourStateBarcode(input string, wr io.Writer, text string, options ...option) Barcode {
	barcode := &fourStateBarcode{
		input:           input,
		wr:              wr,
		text:            text,
		padding:         barcodeDefaultPadding,
		fontSize:        barcodeDefaultFontSize,
		barRatio:        2,
		encoder:         NewReedSolomon(),
		backgroundColor: barcodeDefaultBackgroundColor,
		fontColor:       barcodeDefaultFontColor,
	}

	for _, option := range options {
		option(barcode)
	}

	return barcode
}
func (b *fourStateBarcode) Generate() error {
	var customerMaxDigits int
	var customerMaxChars int
	var customerMaxBars int
	var mandatoryFillers int

	fcc, err := strconv.ParseInt(b.input[0:2], 10, 64)
	if err != nil {
		return errors.Wrap(err, "error while extracting the control code field")
	}

	dpid, err := strconv.ParseInt(b.input[2:10], 10, 64)
	if err != nil {
		return errors.Wrap(err, "error while extracting the delivery point id field")
	}

	switch fcc {
	case barcodeTypeStandard:
		customerMaxDigits = 0
		customerMaxChars = 0
		customerMaxBars = 0
		mandatoryFillers = 1
	case barcodeTypeTwo:
		customerMaxDigits = 8
		customerMaxChars = 5
		customerMaxBars = 16
		mandatoryFillers = 0
	case barcodeTypeThree:
		customerMaxDigits = 15
		customerMaxChars = 10
		customerMaxBars = 31
		mandatoryFillers = 0
	default:
		return errors.Errorf("unknown or unsupported control code field %d", fcc)
	}

	customerInfo := b.input[10:]
	encodeTable := cEncodingTable
	_, err = strconv.ParseFloat(customerInfo, 64)
	if err == nil {
		if len(customerInfo) > customerMaxDigits {
			return errors.Errorf("invalid digits length for the customer information field %d, however %d expected",
				len(customerInfo), customerMaxDigits)
		}
		encodeTable = nEncodingTable
	} else {
		if len(customerInfo) > customerMaxChars {
			return errors.Errorf("invalid char length for the customer information field %d, however %d expected",
				len(customerInfo), customerMaxChars)
		}
	}

	charPosition := func(value byte, array []byte) (int, error) {
		for i := 0; i < len(array); i++ {
			if value == array[i] {
				return i, nil
			}
		}

		return 0, errors.New("unable to find character " + string(value) + " in bytes array.")
	}

	encodeString := func(s string, encodeTable []string) (r string, err error) {
		for i := 0; i < len(s); i++ {
			p, err := charPosition(byte(s[i]), characterSet)
			if err != nil {
				return "", errors.Wrap(err, "error while doing a byte lookup")
			}
			r += encodeTable[p]
		}
		return
	}

	var customerInfoEncoded string

	fccEncoded, err := encodeString(strconv.FormatInt(fcc, 10), nEncodingTable)
	if err != nil {
		return errors.Wrap(err, "error while encoding control code field")
	}

	dpidEncoded, err := encodeString(strconv.FormatInt(dpid, 10), nEncodingTable)
	if err != nil {
		return errors.Wrap(err, "error while encoding delivery point id field")
	}

	if len(customerInfo) > 0 {
		customerInfoEncoded, err = encodeString(customerInfo, encodeTable)
		if err != nil {
			return errors.Wrap(err, "error while encoding customer information field")
		}

		customerInfoEncoded += strings.Repeat("3", customerMaxBars-len(customerInfoEncoded))
	}

	customerInfoEncoded += strings.Repeat("3", mandatoryFillers)

	encodedValues := fccEncoded + dpidEncoded + customerInfoEncoded

	var triples []uint

	for i := 0; i < len(encodedValues); i += 3 {
		triple := encodedValues[i : i+3]
		first := (triple[0] - '0') << 4
		second := (triple[1] - '0') << 2
		third := triple[2] - '0'

		value := first + second + third
		triples = append(triples, uint(value))
	}

	parityValues := b.encoder.Encode(triples)

	for i := 0; i < len(parityValues); i++ {
		encodedValues += barValueTable[parityValues[i]]
	}

	encodedValues = fmt.Sprintf("13%s13", encodedValues)

	var barcodeWidth, barcodeHeight int
	textHeight := 0

	if b.text != "" {
		textHeight = b.fontSize
	}

	switch fcc {
	case barcodeTypeStandard:
		barcodeWidth = 146
		barcodeHeight = 8*b.barRatio + b.padding + b.padding + textHeight
	case barcodeTypeTwo:
		barcodeWidth = 206
		barcodeHeight = 8*b.barRatio + b.padding + b.padding + textHeight
	case barcodeTypeThree:
		barcodeWidth = 266
		barcodeHeight = 8*b.barRatio + b.padding + b.padding + textHeight
	}

	b.draw(encodedValues, barcodeWidth, barcodeHeight, textHeight)

	return nil
}

func (b *fourStateBarcode) draw(encodedValues string, barcodeWidth int, barcodeHeight int, textHeight int) error {
	var barWidth, barHeight, barXpos, barYpos int

	canvas := svg.New(b.wr)
	canvas.Start(barcodeWidth, barcodeHeight)
	barWidth = 2
	barXpos = 0
	canvas.Rect(0, 0, barcodeWidth, barcodeHeight, fmt.Sprintf("fill:%s", b.backgroundColor))

	for i := 0; i < len(encodedValues); i++ {
		switch string(encodedValues[i]) {
		case "0":
			barYpos = 0
			barHeight = 8
		case "1":
			barYpos = 0
			barHeight = 5
		case "2":
			barYpos = 3
			barHeight = 5
		case "3":
			barYpos = 3
			barHeight = 2
		}

		canvas.Roundrect(barXpos,
			(barYpos*b.barRatio)+b.padding+textHeight, barWidth, barHeight*b.barRatio,
			1,
			1,
			"fill:black")

		barXpos += barWidth * 2
	}

	if b.text != "" {
		canvas.Text(0,
			b.fontSize,
			b.text,
			fmt.Sprintf("font-size:%dpx; fill:%s", b.fontSize, b.fontColor),
			fmt.Sprintf("font-family=\"%s\"", barcodeDefaultFont))
	}

	canvas.End()

	return nil
}
