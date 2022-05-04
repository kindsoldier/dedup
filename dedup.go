/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
 * MA 02110-1301, USA.
 *
 */

package main

import (
    "errors"
    "encoding/hex"
    "flag"
    "fmt"
    "io"
    "os"
    "math"
    "regexp"
    "path/filepath"
    "strings"
    "strconv"

    "github.com/minio/highwayhash"
)

const successExit int = 0
const errorExit int = 1

func main() {
    var err error
    app := NewApp()
    err = app.GetOptions()
    if err != nil {
        fmt.Println(err)
        os.Exit(errorExit)
    }
    app.SearchDups()
}

type FileMap = map[string]*File

type App struct {
    Dirs []string
    MaxDepth int
    HMinSize string
    HMaxSize string
    MinSize int64
    MaxSize int64
    Limit int64
    HLimit string
    FileColl FileMap
    Pattern string

    RunQuiet bool
    PrintStats bool
    DoLink bool
    DoBackup bool
}

type File struct {
    Name string
    Size int64
}

func NewApp() *App {
    var app App
    app.MaxDepth = 5
    app.HMinSize = "1Kb"
    app.HMaxSize = "1Gb"
    app.Dirs = make([]string, 0, 10)
    app.FileColl = make(FileMap, 0)
    app.HLimit = "0b"
    app.Pattern = "*"

    app.RunQuiet = false
    app.PrintStats = false
    app.DoLink = false
    app.DoBackup = false
    return &app
}



func (this *App) GetOptions() error {
    var err error
    exeName := filepath.Base(os.Args[0])

    flag.StringVar(&this.Pattern, "pat", this.Pattern, "file name pattern")
    flag.StringVar(&this.HLimit, "limit", this.HLimit, "read only first bytes")
    flag.StringVar(&this.HMinSize, "min", this.HMinSize, "minimal file size")
    flag.StringVar(&this.HMaxSize, "max", this.HMaxSize, "maximal file size")
    flag.IntVar(&this.MaxDepth, "depth", this.MaxDepth, "maximal depth")
    flag.BoolVar(&this.RunQuiet, "quiet", this.RunQuiet, "supress listing")
    flag.BoolVar(&this.PrintStats, "stats", this.PrintStats, "print addtional stats")
    flag.BoolVar(&this.DoLink, "link", this.DoLink, "do hard link of duplicate, use by carefull")
    flag.BoolVar(&this.DoBackup, "bak", this.DoBackup, "rename duplicate before link")

    help := func() {
        fmt.Println("")
        fmt.Printf("Usage: %s [option] dirs...\n", exeName)
        fmt.Println("")
        fmt.Println("Options:")
        flag.PrintDefaults()
        fmt.Println("")
    }
    flag.Usage = help
    flag.Parse()

    this.Dirs = flag.Args()

    if len(this.Dirs) == 0 {
        //pwd, _ := os.Getwd()
        //this.Dirs = append(this.Dirs, pwd)
        help()
        os.Exit(successExit)
    }

    this.MinSize, err = UnhumanSize(this.HMinSize)
    if err != nil {
        return err
    }
    this.MaxSize, err = UnhumanSize(this.HMaxSize)
    if err != nil {
        return err
    }
    this.Limit, err = UnhumanSize(this.HLimit)
    if err != nil {
        return err
    }
    return err
}


