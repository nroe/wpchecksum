package main

import (
	"flag"
	"fmt"
	"time"

/*	"io/ioutil"*/
	"os"
	"path/filepath"

/*	"hash"*/
/*	"math"*/
	//	"crypto"
/*	"crypto/md5"*/

	"github.com/cheggaaa/pb"
	"io"
	"net/http"
	"strconv"

/*	"encoding/csv"*/
)

const (
	TimeFormatDate              = "20060102"
	TempDirPrefix               = "__wpchecuksum"
	WordpressPackageGitURL      = "https://github.com/WordPress/WordPress/archive/"
	WordpressPackageGitFilename = "%s.zip"

	ChecksumDefaultFilename = "%s.checksum.csv"

	FileChunk = 8192
)


func main() {
	pVer := flag.String("ver", "", "wordpress version")
	pDirectory := flag.String("d", "", "check directory")

	flag.Parse()

	if *pVer == "" {
		printUsage()
		return
	}

    checkData := "/Users/wlnroe/workspaces/github/wpchecksum/3.9.csv"


	checkDir, err := filepath.Abs(*pDirectory)
	if err != nil {
		die("> get directory absolute representation of path failed. message: %s", err)
	}

	_stat, _ := os.Stat(checkDir)
	if ! _stat.IsDir() {
        die ("> %s must be wordpress root directory", checkDir)
	}

    fmt.

}


func downloadPackage(sourceName string, destName string) {
	var source io.Reader
	var sourceSize int64

	// open as url
	resp, err := http.Get(sourceName)
	if err != nil {
		die("> Can't get %s: %v\n", sourceName, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		die("> Server return non-200 status: %v\n", resp.Status)
		return
	}
	i, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	sourceSize = int64(i)
	source = resp.Body

	// create dest
	dest, err := os.Create(destName)
	if err != nil {
		die("> Can't create %s: %v\n", destName, err)
		return
	}
	defer dest.Close()

	// create bar
	bar := pb.New(int(sourceSize)).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.Start()

	// create multi writer
	writer := io.MultiWriter(dest, bar)

	// and copy
	io.Copy(writer, source)
	bar.FinishPrint("Download complete .")
}

func die(format string, v ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf("\x1b[1;4;31m"+format+"\x1b[0m\n", v...))
	os.Exit(1)
}

func printUsage() {
	fmt.Println("wpchecksum -ver=[wordpress version] -d=[check directory]")
}
