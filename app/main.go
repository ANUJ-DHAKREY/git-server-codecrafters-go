package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":

		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		readFlag := os.Args[2];
		if(readFlag != "-p") {
			fmt.Println("please pass -p flag to read the object content");
		}
		hashId := os.Args[3];

		if len(hashId) != 40 {
			fmt.Println("please pass a valid sha ID to read the object content");
		}

		workingDirectory, err := os.Getwd(); if err != nil {
			panic(err);
		}

		gitRootDir := getRepoRootDir(workingDirectory);
		subDir := hashId[:2];
		fileName := hashId[2:];
		filePath :=  filepath.Join(gitRootDir,".git","objects", subDir,fileName);
		data , err := os.ReadFile(filePath); if err != nil {
			panic(err);
		}

		val,err := zlib.NewReader(bytes.NewReader(data)); if err != nil {
			panic(err);
		}

		defer val.Close();
		data, err = io.ReadAll(val); if err != nil {
			panic(err)
		}
		data = extractContent(data);
		fmt.Print(string(data))
	case "hash-object" : 

		readFlag := os.Args[2];

		if(readFlag != "-w") {
			fmt.Println("please pass -w flag to generate sha and write the object content");
		}

		filePath := os.Args[3];

		f,err := os.ReadFile(filePath); if err != nil {
			panic(err);
		}
		contentLength := []byte(strconv.Itoa(len(f)));
		//"blob "
		blob := []byte {98, 108, 111, 98, 32};
		content := append(blob,contentLength...);
		content =  append(content,0);
		content = append(content,f...);
		shaObject := sha1.New();	
		shaObject.Write(content);
		generatedHash := hex.EncodeToString(shaObject.Sum(nil));
		currDir, err := os.Getwd(); if err != nil {
			panic(err);
		}
		filePath = getRepoRootDir(currDir);
		subDir := generatedHash[:2];
		fileName := generatedHash[2:];
		filePath = filepath.Join(filePath,".git/objects",subDir,fileName);
		dir := filepath.Dir(filePath); 
		err = os.MkdirAll(dir,0755); if err != nil {
			panic(err);
		} 
		writeFileErr := os.WriteFile(filePath,content,0644); if writeFileErr != nil {
			panic(writeFileErr);
		}
		fmt.Print(generatedHash);
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func extractContent(data []byte) []byte {
	i := len("blob ");
	for ;i < len(data); i++ {
		if data[i] == 0 {
			i++;
			break;
		}
	}

	return data[i:];
}

func getRepoRootDir(directoryPath string) string{
	rootDir := filepath.VolumeName(directoryPath);
	if rootDir == "" {
		rootDir = "/";
	} else{
		rootDir += filepath.Join(rootDir,"");
	}
	if directoryPath == rootDir {
		fmt.Println("fatal: not a git repository (or any of the parent directories): .git");
		os.Exit(1);
	}

	directoryList,err := os.ReadDir(directoryPath); if err != nil {
		panic(err);
	}

	for _,dir := range directoryList {
		if(dir.IsDir() && dir.Name() == ".git") {
			return directoryPath;
		}	
	}
	getRepoRootDir(filepath.Dir(directoryPath));	
	return "";
}
