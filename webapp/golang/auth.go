package main

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/catatsuy/private-isu/webapp/golang/grpc"
	"github.com/catatsuy/private-isu/webapp/golang/templates"
	"github.com/catatsuy/private-isu/webapp/golang/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/quicktemplate"
	"google.golang.org/protobuf/proto"
)

type Session struct {
	Session *grpc.Session
	User    *types.User
}

func tryLogin(accountName, password string) *types.User {
	u := types.User{}
	err := db.Get(&u, "SELECT * FROM users WHERE account_name = ? AND del_flg = 0", accountName)
	if err != nil {
		return nil
	}

	if calculatePasshash(u.AccountName, password) == u.Passhash {
		return &u
	} else {
		return nil
	}
}

func validateUser(accountName, password string) bool {
	return regexp.MustCompile(`\A[0-9a-zA-Z_]{3,}\z`).MatchString(accountName) &&
		regexp.MustCompile(`\A[0-9a-zA-Z_]{6,}\z`).MatchString(password)
}
func digest(src string) string {
	// opensslのバージョンによっては (stdin)= というのがつくので取る
	return fmt.Sprintf("%x", sha512.Sum512([]byte(src)))
}

func calculateSalt(accountName string) string {
	return digest(accountName)
}

func calculatePasshash(accountName, password string) string {
	return digest(password + ":" + calculateSalt(accountName))
}

// ランダム文字列を生成する関数
func GenerateSecureRandomString(length int) string {
	// 指定された長さのバイト配列を作成
	randomBytes := make([]byte, length)

	// 暗号学的に安全なランダムバイトを生成
	rand.Read(randomBytes)

	// バイト配列をBase64エンコードして文字列に変換
	return base64.RawURLEncoding.EncodeToString(randomBytes)
}

func getSession(r *http.Request) *Session {
	var session grpc.Session
	var user grpc.User
	var myuser *types.User
	var err error
	var items map[string]*memcache.Item

	cookie := r.Header.Get("Cookie")
	var sessionParts []string
	for _, c := range strings.Split(cookie, ";") {
		if strings.HasPrefix(c, "isuconp-go.session=") {
			sessionParts = strings.Split(c, "=")
			break
		}
	}
	if len(sessionParts) != 3 {
		goto new
	}
	// isuconp-go.session=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=user:123
	items, _ = memcacheClient.GetMulti([]string{sessionParts[1], sessionParts[2]})

	if items[sessionParts[1]] == nil {
		goto new
	}
	if err = proto.Unmarshal(items[sessionParts[1]].Value, &session); err != nil {
		log.Printf("failed to unmarshal session: %v", err)
		goto new
	}

	if "user:"+strconv.Itoa(int(session.UserID)) != sessionParts[2] {
		goto new
	}

	if items[sessionParts[2]] != nil {
		if err := proto.Unmarshal(items[sessionParts[2]].Value, &user); err != nil {
			log.Printf("failed to unmarshal user: %v", err)
			goto new
		}
		myuser = &types.User{
			ID:          int(user.ID),
			AccountName: user.AccountName,
			Authority:   int(user.Authority),
			DelFlg:      int(user.DelFlg),
			CreatedAt:   user.CreatedAt.AsTime(),
		}
	}

	return &Session{
		Session: &session,
		User:    myuser,
	}

new:
	return &Session{
		Session: &grpc.Session{
			SessionID: GenerateSecureRandomString(16),
		},
	}
}

func saveSession(w http.ResponseWriter, session *Session) {
	maxage := 86400
	if session.Session.UserID == -1 {
		maxage = -1
	}
	setcookie := fmt.Sprintf("isuconp-go.session=%s=user:%d; Max-Age=%d", session.Session.SessionID, session.Session.UserID, maxage)
	w.Header().Set("Set-Cookie", setcookie)

	sessionBytes, _ := proto.Marshal(session.Session)
	memcacheClient.Set(&memcache.Item{
		Key:   session.Session.SessionID,
		Value: sessionBytes,
	})
	if session.User != nil {
		cacheUsers([]types.User{*session.User})
	}
}

func getSessionUser(r *http.Request) types.User {
	session := getSession(r)
	if session.User != nil {
		return *session.User
	}

	uid := session.Session.UserID
	if uid == 0 {
		return types.User{}
	}

	u := types.User{}

	err := db.Get(&u, "SELECT * FROM `users` WHERE `id` = ?", uid)
	if err != nil {
		return types.User{}
	}

	session.User = &u

	return u
}

func isLogin(u types.User) bool {
	return u.ID != 0
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	me := getSessionUser(r)

	if isLogin(me) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	templates.WriteLayout(w, func(qw *quicktemplate.Writer) {
		templates.StreamLoginPage(qw, getFlash(w, r))
	}, me)
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	if isLogin(getSessionUser(r)) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	u := tryLogin(r.FormValue("account_name"), r.FormValue("password"))

	if u != nil {
		session := getSession(r)
		session.Session.UserID = int32(u.ID)
		session.Session.CsrfToken = secureRandomStr(16)
		saveSession(w, session)

		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		session := getSession(r)
		session.Session.Notice = "アカウント名かパスワードが間違っています"
		saveSession(w, session)

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func getRegister(w http.ResponseWriter, r *http.Request) {
	if isLogin(getSessionUser(r)) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	templates.WriteLayout(w, func(qw *quicktemplate.Writer) {
		templates.StreamRegisterPage(qw, getFlash(w, r))
	}, types.User{})
}

func postRegister(w http.ResponseWriter, r *http.Request) {
	if isLogin(getSessionUser(r)) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	accountName, password := r.FormValue("account_name"), r.FormValue("password")

	validated := validateUser(accountName, password)
	if !validated {
		session := getSession(r)
		session.Session.Notice = "アカウント名は3文字以上、パスワードは6文字以上である必要があります"
		saveSession(w, session)

		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	exists := 0
	// ユーザーが存在しない場合はエラーになるのでエラーチェックはしない
	db.Get(&exists, "SELECT 1 FROM users WHERE `account_name` = ?", accountName)

	if exists == 1 {
		session := getSession(r)
		session.Session.Notice = "アカウント名がすでに使われています"
		saveSession(w, session)

		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	query := "INSERT INTO `users` (`account_name`, `passhash`) VALUES (?,?)"
	result, err := db.Exec(query, accountName, calculatePasshash(accountName, password))
	if err != nil {
		log.Print(err)
		return
	}

	session := getSession(r)
	uid, err := result.LastInsertId()
	if err != nil {
		log.Print(err)
		return
	}
	session.Session.UserID = int32(uid)
	session.Session.CsrfToken = secureRandomStr(16)
	saveSession(w, session)

	http.Redirect(w, r, "/", http.StatusFound)
}

func getLogout(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	session.Session.UserID = -1
	saveSession(w, session)

	http.Redirect(w, r, "/", http.StatusFound)
}
