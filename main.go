// Copyright 2011-2015, The xlsx2csv Authors.
// All rights reserved.
// For details, see the LICENSE file.

package main

import (
	"encoding/csv"
	"strings"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tealeg/xlsx/v3"
)

func generateCSVFromXLSXFile(w io.Writer, excelFileName string, sheetIndex int, delimiter string, csvOpts csvOptSetter) error {
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		return err
	}
	sheetLen := len(xlFile.Sheets)
	switch {
	case sheetLen == 0:
		return errors.New("This XLSX file contains no sheets.")
	case sheetIndex >= sheetLen:
		return fmt.Errorf("No sheet %d available, please select a sheet between 0 and %d\n", sheetIndex, sheetLen-1)
	}
	cw := csv.NewWriter(w)
	if csvOpts != nil {
		csvOpts(cw)
	}
	sheet := xlFile.Sheets[sheetIndex]
	var vals []string
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		if row != nil {
			vals = vals[:0]
			err := row.ForEachCell(func(cell *xlsx.Cell) error {
				str, err := cell.FormattedValue()
				str = strings.Replace(str, delimiter, "_sys_temp_delimiter_temp_sys_", -1)
				str = strings.Replace(str, "\n",      "_sys_temp_newline_temp_sys_",   -1)
				str = strings.Replace(str, "\r",      "",   -1)
				if err != nil {
					return err
				}
				vals = append(vals, str)
				return nil
			})
			if err != nil {
				return err
			}
		}
		cw.Write(vals)
		return nil
	})
	if err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

type csvOptSetter func(*csv.Writer)

func main() {
	var (
		outFile    = flag.String("o", "-", "filename to output to. -=stdout")
		sheetIndex = flag.Int("i", 0, "Index of sheet to convert, zero based")
		delimiter  = flag.String("d", "\t", "Delimiter to use between fields")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `%s
	dumps the given xlsx file's chosen sheet as a CSV,
	with the specified delimiter, into the specified output.

Usage:
	%s [flags] <xlsx-to-be-read>
`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	out := os.Stdout
	if !(*outFile == "" || *outFile == "-") {
		var err error
		if out, err = os.Create(*outFile); err != nil {
			log.Fatal(err)
		}
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			log.Fatal(closeErr)
		}
	}()

	if err := generateCSVFromXLSXFile(out, flag.Arg(0), *sheetIndex, *delimiter,
		func(cw *csv.Writer) { cw.Comma = ([]rune(*delimiter))[0] },
	); err != nil {
		log.Fatal(err)
	}
}
