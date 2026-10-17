package v1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IceWhaleTech/CasaOS/model"
	"github.com/IceWhaleTech/CasaOS/pkg/platform"
	"github.com/IceWhaleTech/CasaOS/pkg/utils/common_err"
	"github.com/IceWhaleTech/CasaOS/pkg/utils/encryption"
	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	"github.com/labstack/echo/v4"
)

type UserRecord struct {
	ID        int                    `json:"id"`
	Username  string                 `json:"username"`
	Password  string                 `json:"password"` // MD5 hash
	Nickname  string                 `json:"nickname"`
	Role      string                 `json:"role"`
	Avatar    string                 `json:"avatar"`
	Email     string                 `json:"email"`
	Custom    map[string]interface{} `json:"custom"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

type UserStore struct {
	Users       map[string]*UserRecord `json:"users"`
	Initialized bool                   `json:"initialized"`
}

var (
	userStoreLock sync.RWMutex
	userStore     *UserStore

	jwtPrivateKey *ecdsa.PrivateKey
	jwtPublicKey  *ecdsa.PublicKey
	jwksJSON      []byte
	jwtLock       sync.RWMutex
)

func init() {
	userStore = &UserStore{
		Users:       make(map[string]*UserRecord),
		Initialized: false,
	}
	_ = initJWTKeys()
}

func getUserConfigPath() string {
	dir := ""
	if platform.Provider != nil {
		dir = platform.Provider.ConfigDir()
	}
	if dir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		dir = filepath.Join(home, ".casaos", "conf")
	}
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "users.json")
}

func getJWTKeyPath() string {
	dir := ""
	if platform.Provider != nil {
		dir = platform.Provider.ConfigDir()
	}
	if dir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			home = "/data/data/com.termux/files/home"
		}
		dir = filepath.Join(home, ".casaos", "conf")
	}
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "jwt.key")
}

func initJWTKeys() error {
	jwtLock.Lock()
	defer jwtLock.Unlock()

	keyPath := getJWTKeyPath()
	if data, err := os.ReadFile(keyPath); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
				jwtPrivateKey = key
				jwtPublicKey = &key.PublicKey
				jwksJSON, _ = jwt.GenerateJwksJSON(jwtPublicKey)
				return nil
			}
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	jwtPrivateKey = key
	jwtPublicKey = &key.PublicKey
	jwksJSON, _ = jwt.GenerateJwksJSON(jwtPublicKey)

	if der, err := x509.MarshalECPrivateKey(key); err == nil {
		pemBytes := pem.EncodeToMemory(&pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: der,
		})
		_ = os.WriteFile(keyPath, pemBytes, 0600)
	}

	return nil
}

func GetJWKSPublicKey() *ecdsa.PublicKey {
	jwtLock.RLock()
	defer jwtLock.RUnlock()
	return jwtPublicKey
}

func GetJWKSBytes() []byte {
	jwtLock.RLock()
	defer jwtLock.RUnlock()
	return jwksJSON
}

func loadUserStore() {
	path := getUserConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var store UserStore
	if err := json.Unmarshal(data, &store); err == nil {
		if store.Users == nil {
			store.Users = make(map[string]*UserRecord)
		}
		userStore = &store
	}
}

func saveUserStore() error {
	path := getUserConfigPath()
	data, err := json.MarshalIndent(userStore, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GET /v1/users/status
func GetUserStatus(ctx echo.Context) error {
	userStoreLock.Lock()
	loadUserStore()
	initialized := userStore.Initialized && len(userStore.Users) > 0
	userStoreLock.Unlock()

	data := map[string]interface{}{
		"initialized": initialized,
		"key":         "casaos_init_key",
		"gpus":        0,
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data:    data,
	})
}

// POST /v1/users/register
func PostUserRegister(ctx echo.Context) error {
	userStoreLock.Lock()
	defer userStoreLock.Unlock()
	loadUserStore()

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Nickname string `json:"nickname"`
		Role     string `json:"role"`
		Key      string `json:"key"`
	}

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, model.Result{
			Success: common_err.CLIENT_ERROR,
			Message: "Invalid registration payload: " + err.Error(),
		})
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Password) < 1 {
		return ctx.JSON(http.StatusBadRequest, model.Result{
			Success: common_err.CLIENT_ERROR,
			Message: "Username and password cannot be empty",
		})
	}

	if req.Nickname == "" {
		req.Nickname = req.Username
	}
	if req.Role == "" {
		req.Role = "admin"
	}

	if userStore.Initialized && len(userStore.Users) > 0 {
		// Only allow registration if uninitialized or explicitly permitted
		if _, exists := userStore.Users[req.Username]; exists {
			return ctx.JSON(http.StatusBadRequest, model.Result{
				Success: common_err.CLIENT_ERROR,
				Message: "User already exists",
			})
		}
	}

	user := &UserRecord{
		ID:        len(userStore.Users) + 1,
		Username:  req.Username,
		Password:  encryption.GetMD5ByStr(req.Password),
		Nickname:  req.Nickname,
		Role:      req.Role,
		Avatar:    "",
		Custom:    make(map[string]interface{}),
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	userStore.Users[req.Username] = user
	userStore.Initialized = true
	_ = saveUserStore()

	jwtLock.RLock()
	privKey := jwtPrivateKey
	jwtLock.RUnlock()

	accessToken, _ := jwt.GetAccessToken(user.Username, privKey, user.ID)
	refreshToken, _ := jwt.GetRefreshToken(user.Username, privKey, user.ID)

	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   86400 * 7,
	}
	ctx.SetCookie(cookie)

	data := map[string]interface{}{
		"token":         accessToken,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    86400,
		"id":            user.ID,
		"username":      user.Username,
		"nickname":      user.Nickname,
		"role":          user.Role,
		"avatar":        user.Avatar,
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data:    data,
	})
}

// POST /v1/users/login
func PostUserLogin(ctx echo.Context) error {
	userStoreLock.Lock()
	defer userStoreLock.Unlock()
	loadUserStore()

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, model.Result{
			Success: common_err.CLIENT_ERROR,
			Message: "Invalid login payload",
		})
	}

	req.Username = strings.TrimSpace(req.Username)
	user, exists := userStore.Users[req.Username]
	if !exists {
		return ctx.JSON(http.StatusOK, model.Result{
			Success: common_err.USER_NOT_EXIST,
			Message: "User does not exist",
		})
	}

	hashPwd := encryption.GetMD5ByStr(req.Password)
	if user.Password != hashPwd && user.Password != req.Password {
		return ctx.JSON(http.StatusOK, model.Result{
			Success: common_err.PWD_INVALID,
			Message: "Incorrect password",
		})
	}

	jwtLock.RLock()
	privKey := jwtPrivateKey
	jwtLock.RUnlock()

	accessToken, _ := jwt.GetAccessToken(user.Username, privKey, user.ID)
	refreshToken, _ := jwt.GetRefreshToken(user.Username, privKey, user.ID)

	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   86400 * 7,
	}
	ctx.SetCookie(cookie)

	data := map[string]interface{}{
		"token":         accessToken,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    86400,
		"id":            user.ID,
		"username":      user.Username,
		"nickname":      user.Nickname,
		"role":          user.Role,
		"avatar":        user.Avatar,
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data:    data,
	})
}

// POST /v1/users/refresh
func PostUserRefresh(ctx echo.Context) error {
	userStoreLock.RLock()
	defer userStoreLock.RUnlock()

	username := "admin"
	userID := 1
	for _, u := range userStore.Users {
		username = u.Username
		userID = u.ID
		break
	}

	jwtLock.RLock()
	privKey := jwtPrivateKey
	jwtLock.RUnlock()

	accessToken, _ := jwt.GetAccessToken(username, privKey, userID)
	refreshToken, _ := jwt.GetRefreshToken(username, privKey, userID)

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data: map[string]interface{}{
			"token":         accessToken,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

// GET /v1/users/current
func GetCurrentUser(ctx echo.Context) error {
	userStoreLock.RLock()
	defer userStoreLock.RUnlock()
	loadUserStore()

	userIDStr := ctx.Request().Header.Get("user_id")
	var targetUser *UserRecord

	if userIDStr != "" {
		if uid, err := strconv.Atoi(userIDStr); err == nil {
			for _, u := range userStore.Users {
				if u.ID == uid {
					targetUser = u
					break
				}
			}
		}
	}

	if targetUser == nil {
		for _, u := range userStore.Users {
			targetUser = u
			break
		}
	}

	if targetUser == nil {
		return ctx.JSON(http.StatusOK, model.Result{
			Success: common_err.SUCCESS,
			Message: "ok",
			Data: map[string]interface{}{
				"id":       1,
				"username": "admin",
				"nickname": "admin",
				"role":     "admin",
				"avatar":   "",
				"custom":   map[string]interface{}{},
			},
		})
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data: map[string]interface{}{
			"id":       targetUser.ID,
			"username": targetUser.Username,
			"nickname": targetUser.Nickname,
			"role":     targetUser.Role,
			"avatar":   targetUser.Avatar,
			"custom":   targetUser.Custom,
		},
	})
}

// GET /v1/users/current/custom/:key
func GetUserCustom(ctx echo.Context) error {
	key := ctx.Param("key")
	userStoreLock.RLock()
	defer userStoreLock.RUnlock()

	var customVal interface{} = ""
	for _, u := range userStore.Users {
		if u.Custom != nil {
			if val, ok := u.Custom[key]; ok {
				customVal = val
			}
		}
		break
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data:    customVal,
	})
}

// POST /v1/users/current/custom/:key
func PostUserCustom(ctx echo.Context) error {
	key := ctx.Param("key")
	var body map[string]interface{}
	_ = ctx.Bind(&body)

	userStoreLock.Lock()
	defer userStoreLock.Unlock()

	for _, u := range userStore.Users {
		if u.Custom == nil {
			u.Custom = make(map[string]interface{})
		}
		if val, ok := body["value"]; ok {
			u.Custom[key] = val
		} else {
			u.Custom[key] = body
		}
		_ = saveUserStore()
		break
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
	})
}

// GET /v1/users/name
func GetUserNameList(ctx echo.Context) error {
	userStoreLock.RLock()
	defer userStoreLock.RUnlock()

	list := make([]map[string]interface{}, 0)
	for _, u := range userStore.Users {
		list = append(list, map[string]interface{}{
			"id":       u.ID,
			"username": u.Username,
			"nickname": u.Nickname,
			"role":     u.Role,
			"avatar":   u.Avatar,
		})
	}

	return ctx.JSON(http.StatusOK, model.Result{
		Success: common_err.SUCCESS,
		Message: common_err.GetMsg(common_err.SUCCESS),
		Data:    list,
	})
}

// GET /v1/users/avatar
func GetUserAvatar(ctx echo.Context) error {
	return ctx.Redirect(http.StatusTemporaryRedirect, "/default-avatar.svg")
}

// GET /.well-known/jwks.json
func GetJWKS(ctx echo.Context) error {
	jwtLock.RLock()
	bytes := jwksJSON
	jwtLock.RUnlock()
	return ctx.Blob(http.StatusOK, "application/json", bytes)
}
