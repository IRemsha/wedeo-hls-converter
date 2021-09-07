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

const pathToOutputDir = "app/output"
const minioPreffix = "videos/"

func goDotEnvVariable(key string) string {
    err := godotenv.Load(".env")
    if err != nil {
      log.Fatalf("Error loading .env file")
    }
    return os.Getenv(key)
}

func ConvertVideo(pathToFile string, pathToOutputDir string) error {
    pathToOutputFile := pathToOutputDir + "/output.m3u8"
	cmd := exec.Command(
		"ffmpeg",
		"-i", pathToFile,
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
            var pathToNewFile = val.S3.Object.Key
            var fileName = path.Base(pathToNewFile)

            pathToOutputFile := fmt.Sprintf("app/input/%s, fileName")
            err = minioClient.FGetObject(goDotEnvVariable("MINIO_BUCKET"), pathToNewFile, pathToOutputFile, minio.GetObjectOptions{})
            if err != nil {
                log.Println(err)
                return
            }
            log.Printf("Success download file %s to %s\n", fileName, pathToOutputFile)

            videoTitle := strings.Split(fileName, ".")[0]
            
            start := time.Now()
            log.Println("Start convert video:")
            err := ConvertVideo(pathToOutputFile, pathToOutputDir)
            if err != nil {
                log.Println(err)
                return
            }
            elapsed := time.Since(start)
            log.Printf("Complete convert %s in %s\n", fileName, elapsed.String())

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
