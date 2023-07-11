package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	db "spa-api/database"
	"spa-api/middleware"
	"spa-api/models"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// Requirements: .env PostgreSQL credentials, "test" database, RSA keys

/*
    Основной целью данного тестирования является закрепление
    усвоенных знаний на практике, а не исчерпывающий охват всех 
    возможных тестовых случаев.
*/

// Testing for inbound data in the handlers.SignUp() function.
func TestSignUp(t *testing.T) {
	type args struct {
		username string
		password string
		validity bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Valid data was accepted and password was hashed",
			args: args{
				username: "testuser",
				password: "abcdEFGH1234!@#$",
				validity: true,
			},
		},
		{
			name: "Missing username was not accepted",
			args: args{
				password: "abcdEFGH1234!@#$",
				validity: false,
			},
		},
		{
			name: "Missing password was not accepted",
			args: args{
				username: "testuser",
				validity: false,
			},
		},
		{
			name: "An empty string in the username was not accepted",
			args: args{
				username: "",
				password: "abcdEFGH1234!@#$",
				validity: false,
			},
		},
		{
			name: "Less than 16-chars password was not accepted",
			args: args{
				username: "testuser",
				password: "abcdEFGH1234!@#",
				validity: false,
			},
		},
		{
			name: "Longer than 72-chars password was not accepted",
			args: args{
				username: "testuser",
				password: "aB!73Chars73Chars73Chars73Chars73Chars73Chars73Chars73Chars73Chars73Chars",
				validity: false,
			},
		},
		{
			name: "Password without lowercase chars was not accepted",
			args: args{
				username: "testuser",
				password: "ABCDEFGH1234!@#$",
				validity: false,
			},
		},
		{
			name: "Password without uppercase chars was not accepted",
			args: args{
				username: "testuser",
				password: "abcdefgh1234!@#$",
				validity: false,
			},
		},
		{
			name: "Password without numbers was not accepted",
			args: args{
				username: "testuser",
				password: "abcdEFGHHHHH!@#$",
				validity: false,
			},
		},
		{
			name: "Password without special symbols was not accepted",
			args: args{
				username: "testuser",
				password: "abcdEFGH12344444",
				validity: false,
			},
		},
		{
			name: "Password with non-ASCII symbols was not accepted",
			args: args{
				username: "testuser",
				password: "фbcdEFGH1234!@#$",
				validity: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			gin.SetMode(gin.TestMode)
			db.Connect()
			db.C.AutoMigrate(&models.User{}, &models.File{})
			defer db.C.Migrator().DropTable(
				&models.User{},
				&models.File{},
			)

			// Create testing data
			send := models.User{
				Username: tt.args.username,
				Password: tt.args.password,
			}
			jsonData, err := json.Marshal(send)
			assert.NoError(t, err)

			// Setup router
			r := router()
			request, err := http.NewRequest(
				"POST",
				"http://127.0.0.1:8080/api/pub/signup",
				bytes.NewBuffer(jsonData),
			)
			assert.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			// Get database values
			var entry models.User
			dbReq := db.C.Where("username = ?", send.Username).First(&entry)

			// Estimation of values
			if tt.args.validity {
				assert.Equal(t, http.StatusOK, response.Code)
				assert.NoError(t, dbReq.Error)
				assert.NotEqual(t, send.Password, entry.Password)
				err := bcrypt.CompareHashAndPassword(
					[]byte(entry.Password),
					[]byte(send.Password),
				)
				assert.NoError(t, err)
			} else {
				assert.NotEqual(t, http.StatusOK, response.Code)
				assert.Error(t, dbReq.Error)
			}
		})
	}
}

