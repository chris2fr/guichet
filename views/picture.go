package views

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"image"
	"image/jpeg"
	_ "image/png"

	"mime/multipart"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nfnt/resize"
)

func newMinioClient() (*minio.Client, error) {
	endpoint := config.S3Endpoint
	accessKeyID := config.S3AccessKey
	secretKeyID := config.S3SecretKey
	useSSL := true

	//Initialize Minio
	minioCLient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKeyID, ""),
		Secure: useSSL,
		Region: config.S3Region,
	})

	if err != nil {
		return nil, err
	}

	return minioCLient, nil
}

// Upload image through guichet server.
func uploadProfilePicture(w http.ResponseWriter, r *http.Request, login *LoginStatus) (string, error) {
	file, _, err := r.FormFile("image")

	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer file.Close()

	err = checkImage(file)
	if err != nil {
		return "", err
	}

	buffFull := bytes.NewBuffer([]byte{})
	buffThumb := bytes.NewBuffer([]byte{})
	err = resizePicture(file, buffFull, buffThumb)
	if err != nil {
		return "", err
	}

	mc, err := newMinioClient()
	if err != nil || mc == nil {
		return "", err
	}

	// If a previous profile picture existed, delete it
	// (don't care about errors)
	if nameConsul := login.UserEntry.GetAttributeValue(FIELD_NAME_PROFILE_PICTURE); nameConsul != "" {
		mc.RemoveObject(context.Background(), config.S3Bucket, nameConsul, minio.RemoveObjectOptions{})
		mc.RemoveObject(context.Background(), config.S3Bucket, nameConsul+"-thumb", minio.RemoveObjectOptions{})
	}

	// Generate new random name for picture
	nameFull := uuid.New().String()
	nameThumb := nameFull + "-thumb"

	_, err = mc.PutObject(context.Background(), config.S3Bucket, nameThumb, buffThumb, int64(buffThumb.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	_, err = mc.PutObject(context.Background(), config.S3Bucket, nameFull, buffFull, int64(buffFull.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	return nameFull, nil
}

func checkImage(file multipart.File) error {
	buff := make([]byte, 512) //Detect read only the first 512 bytes
	_, err := file.Read(buff)
	if err != nil {
		return err
	}
	file.Seek(0, 0)

	fileType := http.DetectContentType(buff)
	fileType = strings.Split(fileType, "/")[0]
	if fileType != "image" {
		return errors.New("bad type")
	}

	return nil
}

func resizePicture(file multipart.File, buffFull, buffThumb *bytes.Buffer) error {
	file.Seek(0, 0)
	picture, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	thumbnail := resize.Thumbnail(90, 90, picture, resize.Lanczos3)
	picture = resize.Thumbnail(480, 480, picture, resize.Lanczos3)

	err = jpeg.Encode(buffFull, picture, &jpeg.Options{
		Quality: 95,
	})
	if err != nil {
		return err
	}

	err = jpeg.Encode(buffThumb, thumbnail, &jpeg.Options{
		Quality: 100,
	})

	return err
}

func HandleDownloadPicture(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	//Check login
	login := checkLogin(w, r)
	if login == nil {
		return
	}

	//Get the object after connect MC
	mc, err := newMinioClient()
	if err != nil {
		http.Error(w, "MinioClient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	obj, err := mc.GetObject(context.Background(), "bottin-pictures", name, minio.GetObjectOptions{})
	if err != nil {
		http.Error(w, "MinioClient: GetObject: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	objStat, err := obj.Stat()
	if err != nil {
		http.Error(w, "MiniObjet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//Send JSON through xhttp
	w.Header().Set("Content-Type", objStat.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(int(objStat.Size)))
	//Copy obj in w
	writting, err := io.Copy(w, obj)

	if writting != objStat.Size || err != nil {
		http.Error(w, fmt.Sprintf("WriteBody: %s, bytes wrote %d on %d", err.Error(), writting, objStat.Size), http.StatusInternalServerError)
		return
	}

}
