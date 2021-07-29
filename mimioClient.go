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

//Upload image through guichet server.
func uploadImage(w http.ResponseWriter, r *http.Request, login *LoginStatus) (bool, string, error) {
	file, _, err := r.FormFile("image")

	if err == http.ErrMissingFile {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	defer file.Close()
	fileType, err := checkImage(file)
	if err != nil {
		return false, "", err
	}
	if fileType == "" {
		return false, "", nil
	}

	buff := bytes.NewBuffer([]byte{})
	buff_thumbnail := bytes.NewBuffer([]byte{})
	err = resizeThumb(file, buff, buff_thumbnail)
	if err != nil {
		return false, "", err
	}

	mc, err := newMimioClient()
	if err != nil {
		return false, "", err
	}
	if mc == nil {
		return false, "", err
	}

	var name, nameFull string

	if nameConsul := login.UserEntry.GetAttributeValue("profilImage"); nameConsul != "" {
		name = nameConsul
		nameFull = "full_" + name
	} else {
		name = uuid.New().String() + ".jpeg"
		nameFull = "full_" + name
	}

	_, err = mc.PutObject(context.Background(), "bottin-pictures", name, buff_thumbnail, int64(buff_thumbnail.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return false, "", err
	}

	_, err = mc.PutObject(context.Background(), "bottin-pictures", nameFull, buff, int64(buff.Len()), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return false, "", err
	}

	return true, name, nil
}

func newMimioClient() (*minio.Client, error) {
	endpoint := config.Endpoint
	accessKeyID := config.AccesKey
	secretKeyID := config.SecretKey
	useSSL := true

	//Initialize Minio
	minioCLient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKeyID, ""),
		Secure: useSSL,
		Region: "garage",
	})

	if err != nil {
		return nil, err
	}

	return minioCLient, nil

}

func checkImage(file multipart.File) (string, error) {
	buff := make([]byte, 512) //Detect read only the first 512 bytes
	_, err := file.Read(buff)
	if err != nil {
		return "", err
	}
	file.Seek(0, 0)

	fileType := http.DetectContentType(buff)
	fileType = strings.Split(fileType, "/")[0]
	switch fileType {
	case "image":
		return fileType, nil
	default:
		return "", errors.New("bad type")
	}

}

func resizeThumb(file multipart.File, buff, buff_thumbnail *bytes.Buffer) error {
	file.Seek(0, 0)
	images, _, err := image.Decode(file)
	if err != nil {
		return errors.New("Decode: " + err.Error())
	}
	//Encode image to jpeg a first time to eliminate all problems
	err = jpeg.Encode(buff, images, &jpeg.Options{
		Quality: 100, //Between 1 to 100, higher is better
	})
	if err != nil {
		return err
	}
	images, _, err = image.Decode(buff)
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
			[]string{"profilImage"},
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
		imageName = sr.Entries[0].GetAttributeValue("profilImage")
		if imageName == "" {
			http.Error(w, "User doesn't have profile image", http.StatusInternalServerError)
			return
		}
	} else {
		imageName = "unknown_profile.jpg"
	}

	if size == "full" {
		imageName = "full_" + imageName
	}

	//Get the object after connect MC
	mc, err := newMimioClient()
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
		http.Error(w, "MinioObjet: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//Send JSON through xhttp
	w.Header().Set("Content-Type", objStat.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(int(objStat.Size)))
	//http.Error(w, fmt.Sprintf("Length buffer: %d", objStat.Size), http.StatusInternalServerError)
	buff := make([]byte, objStat.Size)

	obj.Seek(0, 0)
	n, err := obj.Read(buff)
	if err != nil && err != io.EOF {
		http.Error(w, fmt.Sprintf("Read Error: %s, bytes Read: %d, bytes in file: %d", err.Error(), n, objStat.Size), http.StatusInternalServerError)
		return
	}
	if int64(n) != objStat.Size {
		http.Error(w, fmt.Sprintf("Read %d bytes on %d bytes", n, objStat.Size), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(buff); err != nil {
		http.Error(w, "WriteBody: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