// Testing for inbound data in the handlers.LogIn() function.
func TestLogIn(t *testing.T) {
	type args struct {
		username string
		password string
		validity bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Existing data was accepted and password was hashed",
			args: args{
				username: "testuser",
				password: "abcdEFGH1234!@#$",
				validity: true,
			},
		},
		{
			name: "Missing username was not accepted",
			args: args{
				password: "abcdEFGH1234!@#$",
				validity: false,
			},
		},
		{
			name: "Missing password was not accepted",
			args: args{
				username: "testuser",
				validity: false,
			},
		},
		{
			name: "Not existing username was not accepted",
			args: args{
				username: "someuser",
				password: "abcdEFGH1234!@#$",
				validity: false,
			},
		},
		{
			name: "Incorrect password was not accepted",
			args: args{
				username: "testuser",
				password: "abcdEFGH1234!@#",
				validity: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			gin.SetMode(gin.TestMode)
			db.Connect()
			db.C.AutoMigrate(&models.User{}, &models.File{})
			defer db.C.Migrator().DropTable(
				&models.User{},
				&models.File{},
			)
			hashedPass, err := bcrypt.GenerateFromPassword(
				[]byte("abcdEFGH1234!@#$"), bcrypt.DefaultCost,
			)
			assert.NoError(t, err)
			user := models.User{
				Username: "testuser",
				Password: string(hashedPass),
				Files: []models.File{
					{
						ListName:  "file1",
						Name:      "/1/file1.txt",
						Extension: "txt",
						Path:      "/path/to/file1.txt",
						Date:      time.Now(),
						Size:      1024,
					},
					{
						ListName:  "file2",
						Name:      "/1/file2.jpg",
						Extension: "jpg",
						Path:      "/path/to/file2.jpg",
						Date:      time.Now(),
						Size:      2048,
					},
				},
			}
			db.C.Create(&user)

			// Create testing data
			send := models.User{
				Username: tt.args.username,
				Password: tt.args.password,
			}
			jsonData, err := json.Marshal(send)
			assert.NoError(t, err)

			// Setup router
			r := router()
			request, err := http.NewRequest(
				"POST",
				"http://127.0.0.1:8080/api/pub/login",
				bytes.NewBuffer(jsonData),
			)
			assert.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			// Get response values
			var result gin.H
			err = json.Unmarshal(response.Body.Bytes(), &result)
			assert.NoError(t, err)

			// Estimation of values
			if tt.args.validity {
				assert.Equal(t, http.StatusOK, response.Code)
				token, exists := result["token"]
				assert.True(t, exists)
				assert.NotEqual(t, token, "")
			} else {
				assert.NotEqual(t, http.StatusOK, response.Code)
				_, exists := result["token"]
				assert.False(t, exists)
			}
		})
	}
}

// Testing authentication for route group "/api/auth".
func TestAuth(t *testing.T) {
	type args struct {
		method string
		url    string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "handlers.List was require authentication",
			args: args{
				method: "GET",
				url:    "http://127.0.0.1:8080/api/auth/files",
			},
		},
		{
			name: "handlers.Download was require authentication",
			args: args{
				method: "GET",
				url:    "http://127.0.0.1:8080/api/auth/download",
			},
		},
		{
			name: "handlers.Upload was require authentication",
			args: args{
				method: "POST",
				url:    "http://127.0.0.1:8080/api/auth/upload",
			},
		},
		{
			name: "handlers.Rename was require authentication",
			args: args{
				method: "POST",
				url:    "http://127.0.0.1:8080/api/auth/rename",
			},
		},
		{
			name: "handlers.Delete was require authentication",
			args: args{
				method: "POST",
				url:    "http://127.0.0.1:8080/api/auth/delete",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			gin.SetMode(gin.TestMode)
			db.Connect()
			db.C.AutoMigrate(&models.User{}, &models.File{})
			defer db.C.Migrator().DropTable(
				&models.User{},
				&models.File{},
			)

			// Setup router
			r := router()
			request, err := http.NewRequest(tt.args.method, tt.args.url, nil)
			assert.NoError(t, err)
			request.Header.Set("Authorization", "Bearer niltoken")
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			// Estimation of values
			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})
	}
}

