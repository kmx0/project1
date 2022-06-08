package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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

	r.POST("/api/user/orders", HandlePostOrders)
	r.POST("/api/user/register", HandleRegister)
	r.POST("/api/user/login", HandleLogin)
	r.GET("/api/user/orders", HandleGetOrders)

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
	} else {

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
			c.SetCookie("session", user.Cookie, 60*60, "", "", false, true)
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
		logrus.Error(err)
		c.Status(http.StatusBadRequest)
		return
	}
	logrus.Info(user)
	user.IP = c.Request.RemoteAddr
	user.UserAgent = c.Request.UserAgent()
	user.ID, user.Cookie, err = storage.LoginUser(user)
	logrus.Info(err)
	if err == nil {

		storage.WriteUserCookie(user)
		c.SetCookie("session", user.Cookie, 60*60, "", "", false, true)
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

func HandlePostOrders(c *gin.Context) {
	logrus.SetReportCaller(true)
	contenType := c.GetHeader("Content-Type")
	logrus.Info(contenType)
	if contenType != "text/plain" {
		c.Status(http.StatusBadRequest)
	} else {
		// cookieHeader := c.GetHeader("Set-Cookie")
		cookie, err := c.Request.Cookie("session")
		logrus.Info(cookie, err)
		// cookie := cookieHeader.

		// c.Header("Content-Type", "text/html; charset=utf-8")
		// Content-Type: text/plain
		body := c.Request.Body
		defer body.Close()
		// crypto.	CookieHash(c.ClientIP(), c.Request.UserAgent(), )
		order, _ := ioutil.ReadAll(body)
		orderInt := string(order)
		// logrus.Info("Need check by LUN")
		if crypto.CalculateLuhn(orderInt) {
			err := storage.LoadNewOrder(cookie.Value, orderInt)
			logrus.Error(err)
			switch {
			case err == nil:
				c.Status(http.StatusAccepted)
			case strings.Contains(err.Error(), `duplicate key value violates unique constraint "orders_number_key"`):
				c.Status(http.StatusOK)
			case strings.Contains(err.Error(), `order belongs other user`):
				c.Status(http.StatusConflict)
			default:
				c.Status(http.StatusInternalServerError)
			}
		} else {
			c.Status(http.StatusUnprocessableEntity)
		}
	}

}
func HandleGetOrders(c *gin.Context) {
	logrus.SetReportCaller(true)
	// cookieHeader := c.GetHeader("Set-Cookie")
	cookie, err := c.Request.Cookie("session")
	c.Header("Content-Type", "application/json")
	logrus.Info(cookie, err)
	ordersList, err := storage.GetOrdersList(cookie.Value)
	if err == nil {
		if len(ordersList) == 0 {
			logrus.Info(ordersList)
			c.JSON(http.StatusNoContent, ordersList)
			return
		}
		logrus.Info(ordersList)
		body, err := json.MarshalIndent(ordersList, "\t", "\t")
		if err != nil {
			logrus.Error(err)
		}
		c.JSON(http.StatusOK, ordersList)
		logrus.Info(string(body))
	} else {
		switch {
		case strings.Contains(err.Error(), "1"):
			c.Status(http.StatusInternalServerError)
		}
	}
}

func HandleAutorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI != "/api/user/register" && c.Request.RequestURI != "/api/user/login" {
			cookie, err := c.Request.Cookie("session")
			// logrus.Info(cookie)
			if err == nil {
				err := storage.CheckCookie(cookie.Value, c.Request.RemoteAddr, c.Request.UserAgent())
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
