package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/kmx0/project1/cmd/accrual"
	"github.com/kmx0/project1/cmd/gophermart/storage"
	"github.com/kmx0/project1/internal/config"
	"github.com/kmx0/project1/internal/crypto"
	"github.com/kmx0/project1/internal/types"
)

var store storage.Storage

var cfg config.Config

func SetRepository(s storage.Storage) {
	store = s
}

func SetupRouter(cf config.Config, store storage.Storage) *gin.Engine {
	//  *storage.InMemory) {
	// store := storage.NewDB()
	cfg = cf
	SetRepository(store)

	r := gin.New()
	r.Use(gin.Recovery(),
		// Compress(),
		// Decompress(),
		HandleAutorize(),
		gin.Logger())

	r.POST("/api/user/register", HandleRegister)
	r.POST("/api/user/login", HandleLogin)
	r.GET("/api/user/orders", HandleGetOrders)
	r.POST("/api/user/orders", HandlePostOrder)
	r.GET("/api/user/balance", HandleGetBalance)
	r.POST("/api/user/balance/withdraw", HandlePostWithdraw)
	r.GET("/api/user/balance/withdrawals", HandleGetWithdrawals)

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
	//add check json format in content-type
	decoder := json.NewDecoder(body)
	var user types.User

	err := decoder.Decode(&user)
	if err != nil {
		c.Status(http.StatusBadRequest)
	} else {

		logrus.Info(user)
		id, err := store.RegisterUser(user)
		logrus.Info(err)
		if err == nil {
			user.ID = id
			user.Cookie = crypto.CookieHash(c.Request.RemoteAddr, c.Request.UserAgent(), user.Login)
			err := store.WriteUserCookie(user, id)
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
	user.ID, user.Cookie, err = store.LoginUser(user)
	logrus.Info(err)
	if err == nil {

		store.WriteUserCookie(user, user.ID)
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

func HandlePostOrder(c *gin.Context) {
	logrus.SetReportCaller(true)
	contenType := c.GetHeader("Content-Type")
	logrus.Info(contenType)
	if contenType != "text/plain" {
		c.Status(http.StatusBadRequest)
	} else {
		// cookieHeader := c.GetHeader("Set-Cookie")
		cookie, _ := c.Request.Cookie("session")
		// logrus.Info(cookie, err)
		// cookie := cookieHeader.

		// c.Header("Content-Type", "text/html; charset=utf-8")
		// Content-Type: text/plain
		body := c.Request.Body
		defer body.Close()
		// crypto.	CookieHash(c.ClientIP(), c.Request.UserAgent(), )
		order, _ := ioutil.ReadAll(body)
		orderInt := string(order)
		// logrus.Info("Need check by LUN")
		// accrual.GetAccrual(cfg.AccSysSddr, orderInt)
		if crypto.CalculateLuhn(orderInt) {
			err := store.LoadNewOrder(cookie.Value, orderInt)
			logrus.Error(err)
			switch {
			case err == nil:
				accrual.GetAccrual(store, cfg.AccSysSddr, orderInt)
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
	ordersList, err := store.GetOrdersList(cookie.Value)
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

func HandleGetBalance(c *gin.Context) {
	logrus.SetReportCaller(true)
	// cookieHeader := c.GetHeader("Set-Cookie")
	cookie, err := c.Request.Cookie("session")
	logrus.Info(cookie, err)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Type", "application/json")
	// cfg.AccSysSddr
	current, err := store.GetBalance(cookie.Value)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	userID, err := store.GetUserID(cookie.Value)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	withdrawals, err := store.GetSUMWithdraws(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var balance types.Balance
	balance.Current = current
	balance.Withdrawn = withdrawals

	c.JSON(http.StatusOK, balance)

}
func HandlePostWithdraw(c *gin.Context) {
	logrus.SetReportCaller(true)
	contenType := c.GetHeader("Content-Type")
	logrus.Info(contenType)
	if contenType != "application/json" {
		c.Status(http.StatusUnprocessableEntity)
		return
	}

	body := c.Request.Body
	defer body.Close()

	decoder := json.NewDecoder(body)
	var withdraw types.Withdraw

	err := decoder.Decode(&withdraw)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusUnprocessableEntity)
		return
	}
	//check order
	if !crypto.CalculateLuhn(withdraw.Order) {
		c.Status(http.StatusUnprocessableEntity)
		return
	}
	//block table
	// get user_id for cookie
	cookie, err := c.Request.Cookie("session")
	logrus.Info(cookie, err)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}

	userID, err := store.GetUserID(cookie.Value)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	withdrawals, err := store.GetWithdrawals(userID)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	//check new order or not
	for _, v := range withdrawals {
		if v.Order == withdraw.Order {
			c.Status(http.StatusUnprocessableEntity)
			return
		}
	}
	current, err := store.GetBalance(cookie.Value)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if current < withdraw.Sum {
		c.Status(http.StatusPaymentRequired)
		return
	}

	err = store.ChangeBalanceValue(withdraw.Sum, "-", userID)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	err = store.WriteWithdraw(withdraw, userID)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)

}

func HandleGetWithdrawals(c *gin.Context) {
	logrus.SetReportCaller(true)
	// cookieHeader := c.GetHeader("Set-Cookie")
	cookie, err := c.Request.Cookie("session")
	c.Header("Content-Type", "application/json")
	logrus.Info(cookie, err)
	userID, err := store.GetUserID(cookie.Value)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	withdrawals, err := store.GetWithdrawals(userID)
	if err != nil {
		logrus.Error(err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if withdrawals == nil {
		c.Status(http.StatusNoContent)
	}
	c.JSON(http.StatusOK, withdrawals)

}

func HandleAutorize() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI != "/api/user/register" && c.Request.RequestURI != "/api/user/login" {
			cookie, err := c.Request.Cookie("session")
			// logrus.Info(cookie)
			if err == nil {
				err := store.CheckCookie(cookie.Value, c.Request.RemoteAddr, c.Request.UserAgent())
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
