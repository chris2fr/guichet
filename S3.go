package main

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

	"github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nfnt/resize"
)

const PROFILE_PICTURE_FIELD_NAME = "profilePicture"

//Upload image through guichet server.
func uploadImage(w http.ResponseWriter, r *http.Request, login *LoginStatus) (string, error) {
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

	buff := bytes.NewBuffer([]byte{})
	buff_thumbnail := bytes.NewBuffer([]byte{})
	err = resizeThumb(file, buff, buff_thumbnail)
	if err != nil {
		return "", err
	}

	mc, err := newMinioClient()
	if err != nil || mc == nil {
		return "", err
	}

	var name, nameFull string

	if nameConsul := login.UserEntry.GetAttributeValue(PROFILE_PICTURE_FIELD_NAME); nameConsul != "" {
		name = nameConsul
		nameFull = "full_" + name
	} else {
		name = uuid.New().String() + ".jpeg"
		nameFull = "full_" + name
	}

	_, err = mc.PutObject(context.Background(), config.S3_Bucket, name, buff_thumbnail, int64(buff_thumbnail.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	_, err = mc.PutObject(context.Background(), config.S3_Bucket, nameFull, buff, int64(buff.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	return name, nil
}

func newMinioClient() (*minio.Client, error) {
	endpoint := config.S3_Endpoint
	accessKeyID := config.S3_AccesKey
	secretKeyID := config.S3_SecretKey
	useSSL := true

	//Initialize Minio
	minioCLient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKeyID, ""),
		Secure: useSSL,
		Region: config.S3_Region,
	})

	if err != nil {
		return nil, err
	}

	return minioCLient, nil

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
	switch fileType {
	case "image":
		return nil
	default:
		return errors.New("bad type")
	}

}

func resizeThumb(file multipart.File, buff, buff_thumbnail *bytes.Buffer) error {
	file.Seek(0, 0)
	images, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	buff.Reset()
	images = resize.Thumbnail(200, 200, images, resize.Lanczos3)
	images_thumbnail := resize.Thumbnail(80, 80, images, resize.Lanczos3)

	err = jpeg.Encode(buff, images, &jpeg.Options{
		Quality: 95,
	})
	if err != nil {
		return err
	}

	err = jpeg.Encode(buff_thumbnail, images_thumbnail, &jpeg.Options{
		Quality: 95,
	})

	return err
}

func handleDownloadImage(w http.ResponseWriter, r *http.Request) {
	//Get input value by user
	dn := mux.Vars(r)["name"]
	size := mux.Vars(r)["size"]
	//Check login
	login := checkLogin(w, r)
	if login == nil {
		return
	}
	var imageName string
	if dn != "unknown_profile" {
		//Search values with ldap and filter

		searchRequest := ldap.NewSearchRequest(
			dn,
			ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
			"(objectclass=*)",
			[]string{PROFILE_PICTURE_FIELD_NAME},
			nil)

		sr, err := login.conn.Search(searchRequest)
		if err != nil {
			http.Error(w, "Search: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if len(sr.Entries) != 1 {
			http.Error(w, fmt.Sprintf("Not found user: %s cn: %s and numberEntries: %d", dn, strings.Split(dn, ",")[0], len(sr.Entries)), http.StatusInternalServerError)
			return
		}
		imageName = sr.Entries[0].GetAttributeValue(PROFILE_PICTURE_FIELD_NAME)
		if imageName == "" {
			http.Error(w, "User doesn't have profile image", http.StatusNotFound)
			return
		}
	} else {
		imageName = "unknown_profile.jpg"
	}

	if size == "full" {
		imageName = "full_" + imageName
	}
	//Get the object after connect MC
	mc, err := newMinioClient()
	if err != nil {
		http.Error(w, "MinioClient: "+err.Error(), http.StatusInternalServerError)
		return
	}
	obj, err := mc.GetObject(context.Background(), "bottin-pictures", imageName, minio.GetObjectOptions{})
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
