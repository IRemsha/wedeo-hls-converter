package main

import (
    "log"
	"os"
    "fmt"
    "time"
    "path"
    "io/ioutil"
    "strings"
	"os/exec"
    "github.com/minio/minio-go"
    "github.com/joho/godotenv"
)

const pathToInputDir = "/app/input"
const pathToOutputDir = "/app/output"
const minioPreffix = "videos/"

func goDotEnvVariable(key string) string {
    err := godotenv.Load(".env")
    if err != nil {
      log.Fatalf("Error loading .env file")
    }
    return os.Getenv(key)
}

func ensureDir(dirName string) error {
    err := os.Mkdir(dirName, os.ModeDir)
    if err == nil || os.IsExist(err) {
        return nil
    } else {
        return err
    }
}

func ConvertVideo(pathToFile string, pathToOutputDir string) error {
    if err := ensureDir(pathToOutputDir); err != nil {
        log.Printf("Directory creation failed with error: %s", err.Error())
        os.Exit(1)
    }
    pathToOutputFile := pathToOutputDir + "/output.m3u8"
	cmd := exec.Command(
		"ffmpeg",
		"-i", pathToFile,
		"-c:a", "aac", "-ac",  "2",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		pathToOutputFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
    endpoint := goDotEnvVariable("MINIO_ENDPOINT")
    minioClient, err := minio.New(endpoint, goDotEnvVariable("MINIO_ACCESS_KEY_ID"), goDotEnvVariable("MINIO_SECRET_ACCESS_KEY"), true)
	if err != nil {
        log.Println(err)
        return
	}
    log.Printf("Connected to %s\n", endpoint)
    

    doneCh := make(chan struct{})
    for notificationInfo := range minioClient.ListenBucketNotification(goDotEnvVariable("MINIO_BUCKET"), minioPreffix, "", []string{
        "s3:ObjectCreated:Put",
        }, doneCh) {
        if notificationInfo.Err != nil {
            log.Println(notificationInfo.Err)
        }
        
        for _, val := range notificationInfo.Records {
            var pathToNewFileInMinio = val.S3.Object.Key
            var fileName = path.Base(pathToNewFileInMinio)

            localPathToNewFile := fmt.Sprintf("%s/%s", pathToInputDir, fileName)
            err = minioClient.FGetObject(goDotEnvVariable("MINIO_BUCKET"), pathToNewFileInMinio, localPathToNewFile, minio.GetObjectOptions{})
            if err != nil {
                log.Println(err)
                return
            }
            log.Printf("Success download file %s to %s\n", fileName, localPathToNewFile)

            videoTitle := strings.Split(fileName, ".")[0]
            
            start := time.Now()
            log.Println("Start convert video:")
            err := ConvertVideo(localPathToNewFile, pathToOutputDir)
            if err != nil {
                log.Println(err)
                return
            }
            elapsed := time.Since(start)
            log.Printf("Complete convert %s in %s\n", fileName, elapsed.String())

            os.RemoveAll(localPathToNewFile)
            log.Printf("File %s delete", localPathToNewFile)

            log.Printf("Start upload %s\n", fileName)
            files, err := ioutil.ReadDir(pathToOutputDir)
            if err != nil {
                log.Fatal(err)
            }
            for _, f := range files {
                _, err := minioClient.FPutObject(
                    goDotEnvVariable("MINIO_BUCKET"),
                    fmt.Sprintf("%s/%s/%s", goDotEnvVariable("MINIO_UPLOAD_FOLDER"), videoTitle, f.Name()),
                    fmt.Sprintf("%s/%s", pathToOutputDir, f.Name()),
                    minio.PutObjectOptions{});
                if err != nil {
                    log.Println(err)
                    return
                }
                log.Printf("File %s/%s upload\n", videoTitle, f.Name())

                os.RemoveAll(path.Join([]string{pathToOutputDir, f.Name()}...))
                log.Printf("File %s/%s delete\n", pathToOutputDir, f.Name())
            }
            log.Printf("Upload %s/%s complete\n", videoTitle, fileName)
        }
    }
}
