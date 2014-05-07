package main

import (
	"flag"
	"fmt"
	"time"
	"sort"
	"strings"

	"os"
	"io/ioutil"
	"path/filepath"

	"hash"
	"math"
	//	"crypto"
	"crypto/md5"

	"github.com/cheggaaa/pb"
	"io"
	"net/http"
	"strconv"

	"encoding/csv"
)

const (
	TimeFormat                  = "2006-01-02 15:04:05"
	TimeFormatDate              = "20060102"
	TempDirPrefix               = "__wpchecuksum"
	WordpressChecksumGitURL     = "https://raw.githubusercontent.com/nroe/wpchecksum/master/checksum/md5/"
	WordpressPackageGitURL      = "https://github.com/WordPress/WordPress/archive/"
	WordpressPackageGitFilename = "%s.zip"

	ChecksumDefaultFilename = "%s.checksum.csv"

	FileChunk = 8192
)

type File struct {
    Name string
    Mode string
    ModTime string
}

func printUsage() {
	fmt.Println("wpchecksum --ver=[wordpress version] --dir=[wordpress root directory] --checksum=[checksum file,  will use system temporary directory default if not specified")
}

func main() {
	pVer := flag.String("ver", "", "wordpress version")
	pDirectory := flag.String("dir", "", "wordpress root directory")
	pChecksumFilename := flag.String("checksum", "", "checksum file")

	flag.Parse()

	if *pVer == "" {
		printUsage()
		return
	}
	
	checkDir, err := filepath.Abs(*pDirectory)
	if err != nil {
		die("> get directory absolute representation of path failed. message: %s", err)
	}

	_stat, _ := os.Stat(checkDir)
	if ! _stat.IsDir() {
		die ("> '%s' must be wordpress root directory", checkDir)
	}
	
	fmt.Printf("start check directory '%s' ...\n", checkDir)
	
	checksumDataFilename := strings.Trim(*pChecksumFilename, " ")
	if checksumDataFilename == "" {
		timenow := time.Now()
		tmpDir, err := ioutil.TempDir(os.TempDir(), TempDirPrefix + "_" + timenow.UTC().Format(TimeFormatDate) + "_");
		if err != nil {
			die("> creates a new temporary directory in the '%s' directory failed. message: %s", os.TempDir(), err)
		}
		
		checksumDataFilename = filepath.Join(tmpDir, *pVer)
		defer func() {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				die("> remove temporary directory failed. message: %s", err)
			}
		}()
	}
	
	if _, cdfErr := os.Stat(checksumDataFilename); cdfErr != nil && os.IsNotExist(cdfErr)  {
		checksumDataGitURL := WordpressChecksumGitURL + *pVer
	
		fmt.Printf("start download checksum data '%s' ...\n", checksumDataGitURL)
		downloadPackage(checksumDataGitURL, checksumDataFilename)
	}
	
	var checkFileNum int64 = 0
	diffFiles := []string{}
	diffFilesInfo := make(map[string]File)
	
	lostFiles := []string{}
	
	cd, err := os.Open(checksumDataFilename)
	if err != nil {
		die("> open file failed. message: %s", err)
	}
	defer cd.Close()
	csvReader := csv.NewReader(cd)
	for {
		line, err := csvReader.Read()
		
		if err == io.EOF {
			break
		} else if err != nil {
			die("> read line failed. message: %s", err)
		}
		
		checkFileNum ++
		
		f, err :=os.Open(filepath.Join(checkDir, line[0]))
		if (err != nil ) {
			if os.IsNotExist(err) {
				lostFiles = append(lostFiles, line[0])
				continue
			} else {
				die("> open file failed. message: %s", err)
			}
		}
		
		fi, err := f.Stat()
		if (err != nil) {
			die("> structure describing file can not be obtained. message: %s", err)
		}
		
		if fi.IsDir() {
			continue
		}
		
		size := fi.Size()
		chunks := uint64(math.Ceil(float64(size) / float64(FileChunk)))
		

		var h hash.Hash
		h = md5.New()

		for i := uint64(0); i < chunks; i++ {
			csize := int(math.Min(FileChunk, float64(size-int64(i*FileChunk))))
			buf := make([]byte, csize)

			f.Read(buf)
			io.WriteString(h, string(buf))
		}
		
		md5Sum := fmt.Sprintf("%x", h.Sum(nil))
		if line[1] != md5Sum {
			diffFiles = append(diffFiles, line[0])
			diffFilesInfo[line[0]] = File{fi.Name(), fi.Mode().String(), fi.ModTime().UTC().Format(TimeFormat)}
		}
		
		f.Close()
	}
	
	/**
	 * @see http://en.wikipedia.org/wiki/ANSI_escape_code
	 */ 
	fmt.Println("result:")
	fmt.Println("  \x1b[1;01mfile: ", checkFileNum,    "\x1b[0m")
	fmt.Println("  \x1b[1;4;31mdiff: ", len(diffFiles),    "\x1b[0m")
	fmt.Println("  \x1b[1;35mlost: ",   len(lostFiles) , "\x1b[0m")
	
	if len(diffFiles) > 0 {
		fmt.Println("diff files:")
		fmt.Println("----------------------------")
		
		sort.Strings(diffFiles)
		for _, path := range diffFiles {
			fmt.Println("    ", diffFilesInfo[path].ModTime, "\x1b[1;31m", path, "\x1b[0m")
		}
		fmt.Println("")
	}
	
	if len(lostFiles) > 0 {
		fmt.Println("lost files:")
		fmt.Println("----------------------------")
		
		sort.Strings(lostFiles)
		for _, path := range lostFiles {
			fmt.Println("    \x1b[1;35m", path, "\x1b[0m")
		}
		fmt.Println("")
	}
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
	bar.Finish()
}

func die(format string, v ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf("\x1b[1;4;31m"+format+"\x1b[0m\n", v...))
	os.Exit(1)
}
