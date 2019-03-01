package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log"

	"github.com/bitrise-io/go-utils/cmdex"
)

type Message struct {
	ReleaseDate   string
	MD5_Hash      string
	Major_Version string
	Minor_Version string
	Build_Number  string
	File_Name     string
}

type ConfigsModel struct {
	ReleaseDate     string
	Version         string
	BuildNumber     string
	FileName        string
	FilePath        string
	DestinationPath string
	SkipReleaseDate string
}

func createConfigsModelFromEnvs() ConfigsModel {

	ret := ConfigsModel{
		Version:         os.Getenv("version_string"),
		BuildNumber:     os.Getenv("build_number"),
		FilePath:        os.Getenv("file_path"),
		DestinationPath: os.Getenv("destination_path"),
		SkipReleaseDate: os.Getenv("skip_release_date"),
	}
	ret.FileName = filepath.Base(ret.FilePath)

	return ret
}

func (configs ConfigsModel) print() {
	fmt.Println()
	log.Printf("Configs:")
	log.Printf(" - Version: %s \n", configs.Version)
	log.Printf(" - Build Number: %s \n", configs.BuildNumber)
	log.Printf(" - file path: %s \n", configs.FilePath)
	log.Printf(" - Filename: %s \n", configs.FileName)
	log.Printf(" - Skip Release date: %s \n", configs.SkipReleaseDate)
}

func (configs ConfigsModel) validate() error {
	// required
	if configs.Version == "" {
		return errors.New("No Version parameter specified!")
	}
	if configs.BuildNumber == "" {
		return errors.New("No Build number specified!")
	}
	if configs.FileName == "" {
		return errors.New("No filename specified")
	}

	if configs.DestinationPath == "" {
		return errors.New("No destination file path specified")
	}

	return nil
}

func exportEnvironmentWithEnvman(keyStr, valueStr string) error {
	cmd := cmdex.NewCommand("envman", "add", "--key", keyStr)
	cmd.SetStdin(strings.NewReader(valueStr))
	return cmd.Run()
}

func md5sum(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return
	}

	result = strings.ToUpper(hex.EncodeToString(hash.Sum(nil)))
	return
}

func main() {
	configs := createConfigsModelFromEnvs()
	configs.print()
	if err := configs.validate(); err != nil {
		log.Fatalf("Issue with input: %s", err)
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter Release date (DD.MM.YYYY hh:mm) :")

		if configs.SkipReleaseDate == "false" {
			text, _ := reader.ReadString('\n')
			if text == "\n" {
				configs.ReleaseDate = time.Now().Format("2006-01-02T15:04:05.00Z")
				break
			}

			date, error := time.Parse("02.01.2006 15:04\n", text)

			if error != nil {
				log.Printf("wrong date input: %s\n", error.Error())
			} else {
				configs.ReleaseDate = date.Format("2006-01-02T15:04:05.000Z")
				break
			}

		} else {
			break
		}

	}

	var toSerialize Message

	splits := strings.Split(configs.Version, ".")

	toSerialize.Build_Number = configs.BuildNumber
	toSerialize.Major_Version = splits[0]
	toSerialize.File_Name = configs.FileName
	toSerialize.ReleaseDate = configs.ReleaseDate
	if len(splits) > 1 {
		toSerialize.Minor_Version = configs.Version[len(splits[0])+1 : len(configs.Version)]
	}

	var error error

	toSerialize.MD5_Hash, error = md5sum(configs.FilePath)

	if error != nil {
		log.Fatalf("Error in md5 hashing: %s\n", error.Error())

	}

	fmt.Printf("Major Version: %s\n", toSerialize.Major_Version)
	fmt.Printf("Minor version: %s\n", toSerialize.Minor_Version)
	fmt.Printf("Build Number: %s\n", toSerialize.Build_Number)
	fmt.Printf("MD5 hash: %s\n", toSerialize.MD5_Hash)
	fmt.Printf("Release date: %s\n", toSerialize.ReleaseDate)
	fmt.Printf("Filename: %s\n", toSerialize.File_Name)

	var marshalled []byte

	marshalled, error = json.MarshalIndent(toSerialize, "", "   ")

	if error != nil {
		log.Fatalf("error in marshalling: %s \n", error.Error())
	}

	error = ioutil.WriteFile(configs.DestinationPath, marshalled, 0644)

	if error != nil {
		log.Fatalf("error in marshalling to file: %s\n", error.Error())
	}

}
