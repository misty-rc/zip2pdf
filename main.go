package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/signintech/gopdf"
)

var workDir string = "zip2pdf-work"
var srcDir string = "zip2pdf-src"
var destDir string = "zip2pdf-dest"

type AppContext struct {
	home string //home dir
	src  string //dir include zip files
	dest string //pdf output dir
	work string //work dir(tmp)
}

func main() {
	//初期設定
	app := initializer()

	zipfiles, err := filepath.Glob("*.zip")
	check(err)

	var wg sync.WaitGroup
	for _, zipFile := range zipfiles {
		wg.Add(1)
		go func(zipFile string) {
			defer wg.Done()
			//処理関数
			extractAndCreatePDF(app, zipFile)
		}(zipFile)
	}
	wg.Wait()

	//finish
	cleanUp(app)
}

func extractAndCreatePDF(app AppContext, zipFile string) {
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
	createPDF2(images, strings.TrimSuffix(zipFile, ".zip")+".pdf")
}

func createPDF(images []string, pdfFileName string) {

}

func createPDF2(images []string, pdfFileName string) {
	// pdf := gofpdf.New("P", "mm", "A4", "")
	pdf := gopdf.GoPdf{}
	// pdf.Start(gopdf.Config{})
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeB4})
	for _, img := range images {

		pdf.AddPageWithOption(gopdf.PageOption{PageSize: &gopdf.Rect{W: 1147, H: 1600}})
		// pdf.AddPage()
		pdf.Image(img, 0, 0, nil)
	}

	// err := pdf.OutputFileAndClose(pdfFileName)
	err := pdf.WritePdf(pdfFileName)
	check(err)
	fmt.Printf("PDF created: %s\n", pdfFileName)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func initializer() AppContext {
	//初期化
	home := os.Getenv("HOME")
	if len(home) == 0 {
		//home not set
		fmt.Printf("init error: %s", "HOMEが設定されていません")
	}
	tmp := os.Getenv("TMP")
	if len(tmp) == 0 {
		//tmp not set
		fmt.Printf("init error: %s", "TMPが設定されていません")
	}

	//オプションで任意設定できるようにする予定
	context := AppContext{
		home: home,
		src:  filepath.Join(home, srcDir),
		dest: filepath.Join(home, destDir),
		work: filepath.Join(tmp, workDir)}

	//作業ディレクトリ作成
	os.Mkdir(context.work, os.ModePerm)

	return context
}

func cleanUp(context AppContext) {
	//掃除
	err := os.Remove(context.work)
	if err != nil {
		fmt.Printf("clean up error: %s", "作業フォルダ削除に失敗しました")
	}
}
