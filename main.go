package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/sunshineplan/imgconv"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {

	zipfiles, err := filepath.Glob("*.zip")
	check(err)

	var wg sync.WaitGroup
	for _, zipFile := range zipfiles {
		wg.Add(1)
		go func(zipFile string) {
			defer wg.Done()
			//処理関数
			extractAndCreatePDF(zipFile)
		}(zipFile)
	}
	wg.Wait()

}

func extractAndCreatePDF(zipFile string) {
	tempDir, err := os.MkdirTemp("", "extracted")
	check(err)
	defer os.RemoveAll(tempDir)

	zf, err := zip.OpenReader(zipFile)
	check(err)
	defer zf.Close()

	var images []string
	for _, file := range zf.File {
		fPath := filepath.Join(tempDir, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(fPath, os.ModePerm)
		} else {
			outFile, err := os.Create(fPath)
			check(err)
			defer outFile.Close()

			rc, err := file.Open()
			check(err)
			defer rc.Close()

			_, err = io.Copy(outFile, rc)
			check(err)

			if filepath.Ext(fPath) == ".jpg" || filepath.Ext(fPath) == ".png" {
				images = append(images, fPath)
			}
		}
	}
	createPDF(zipFile, images)
}

func createPDF(zipFile string, images []string) {

	var pdfarray []string

	//PDF filename
	pdfFile := strings.TrimSuffix(zipFile, filepath.Ext(zipFile)) + ".pdf"
	sort.Slice(images, func(i, j int) bool {
		return images[i] < images[j]
	})

	//image file convert to pdf
	for _, img := range images {
		src, err := imgconv.Open(img)
		check(err)

		//debug
		// fmt.Println(img)

		fpdf := replaceExt(img, "pdf")
		pdfarray = append(pdfarray, fpdf)
		ff, err := os.Create(fpdf)
		check(err)
		defer ff.Close()

		err = imgconv.Write(ff, src, &imgconv.FormatOption{Format: imgconv.PDF})
		check(err)
	}

	sort.Slice(pdfarray, func(i, j int) bool {
		return pdfarray[i] < pdfarray[j]
	})
	err := api.MergeCreateFile(pdfarray, pdfFile, api.LoadConfiguration())
	check(err)

}

func replaceExt(file, to string) string {
	return strings.TrimSuffix(file, filepath.Ext(file)) + "." + to
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
