package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/kmx0/project1/cmd/gophermart/storage"
	"github.com/kmx0/project1/internal/config"
	"github.com/kmx0/project1/internal/crypto"
	"github.com/kmx0/project1/internal/types"
)

func SetupRouter(cf config.Config) *gin.Engine {
	//  *storage.InMemory) {
	// store := storage.NewInMemory(cfg)
	// cfg = cf
	// SetRepository(store)

	r := gin.New()
	r.Use(gin.Recovery(),
		// Compress(),
		// Decompress(),
		HandleAutorize(),
		gin.Logger())

	r.POST("/api/user/orders", HandleOrders)
	r.POST("/api/user/register", HandleRegister)
	r.POST("/api/user/login", HandleLogin)

	// r.POST("/update/", HandleUpdateJSON)
	// r.POST("/updates/", HandleUpdateBatchJSON)
	// r.POST("/value/", HandleValueJSON)

	// r.GET("/", HandleAllValues)
	// r.GET("/ping", HandlePing)
	// r.GET("/value/:typem/:metric", HandleValue)
	return r
}

func HandleWithoutID(c *gin.Context) {
	c.Status(http.StatusNotFound)
}

func HandleRegister(c *gin.Context) {
	logrus.SetReportCaller(true)

	body := c.Request.Body
	defer body.Close()

	decoder := json.NewDecoder(body)
	var user types.User

	err := decoder.Decode(&user)
	if err != nil {
		c.Status(http.StatusBadRequest)
	}
	logrus.Info(user)
	id, err := storage.RegisterUser(user)
	logrus.Info(err)
	if err == nil {
		user.ID = id
		user.Cookie = crypto.CookieHash(c.Request.RemoteAddr, c.Request.UserAgent(), user.Login)
		err := storage.WriteUserCookie(user)
		if err != nil {
			logrus.Error(err)
			// c.Status(http.StatusInternalServerError)
		}
		c.Header("session", user.Cookie)
		c.Status(http.StatusOK)
	} else {
		erStr := err.Error()
		switch {
		case strings.Contains(erStr, "duplicate"):
			c.Status(http.StatusConflict)
		default:
			c.Status(http.StatusInternalServerError)
		}
	}

}
func HandleLogin(c *gin.Context) {
	logrus.SetReportCaller(true)

	body := c.Request.Body
	defer body.Close()
	// crypto.CookieHash(c.ClientIP(), c.Request.UserAgent(), )

	decoder := json.NewDecoder(body)
	var user types.User

	err := decoder.Decode(&user)
	if err != nil {
		c.Status(http.StatusBadRequest)
	}
	logrus.Info(user)
	user.IP = c.Request.RemoteAddr
	user.UserAgent = c.Request.UserAgent()
	user.ID, user.Cookie, err = storage.LoginUser(user)
	logrus.Info(err)
	if err == nil {

		storage.WriteUserCookie(user)
		c.Header("session", user.Cookie)

		c.Status(http.StatusOK)
	}
	if err != nil {
		erStr := err.Error()
		switch {
		case strings.Contains(erStr, "incorrect"):
			c.Status(http.StatusUnauthorized)
		default:
			c.Status(http.StatusInternalServerError)
		}
	}

}

func HandleOrders(c *gin.Context) {
	logrus.SetReportCaller(true)
	contenType := c.GetHeader("Content-Type")
	if contenType != "text/html; charset=utf-8" {
		c.Status(http.StatusBadRequest)
	}
	// c.Header("Content-Type", "text/html; charset=utf-8")
	// Content-Type: text/plain
	body := c.Request.Body
	defer body.Close()
	// crypto.CookieHash(c.ClientIP(), c.Request.UserAgent(), )
	// cookie := c.GetHeader("session")
	order, _ := ioutil.ReadAll(body)
	orderInt, _ := strconv.Atoi(string(order))
	// logrus.Info("Need check by LUN")
	if crypto.CalculateLuhn(orderInt) {
		// storage.LoadNewOrder(cookie, order)

		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusUnprocessableEntity)
	}

}

func HandleAutorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI != "/api/user/register" && c.Request.RequestURI != "/api/user/login" {
			cookie := c.GetHeader("session")
			// logrus.Info(cookie)
			if cookie != "" {
				err := storage.CheckCookie(cookie, c.Request.RemoteAddr, c.Request.UserAgent())
				if err != nil {
					logrus.Error(err)
					// c.Status(http.StatusUnauthorized)
					c.AbortWithStatus(http.StatusUnauthorized)
					// c.Redirect(http.StatusUnauthorized, "/api/user/login")
					return
					// logrus.Error()
				} else {
					c.Next()
				}
			} else {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		} else {
			c.Next()
		}
	}
}
