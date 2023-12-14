package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/sunshineplan/imgconv"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {

	zipfiles, err := filepath.Glob("./archive/*.zip")
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

	images, err := unzipAndCopy(zipFile, "zip2pdf-extracted")
	check(err)

	// tempDir, err := os.MkdirTemp("", "extracted")
	// check(err)
	// defer os.RemoveAll(tempDir)

	// zf, err := zip.OpenReader(zipFile)
	// check(err)
	// defer zf.Close()

	// var images []string
	// for _, file := range zf.File {
	// 	fPath := filepath.Join(tempDir, file.Name)

	// 	if file.FileInfo().IsDir() {
	// 		os.MkdirAll(fPath, os.ModePerm)
	// 	} else {
	// 		outFile, err := os.Create(fPath)
	// 		check(err)
	// 		defer outFile.Close()

	// 		rc, err := file.Open()
	// 		check(err)
	// 		defer rc.Close()

	// 		_, err = io.Copy(outFile, rc)
	// 		check(err)

	// 		if filepath.Ext(fPath) == ".jpg" || filepath.Ext(fPath) == ".png" {
	// 			images = append(images, fPath)
	// 		}
	// 	}
	// }
	createPDF(zipFile, images)
}

func unzipAndCopy(src, dest string) ([]string, error) {
	var images []string

	err := os.MkdirAll(dest, os.ModeDir)
	check(err)
	// tempDir, err := os.MkdirTemp("", dest)
	// check(err)
	// defer os.RemoveAll(tempDir)

	//open zip
	r, err := zip.OpenReader(src)
	check(err)
	defer r.Close()

	for _, f := range r.File {

		if IsExcludedFileOrDir(f.Name) {
			continue
		}

		if !utf8.ValidString(f.Name) {
			fname, err := ConvertToUtf8FromShiftJis(f.Name)
			check(err)
			f.Name = fname
		}

		if f.Mode().IsDir() {
			continue
		}

		if !isImageFile(f.Name) {
			continue
		}

		err := os.MkdirAll(filepath.Dir(dest), f.Mode())
		check(err)

		destPath := filepath.Join(dest, f.Name)

		// 脆弱性Zip Slipの対策 https://snyk.io/research/zip-slip-vulnerability#go
		if !strings.HasPrefix(destPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return images, fmt.Errorf("%s: illegal file path", destPath)
		}

		err = os.MkdirAll(filepath.Dir(destPath), f.Mode())
		check(err)

		if f.FileInfo().IsDir() {
			os.MkdirAll(f.Name, os.ModePerm)
		} else {
			images = append(images, destPath)
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		check(err)
		defer destFile.Close()

		rc, err := f.Open()
		check(err)
		defer rc.Close()

		_, err = io.Copy(destFile, rc)
		check(err)
	}
	return images, nil
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

// isImageFile 画像ファイルかどうかを判定する
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func ConvertToUtf8FromShiftJis(sjis string) (string, error) {
	if utf8.ValidString(sjis) {
		// 元々utf8のため変換しない
		return sjis, nil
	}
	utf8str, _, err := transform.String(japanese.ShiftJIS.NewDecoder(), sjis)
	return utf8str, err
}

// 解凍の対象外のファイル,ディレクトリかチェック
func IsExcludedFileOrDir(checkTarget string) bool {

	// macOS特有のディレクトリは除去
	if strings.HasPrefix(checkTarget, "__MACOSX") {
		return true
	}

	// macOS特有のファイルは除去
	if strings.Contains(checkTarget, ".DS_Store") {
		return true
	}

	return false
}

func replaceExt(file, to string) string {
	return strings.TrimSuffix(file, filepath.Ext(file)) + "." + to
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