// Testing the returned data from the handlers.List() function.
func TestList(t *testing.T) {
	type args struct {
		user models.User
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "The entries list with 2 files was return",
			args: args{
				user: models.User{
					Username: "testuser",
					Password: "testpassword",
					Files: []models.File{
						{
							ListName:  "file1",
							Name:      "/1/file1.txt",
							Extension: "txt",
							Path:      "/path/to/file1.txt",
							Date:      time.Now(),
							Size:      1024,
						},
						{
							ListName:  "file2",
							Name:      "/1/file2.jpg",
							Extension: "jpg",
							Path:      "/path/to/file2.jpg",
							Date:      time.Now(),
							Size:      2048,
						},
					},
				},
			},
		},
		{
			name: "The empty entries list was return",
			args: args{
				user: models.User{
					Username: "testuser",
					Password: "testpassword",
					Files:    []models.File{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test database
			gin.SetMode(gin.TestMode)
			db.Connect()
			db.C.AutoMigrate(&models.User{}, &models.File{})
			defer db.C.Migrator().DropTable(
				&models.User{},
				&models.File{},
			)

			// Create testing data
			db.C.Create(&tt.args.user)

			// Setup router
			r := router()
			request, err := http.NewRequest(
				"GET",
				"http://127.0.0.1:8080/api/auth/files",
				nil,
			)
			assert.NoError(t, err)
			authJWT := middleware.JWT()
			token, _, _ := authJWT.TokenGenerator(&models.User{
				ID: tt.args.user.ID,
			})
			request.Header.Set("Authorization", "Bearer "+token)
			response := httptest.NewRecorder()
			r.ServeHTTP(response, request)

			// Get database values
			var entries []models.File
			db.C.Where("user_id = ?", tt.args.user.ID).Find(&entries)
			entriesJSON, err := json.Marshal(gin.H{"files": entries})
			assert.NoError(t, err)

			// Estimation of values
			assert.Equal(t, http.StatusOK, response.Code)
			assert.JSONEq(
				t,
				string(entriesJSON),
				strings.TrimSpace(response.Body.String()),
			)
		})
	}
}

// Testing for a correct file response in the handlers.Download()
// function.
func TestDownload(t *testing.T) {
	// Setup test database
	gin.SetMode(gin.TestMode)
	db.Connect()
	db.C.AutoMigrate(&models.User{}, &models.File{})
	defer db.C.Migrator().DropTable(
		&models.User{},
		&models.File{},
	)
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
		Files: []models.File{
			{
				ListName:  "test",
				Name:      "/1/test.file",
				Extension: "file",
				Path:      "upload/test.file",
				Date:      time.Now(),
				Size:      77344,
			},
		},
	}
	db.C.Create(&user)

	// Setup router
	r := router()
	request, err := http.NewRequest(
		"GET",
		"http://127.0.0.1:8080/api/auth/download?id=1&file=test.file",
		nil,
	)
	assert.NoError(t, err)
	authJWT := middleware.JWT()
	token, _, _ := authJWT.TokenGenerator(&models.User{
		ID: user.ID,
	})
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	// Get source bytes value
	source, err := os.ReadFile("upload/test.file")
	assert.NoError(t, err)

	// Estimation of values
	assert.Equal(t, source, response.Body.Bytes())
}

// Testing the correct file upload and creating a database entry in the
// handlers.Upload() function.
func TestUpload(t *testing.T) {
	// Setup test database
	gin.SetMode(gin.TestMode)
	db.Connect()
	db.C.AutoMigrate(&models.User{}, &models.File{})
	defer db.C.Migrator().DropTable(
		&models.User{},
		&models.File{},
	)
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
		Files:    []models.File{},
	}
	db.C.Create(&user)

	// Create testing data
	file, err := os.Open("upload/test.file")
	assert.NoError(t, err)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", "test.file")
	assert.NoError(t, err)
	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	writer.Close()

	// Setup router
	r := router()
	request, err := http.NewRequest(
		"POST",
		"http://127.0.0.1:8080/api/auth/upload",
		body,
	)
	assert.NoError(t, err)
	authJWT := middleware.JWT()
	token, _, _ := authJWT.TokenGenerator(&models.User{
		ID: user.ID,
	})
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	// Get bytes and database values
	loaded, err := os.ReadFile("upload/test/1/test.file")
	assert.NoError(t, err)
	source, err := os.ReadFile("upload/test.file")
	assert.NoError(t, err)
	err = os.RemoveAll("upload/test/1/")
	assert.NoError(t, err)
	var entry models.File
	dbReq := db.C.Where("user_id = ?", user.ID).First(&entry)

	// Estimation of values
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, source, loaded)
	assert.NoError(t, dbReq.Error)
	assert.Equal(t, "/1/test.file", entry.Name)
	assert.Equal(t, "test", entry.ListName)
	assert.Equal(t, ".file", entry.Extension)
	assert.Equal(t, int64(77344), entry.Size)
}

