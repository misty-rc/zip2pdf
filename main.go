package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/signintech/gopdf"
)

func main() {

	zipFileName := "test.zip"
	extractPath := "."

	zipReader, err := zip.OpenReader(zipFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer zipReader.Close()

	var imageFiles []string
	for _, file := range zipReader.File {
		filePath := filepath.Join(extractPath, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if filepath.Ext(filePath) == ".jpg" || filepath.Ext(filePath) == ".png" {
			imageFiles = append(imageFiles, filePath)
			extractFile(file, filePath)
		}
	}

	sort.Strings(imageFiles)

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	for _, imgPath := range imageFiles {
		pdf.AddPage()
		if err := pdf.Image(imgPath, 0, 0, nil); err != nil {
			fmt.Println(err)
			return
		}
	}

	pdf.WritePdf("output.pdf")
}

func extractFile(zf *zip.File, dest string) {

	f, err := zf.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	outfile, err := os.Create(dest)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, f)
	if err != nil {
		fmt.Println(err)
	}
}
