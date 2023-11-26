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
	if err != nil {
		panic(err)
	}

	zipFileName := "(C78) (同人誌) [bolze.] Powerless Flower (ハヤテのごとく！).zip"
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

	//finish
	cleanUp(app)
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