// Testing correct renaming of files in the handlers.Rename() function.
func TestRename(t *testing.T) {
	// Setup test database
	gin.SetMode(gin.TestMode)
	db.Connect()
	db.C.AutoMigrate(&models.User{}, &models.File{})
	defer db.C.Migrator().DropTable(
		&models.User{},
		&models.File{},
	)
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
		Files: []models.File{
			{
				ID:        1,
				ListName:  "file1",
				Name:      "/1/file1.txt",
				Extension: "txt",
				Path:      "/path/to/file1.txt",
				Date:      time.Now(),
				Size:      1024,
			},
		},
	}
	db.C.Create(&user)

	// Create testing data
	send := models.File{
		ID:   1,
		Name: "renamed",
	}
	jsonData, err := json.Marshal(send)
	assert.NoError(t, err)

	// Setup router
	r := router()
	request, err := http.NewRequest(
		"POST",
		"http://127.0.0.1:8080/api/auth/rename",
		bytes.NewBuffer(jsonData),
	)
	assert.NoError(t, err)
	authJWT := middleware.JWT()
	token, _, _ := authJWT.TokenGenerator(&models.User{
		ID: user.ID,
	})
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	// Get database values
	var entry models.File
	dbReq := db.C.Where("user_id = ?", user.ID).First(&entry)

	// Estimation of values
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NoError(t, dbReq.Error)
	assert.Equal(t, send.Name, entry.ListName)
	assert.Equal(t, "/1/renamed", entry.Name)
}

// Testing for correct file deletion in the handlers.Delete() function.
func TestDelete(t *testing.T) {
	// Setup test database
	gin.SetMode(gin.TestMode)
	db.Connect()
	db.C.AutoMigrate(&models.User{}, &models.File{})
	defer db.C.Migrator().DropTable(
		&models.User{},
		&models.File{},
	)
	user := models.User{
		Username: "testuser",
		Password: "testpassword",
		Files: []models.File{
			{
				ListName:  "test",
				Name:      "/1/test.file",
				Extension: "file",
				Path:      "upload/test/1/test.file",
				Date:      time.Now(),
				Size:      77344,
			},
		},
	}
	db.C.Create(&user)

	// Create testing data
	source, err := os.ReadFile("upload/test.file")
	assert.NoError(t, err)
	err = os.MkdirAll("upload/test/1/", os.ModePerm)
	assert.NoError(t, err)
	err = os.WriteFile("upload/test/1/test.file", source, 0666)
	assert.NoError(t, err)
	_, err = os.Stat("upload/test/1/test.file")
	assert.NoError(t, err)
	send := models.File{
		ID: 1,
	}
	jsonData, err := json.Marshal(send)
	assert.NoError(t, err)

	// Setup router
	r := router()
	request, err := http.NewRequest(
		"POST",
		"http://127.0.0.1:8080/api/auth/delete",
		bytes.NewBuffer(jsonData),
	)
	assert.NoError(t, err)
	authJWT := middleware.JWT()
	token, _, _ := authJWT.TokenGenerator(&models.User{
		ID: user.ID,
	})
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)

	// Get database values
	var entries []models.File
	dbReq := db.C.Where("user_id = ?", user.ID).Find(&entries)
	entriesJSON, err := json.Marshal(gin.H{"files": entries})
	assert.NoError(t, err)

	// Estimation of values
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NoError(t, dbReq.Error)
	assert.Equal(t, string(entriesJSON), "{\"files\":[]}")
	_, err = os.Stat("upload/test/1/test.file")
	assert.Error(t, err)
	err = os.RemoveAll("upload/test/1/")
	assert.NoError(t, err)
}