func (this *App) SearchDups() error {
    var err error
    var sizeStat int64
    var countStat int64

    const hwInit string = "000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000"
    hwInitBytes, err := hex.DecodeString(hwInit)

    callback := func(filename string) {

        fileInfo, err := os.Lstat(filename)
        if err != nil {
            return
        }
        fd, err := os.Open(filename)
        if err != nil {
            return
        }
        filesize := fileInfo.Size()
        if filesize < this.MinSize {
            return
        }
        if filesize > this.MaxSize {
            return
        }
        if (fileInfo.Mode() & os.ModeSymlink) == os.ModeSymlink {
            return
        }
        if !fileInfo.Mode().IsRegular() {
            return
        }
        matched, err := filepath.Match(this.Pattern, filepath.Base(filename))
        if !matched {
            return
        }

        if !this.RunQuiet {
            fmt.Printf("%s", filename)
            defer fmt.Printf("\n")
        }

        hasher, err := highwayhash.New(hwInitBytes)
        if err != nil {
            return
        }

        var hashSum string
        if this.Limit == 0 {
            buffer := make([]byte, 1024 * 4)
            for {
                rSize, err := fd.Read(buffer)
                if err == io.EOF {
                    break
                }
                hasher.Write(buffer[0:rSize])
            }
            hashSum = string(hasher.Sum(nil))
        } else {
            remain := this.Limit
            var total int = 0
            const bufferSize int = 1024 * 4
            buffer := make([]byte, bufferSize)

            for {
                nBuffer := buffer
                if remain < int64(bufferSize) {
                    nBuffer = buffer[0:remain]
                }
                rSize, err := fd.Read(nBuffer)
                if remain < 1 {
                    break
                }
                if err == io.EOF {
                    break
                }
                remain -= int64(rSize)
                total += rSize
                hasher.Write(buffer[0:rSize])
            }
            hashSum = string(hasher.Sum(nil))
        }

        prevFile, ok := this.FileColl[hashSum]

        if ok && prevFile.Size == filesize {
            if !this.RunQuiet {
                fmt.Printf(" == %s", prevFile.Name)
            }
            sizeStat += prevFile.Size
            countStat++
            if this.DoLink {
                var err error
                if this.DoBackup {
                    err = os.Rename(filename, filename + ".dedup")
                } else {
                    err = os.Remove(filename)
                }
                if err == nil {
                    os.Link(prevFile.Name, filename)
                }
            }
            return
        }
        this.FileColl[hashSum] = &File{
            Name: filename,
            Size: filesize,
        }
    }

    for _, dir := range this.Dirs {
        err = ScanTree(dir, this.MaxDepth, callback)
        if err != nil {
            continue
        }
    }
    if this.PrintStats {
        fmt.Printf("found duplicates %d with total size %d bytes\n", countStat, sizeStat)
    }
    return err
}

type nameFunc = func(filename string)

func ScanTree(basePath string, depth int, callback nameFunc) error {
    var err error

    const pathSeparator string = "/"

    pathLength := func(path string) int {
        if len(path) == 0 {
            return 0
        }
        path = filepath.Clean(path)
        path = filepath.ToSlash(path)
        path = strings.Trim(path, pathSeparator)
        if len(path) == 0 {
            return 0
        }
        list := strings.Split(path, pathSeparator)
        return len(list)
    }

    depth = depth + pathLength(basePath)

    resolver := func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if pathLength(fullPath) > depth  {
            return filepath.SkipDir
        }
        if !info.IsDir(){
            callback(fullPath)
        }
        return nil
    }
    err = filepath.Walk(basePath, resolver)
    if err != nil {
        return err
    }
    return err
}

func UnhumanSize(hSize string) (int64, error) {
    var err error
    var size int64
    hSize = strings.ToLower(hSize)

    var multi int = 1
    reg := regexp.MustCompile("([01-9.]+)(kb|mb|gb|b|k|m|g)")
    strArr := reg. FindStringSubmatch(hSize)

    if len(strArr) > 2 {
        hSize = strArr[1]
        hMulti := strArr[2]
        switch hMulti {
            case "b": multi = 1
            case "kb", "k": multi = 1024
            case "mb", "m": multi = 1024 * 1024
            case "gb", "g": multi = 1024 * 1024 * 1024
            case "tb", "t": multi = 1024 * 1024 * 1024 * 1024
            default:
                err = errors.New("unkn size multi")
                return size, err
        }
    }
    flSize, err := strconv.ParseFloat(hSize, 32)
    size = int64(math.Round(flSize * float64(multi)))
    return size, err
}
