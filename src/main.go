package main

import (
	"fmt"
	"strconv"

	"strings"

	"github.com/klauspost/reedsolomon"
	"github.com/pkg/errors"
)

var (
	CHARACTER_SET = []byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D',
		'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V',
		'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
		'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', ' ', '#',
	}
	N_ENCODING_TABLE = []string{
		"00", "01", "02", "10", "11", "12", "20", "21", "22", "30",
	}
	C_ENCODING_TABLE = []string{
		"222", "300", "301", "302", "310", "311", "312", "320", "321", "322",
		"000", "001", "002", "010", "011", "012", "020", "021", "022", "100", "101", "102", "110",
		"111", "112", "120", "121", "122", "200", "201", "202", "210", "211", "212", "220", "221",
		"023", "030", "031", "032", "033", "103", "113", "123", "130", "131", "132", "133", "203",
		"213", "223", "230", "231", "232", "233", "303", "313", "323", "330", "331", "332", "333",
		"003", "013",
	}

	BAR_VALUE_TABLE = []string{
		"000", "001", "002", "003", "010", "011", "012", "013", "020", "021",
		"022", "023", "030", "031", "032", "033", "100", "101", "102", "103", "110", "111", "112",
		"113", "120", "121", "122", "123", "130", "131", "132", "133", "200", "201", "202", "203",
		"210", "211", "212", "213", "220", "221", "222", "223", "230", "231", "232", "233", "300",
		"301", "302", "303", "310", "311", "312", "313", "320", "321", "322", "323", "330", "331",
		"332", "333",
	}

	BARCODE_TYPE_STANDARD int64 = 11
	BARCODE_TYPE_TWO      int64 = 59
	BARCODE_TYPE_THREE    int64 = 62
)

func main() {
	var customer_max_digits int
	var customer_max_chars int
	var customerMaxBars int

	input := "5956439111ABA 9"
	fcc, err := strconv.ParseInt(input[0:2], 10, 64)
	if err != nil {
		fmt.Print(err)
	}

	dpid, err := strconv.ParseInt(input[2:10], 10, 64)
	if err != nil {
		fmt.Print(err)
	}

	switch fcc {
	case BARCODE_TYPE_STANDARD:
		customer_max_digits = 0
		customer_max_chars = 0
		customerMaxBars = 0
	case BARCODE_TYPE_TWO:
		customer_max_digits = 8
		customer_max_chars = 5
		customerMaxBars = 16
	case BARCODE_TYPE_THREE:
		customer_max_digits = 15
		customer_max_chars = 10
		customerMaxBars = 31
	default:
		fmt.Println("error, unknown fcc")
	}

	customerInfo := input[10:]
	encodeTable := C_ENCODING_TABLE
	_, err = strconv.ParseFloat(customerInfo, 64)
	if err == nil {
		if len(customerInfo) > customer_max_digits {
			fmt.Println("error lenght for digits")
		}
		encodeTable = N_ENCODING_TABLE
	} else {
		if len(customerInfo) > customer_max_chars {
			fmt.Println("error lenght for chars")
		}
	}

	charPosition := func(value byte, array []byte) (int, error) {
		for i := 0; i < len(array); i++ {
			if value == array[i] {
				return i, nil
			}
		}

		return 0, errors.New("Unable to find character " + string(value) + " in character array.")
	}

	encodeString := func(s string, encodeTable []string) (r string) {
		for i := 0; i < len(s); i++ {
			p, err := charPosition(byte(s[i]), CHARACTER_SET)
			if err != nil {
				fmt.Println(err)
			}
			r += encodeTable[p]
		}
		return
	}

	var customerInfoEncoded string
	fccEncoded := encodeString(strconv.FormatInt(fcc, 10), N_ENCODING_TABLE)
	dpidEncoded := encodeString(strconv.FormatInt(dpid, 10), N_ENCODING_TABLE)

	if len(customerInfo) > 0 {
		customerInfoEncoded = encodeString(customerInfo, encodeTable)
		customerInfoEncoded += strings.Repeat("3", customerMaxBars-len(customerInfoEncoded))
	}

	encodedValues := fccEncoded + dpidEncoded + customerInfoEncoded
	triples := []uint8{}

	for i := 0; i < len(encodedValues); i += 3 {
		triple := encodedValues[i : i+3]
		first := (triple[0] - '0') << 4
		second := (triple[1] - '0') << 2
		third := triple[2] - '0'

		value := first + second + third
		triples = append(triples, value)
	}

	//bs := []byte(strconv.Itoa(27))

	enc, err := reedsolomon.New(len(triples), 4)
	shards, err := enc.Split(triples)
	fmt.Printf("File split into %d data+parity shards with %d bytes/shard.\n", len(shards), len(shards[0]))
	err = enc.Encode(shards)
	ok, err := enc.Verify(shards)
	fmt.Println(ok, err)
	fmt.Println(shards[14])
}
