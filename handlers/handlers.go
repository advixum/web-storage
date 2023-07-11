package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	db "spa-api/database"
	"spa-api/logging"
	"spa-api/models"
	"strings"
	"sync"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var log = logging.Config

// Login handler for gin-jwt/v2 middleware.
func LogIn(c *gin.Context) (interface{}, error) {
	var loginVals models.User
	if err := c.ShouldBind(&loginVals); err != nil {
		log.WithFields(logrus.Fields{
			"username": loginVals.Username,
			"password": loginVals.Password,
		}).Error(logging.F()+"() parsing error:", err)
		return "", jwt.ErrMissingLoginValues
	}
	username := loginVals.Username
	password := loginVals.Password
	log.WithFields(logrus.Fields{
		"username": username,
		"password": password,
	}).Debug(logging.F() + "() login values:")
	var entry models.User
	dbReq := db.C.Where("username = ?", username).First(&entry)
	if dbReq.Error != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	err := bcrypt.CompareHashAndPassword(
		[]byte(entry.Password),
		[]byte(password),
	)
	if err != nil {
		return nil, jwt.ErrFailedAuthentication
	}
	return &entry, nil
}

// Add additional payload data to the webtoken of gin-jwt/v2
// middleware. Return map with user ID.
func Payload(data interface{}) jwt.MapClaims {
	if v, ok := data.(*models.User); ok {
		log.WithFields(logrus.Fields{
			"ID": v.ID,
		}).Debug(logging.F() + "() ID value")
		return jwt.MapClaims{
			"id": v.ID,
		}
	}
	return jwt.MapClaims{}
}

