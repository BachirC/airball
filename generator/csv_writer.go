package generator

import (
	"encoding/csv"
	"log"
	"os"
)

type CSVWriter struct {
	Path     string
	Filename string
	Header   []string
	Rows     [][]string
}

func (writer *CSVWriter) WriteToCSV() error {
	err := os.MkdirAll(writer.Path, os.ModePerm)
	if err != nil {
		log.Fatalln("failed to create parent folders", err)
	}
	f, err := os.Create(writer.Path + writer.Filename + ".csv")
	if err != nil {
		log.Fatalln("failed to create csv file", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write(writer.Header); err != nil {
		log.Fatalln("error writing header to file", err)
	}

	for _, r := range writer.Rows {
		if err := w.Write(r); err != nil {
			log.Fatalln("error writing record to file", err)
		}
	}

	return nil
}
