package main

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
)

type AbsolutePaths struct {
	AbsolutePaths []Path `json:"absolutePaths"`
}

type Path struct {
	Path string `json:"path"`
}

type Userdetail struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
	Dbname   string `json:"dbname"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("Starting the app")

	jsonFile, err := os.Open("users.json")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened json")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var aplha AbsolutePaths

	json.Unmarshal(byteValue, &aplha)

	var oldBackup = filepath.Dir("backup\\")

	os.RemoveAll(oldBackup)

	err = os.Mkdir("backup", 0755)
	check(err)

	var backup = filepath.Dir("backup\\")

	for i := 0; i < len(aplha.AbsolutePaths); i++ {
		fmt.Println("....Creatting Backup for" + aplha.AbsolutePaths[i].Path + "files")
		Dir(aplha.AbsolutePaths[i].Path, backup)
	}

	userFile, err := os.Open("newUser.json")
	fmt.Println(userFile)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened json")

	defer userFile.Close()

	userbyteValue, _ := ioutil.ReadAll(userFile)

	fmt.Println(userbyteValue)

	var user Userdetail

	// var user map[string]interface{}
	// json.Unmarshal([]byte(userbyteValue), &user)

	json.Unmarshal(userbyteValue, &user)
	fmt.Println(user)

	// fmt.Println("user.Username" + )
	// fmt.Println("user.Password" + user.Password)
	// fmt.Println("user.Hostname" + user.Hostname)
	// fmt.Println("user.Port" + user.Port)
	// fmt.Println("user.Dbname" + user.Dbname)

	username := "root"
	password := "mysqlpassword"
	hostname := "localhost"
	port := "3306"
	dbname := "backupfiles"

	dumpDir := "backup"                                             // you should create this directory
	dumpFilenameFormat := fmt.Sprintf("%s-20060102T150405", dbname) // accepts time layout string and add .sql at the end of file

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname))
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

	// Register database with mysqldump
	dumper, err := mysqldump.Register(db, dumpDir, dumpFilenameFormat)
	if err != nil {
		fmt.Println("Error registering databse:", err)
		return
	}

	// Dump database to file
	resultFilename, err := dumper.Dump()
	if err != nil {
		fmt.Println("Error dumping:", err)
		return
	}
	fmt.Printf("File is saved to %s", resultFilename)

	// Close dumper and connected database
	dumper.Close()

	//Zipping the file here
	var files []string

	root := "backup"
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	// for _, file := range files {
	// 	fmt.Println(file)
	// }

	output := "zip\\backup.zip"

	if err := zipit(root, output); err != nil {
		panic(err)
	}
	fmt.Println("Zipped File:", output)

}

func Dir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = Dir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = File(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func File(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