// Sign up handler. Clears user input, hashes the password, creates a
// new entry in the database. Return a message about the result of
// data processing.
func SignUp(c *gin.Context) {
	var regVals models.User
	if err := c.ShouldBind(&regVals); err != nil {
		log.WithFields(logrus.Fields{
			"username": regVals.Username,
			"password": regVals.Password,
		}).Error(logging.F()+"() parsing error:", err)
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Failed to register. Fields missing."},
		)
		return
	}
	user, pass := regVals.Username, regVals.Password
	log.WithFields(logrus.Fields{
		"username": user,
		"password": pass,
	}).Debug(logging.F() + "() register data:")
	if user == "" {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Fields cannot be empty."},
		)
		return
	}
	if !checkPass(pass) {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": `
				The password must be at least 16 characters long, contain at
				least one lowercase letter, one uppercase letter, one number,
				and one special character.
			`},
		)
		return
	}
	hashedPass, err := bcrypt.GenerateFromPassword(
		[]byte(pass), bcrypt.DefaultCost,
	)
	if err != nil {
		log.WithFields(logrus.Fields{
			"password": pass,
		}).Error(logging.F()+"() bcrypt hashing error:", err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{"message": "Failed to register. Password problem."},
		)
		return
	}
	entry := models.User{
		Username: user,
		Password: string(hashedPass),
	}
	dbReq := db.C.Create(&entry)
	if dbReq.Error != nil {
		c.JSON(
			http.StatusConflict,
			gin.H{"message": "Failed to register. This user exists."},
		)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success registration."})
}

// Checks that the password consists only of ASCII characters, is at
// least 16 characters long, contains at least one lowercase letter,
// one uppercase letter, one digit, and one special character. Returns
// true if the password matches the requirements, false otherwise.
func checkPass(password string) bool {
	if len(password) < 16 {
		return false
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false
	for _, char := range password {
		switch {
		case 'a' <= char && char <= 'z':
			hasLower = true
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case '0' <= char && char <= '9':
			hasDigit = true
		default:
			if char < 32 || char > 126 {
				return false
			}
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSpecial
}

// Return a map with a list of files for a specific user. Data can be
// sorted in ascending and descending order by a column name.
func List(c *gin.Context) {
	sortOrd := c.Query("ord")
	sortCol := c.Query("col")
	log.WithFields(logrus.Fields{
		"sortOrd": sortOrd,
		"sortCol": sortCol,
	}).Debug(logging.F() + "() sorting values:")
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))
	//userID = claims["id"].(uint) // Panic log test
	log.WithFields(logrus.Fields{
		"ID": userID,
	}).Debug(logging.F() + "() token ID value:")
	var files []models.File
	dbReq := db.C.Where("user_id = ?", userID).Find(&files)
	if dbReq.Error != nil {
		log.Error(logging.F()+"() cannot find userID:", dbReq.Error)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Can't find user."})
		return
	}
	switch sortCol {
	case "ListName":
		sort.Slice(files, func(i, j int) bool {
			if sortOrd == "asc" {
				return files[i].ListName < files[j].ListName
			} else {
				return files[i].ListName > files[j].ListName
			}
		})
	case "Extension":
		sort.Slice(files, func(i, j int) bool {
			if sortOrd == "asc" {
				return files[i].Extension < files[j].Extension
			} else {
				return files[i].Extension > files[j].Extension
			}
		})
	case "Date":
		sort.Slice(files, func(i, j int) bool {
			if sortOrd == "asc" {
				return files[i].Date.Before(files[j].Date)
			} else {
				return files[i].Date.After(files[j].Date)
			}
		})
	case "Size":
		sort.Slice(files, func(i, j int) bool {
			if sortOrd == "asc" {
				return files[i].Size < files[j].Size
			} else {
				return files[i].Size > files[j].Size
			}
		})
	default:
		log.WithFields(logrus.Fields{
			"sortCol": sortCol,
		}).Error(logging.F() + "() unexpected column name:")
	}
	log.WithFields(logrus.Fields{
		"files": files,
	}).Debug(logging.F() + "() files list from DB:")
	c.JSON(http.StatusOK, gin.H{"files": files})
}

// Return the specified by file ID and user ID file into the body
// stream.
func Download(c *gin.Context) {
	fileID := c.Query("id")
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))
	log.WithFields(logrus.Fields{
		"ID": userID,
	}).Debug(logging.F() + "() token ID value:")
	var entry models.File
	dbReq := db.C.First(&entry, "id = ? AND user_id = ?", fileID, userID)
	if dbReq.Error != nil {
		log.Error(logging.F()+"() cannot find entry:", dbReq.Error)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Can't find an entry."})
		return
	}
	log.WithFields(logrus.Fields{
		"file": entry,
	}).Debug(logging.F() + "() entry for Blob response:")
	c.File(entry.Path)
}

// Checks the uniqueness of the file in the user's folder, saves the
// file and creates an entry in the database. Return a message about
// the result of data processing.
func Upload(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))
	log.WithFields(logrus.Fields{
		"ID": userID,
	}).Debug(logging.F() + "() token ID value:")
	userDir := ""
	if gin.Mode() == gin.TestMode {
		userDir = filepath.Join("./upload/test/", fmt.Sprintf("%d", userID))
	} else {
		userDir = filepath.Join("./upload/", fmt.Sprintf("%d", userID))
	}
	log.WithFields(logrus.Fields{
		"userDir": userDir,
	}).Debug(logging.F() + "() personal user directory: ")
	form, err := c.MultipartForm()
	if err != nil {
		log.Error(logging.F()+"() multipart form parse error:", err)
		return
	}
	files := form.File["files"]
	log.WithFields(logrus.Fields{
		"files": files,
	}).Debug(logging.F() + "() files list from MultipartForm:")
	loadList := []string{}
	var tasksGroup sync.WaitGroup
	chSemaphore := make(chan int, 3)
	var status int
	for _, file := range files {
		tasksGroup.Add(1)
		chSemaphore <- 1
		go func(
			file *multipart.FileHeader,
			ch chan int,
			tg *sync.WaitGroup,
		) {
			defer tg.Done()
			fileBase := filepath.Base(file.Filename)
			filePath := filepath.Join(userDir, fileBase)
			fileExt := filepath.Ext(fileBase)
			fileName := strings.TrimSuffix(fileBase, fileExt)
			num := 1
			for {
				_, err := os.Stat(filePath)
				if err != nil {
					if os.IsNotExist(err) {
						break
					} else {
						log.Error(logging.F()+"() file check error:", err)
						break
					}
				} else {
					numStr := fmt.Sprintf("(%d)", num)
					numName := fileName + numStr
					numPath := numName + fileExt
					filePath = filepath.Join(userDir, numPath)
					fileBase = filepath.Base(filePath)
					num++
				}
			}
			log.WithFields(logrus.Fields{
				"fileBase": fileBase,
				"filePath": filePath,
				"fileExt":  fileExt,
				"fileName": fileName,
			}).Debug(logging.F() + "() name parts and path:")
			entry := models.File{
				UserID:    userID,
				ListName:  strings.TrimSuffix(fileBase, fileExt),
				Name:      fmt.Sprintf("/%d/", userID) + fileBase,
				Extension: fileExt,
				Path:      filePath,
				Date:      time.Now(),
				Size:      file.Size,
			}
			dbReq := db.C.Create(&entry)
			if dbReq.Error != nil {
				log.Errorf(
					logging.F()+"() %v is not unique: %v",
					fileBase,
					dbReq.Error,
				)
				status = http.StatusConflict
				loadList = append(loadList, fileBase+" FAILED!")
				return
			}
			c.SaveUploadedFile(file, filePath)
			loadList = append(loadList, fileBase)
			<-ch
		}(file, chSemaphore, &tasksGroup)
	}
	tasksGroup.Wait()
	message := "Loaded files: "
	length := len(loadList)
	for i, val := range loadList {
		if i < length-1 {
			message = message + val + ", "
		} else {
			message = message + val
		}
	}
	if status == 0 {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{"message": message})
}

// Changes the Name and ListName entries for the specified user file.
// Return a message about the result of data processing.
func Rename(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))
	log.WithFields(logrus.Fields{
		"ID": userID,
	}).Debug(logging.F() + "() token ID value:")
	var renameVals models.File
	if err := c.ShouldBind(&renameVals); err != nil {
		log.WithFields(logrus.Fields{
			"ID":        renameVals.ID,
			"Name":      renameVals.Name,
			"Extension": renameVals.Extension,
		}).Error(logging.F()+"() parsing error:", err)
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Cannot rename a file."},
		)
		return
	}
	log.WithFields(logrus.Fields{
		"ID":        renameVals.ID,
		"Name":      renameVals.Name,
		"Extension": renameVals.Extension,
	}).Debug(logging.F() + "() renaming values:")
	dbReq := db.C.Model(&models.File{}).
		Where("id = ? AND user_id = ?", renameVals.ID, userID).
		Updates(map[string]interface{}{
			"name": fmt.Sprintf(
				"/%d/%s%s",
				userID,
				renameVals.Name,
				renameVals.Extension,
			),
			"list_name": renameVals.Name,
		})
	if dbReq.Error != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": fmt.Sprintf(
					"File: \"%s%s\" was exist. Enter another name.",
					renameVals.Name,
					renameVals.Extension,
				),
			},
		)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// Deletes a file from the user's folder and its record from the
// database. Return a message about the result of data processing.
func Delete(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	userID := uint(claims["id"].(float64))
	log.WithFields(logrus.Fields{
		"ID": userID,
	}).Debug(logging.F() + "() token ID value:")
	var delVals models.File
	if err := c.ShouldBind(&delVals); err != nil {
		log.WithFields(logrus.Fields{
			"ID": delVals.ID,
		}).Error(logging.F()+"() parsing error:", err)
		c.JSON(
			http.StatusBadRequest,
			gin.H{"message": "Cannot delete a file."},
		)
		return
	}
	log.WithFields(logrus.Fields{
		"ID": delVals.ID,
	}).Debug(logging.F() + "() deleting ID:")
	var entry models.File
	dbReq := db.C.First(&entry, "id = ? AND user_id = ?", delVals.ID, userID)
	if dbReq.Error != nil {
		log.Error(logging.F()+"() cannot find entry:", dbReq.Error)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Can't find an entry."})
		return
	}
	err := os.Remove(entry.Path)
	if err != nil {
		log.Error(logging.F()+"() removing error: ", err)
		return
	}
	db.C.Unscoped().Delete(&entry)
	c.JSON(http.StatusOK, gin.H{})
}
