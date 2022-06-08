package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/kmx0/project1/internal/crypto"
	"github.com/kmx0/project1/internal/types"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var DB *sql.DB
var Conn *pgx.Conn
var DBName = "gophermart"

func PingDB(ctx context.Context, urlExample string) bool {
	// urlExample = "postgres://postgres:postgres@localhost:5432/gophermart"
	logrus.Info(urlExample)
	var err error
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	Conn, err = pgx.Connect(context.Background(), urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	// defer conn.Close(context.Background())

	err = Conn.Ping(context.Background())
	if err != nil {
		logrus.Error(err)
		return false
	}

	logrus.Info("Successfully connected!")
	logrus.Info(CheckDBExist())
	// if !CheckDBExist() {

	// }
	TableName := "users"
	if !CheckTableExist(TableName) {
		req := fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			login varchar(255) UNIQUE,
			password varchar(255)
		);`, TableName)
		AddTabletoDB(req)
	}
	TableName = "sessions"

	if !CheckTableExist(TableName) {
		req := fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			user_id  integer,
			cookie varchar(255),
			ttl timestamp with time zone,
			CONSTRAINT fk_user
			FOREIGN KEY(user_id) 
			REFERENCES users(id)
			ON DELETE CASCADE
		);`, TableName)
		AddTabletoDB(req)
	}
	TableName = "orders"

	if !CheckTableExist(TableName) {
		req := fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			user_id  integer,
			number bigint UNIQUE,
			status varchar(255),
			accrual integer,
			uploaded_at timestamp with time zone,
			CONSTRAINT fk_user
			FOREIGN KEY(user_id) 
			REFERENCES users(id)
			ON DELETE CASCADE
		);`, TableName)
		AddTabletoDB(req)
	}
	// logrus.Info(CheckTableExist())
	return true
}

func CheckDBExist() bool {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return false
	}
	listDB := `SELECT datname FROM pg_database;`
	rows, err := Conn.Query(context.Background(), listDB)
	if err != nil {
		logrus.Error(err)
	}
	// c, _ := result
	defer rows.Close()
	for rows.Next() {
		var res string
		rows.Scan(&res)
		// logrus.Info(res)
		if res == DBName {
			return true
		}
	}
	err = rows.Err()
	if err != nil {
		return false
	}
	return false
}

func CheckTableExist(TableName string) bool {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return false
	}
	listTables := `SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_schema='public';`
	rows, err := Conn.Query(context.Background(), listTables)
	if err != nil {
		logrus.Error(err)
	}
	// c, _ := result
	defer rows.Close()
	for rows.Next() {
		var res string
		rows.Scan(&res)
		logrus.Info(res)
		if res == TableName {
			return true
		}
	}
	err = rows.Err()
	if err != nil {
		return false
	}
	return false
}

func AddTabletoDB(req string) {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return
	}
	rows, err := Conn.Query(context.Background(), req)
	if err != nil {
		logrus.Error(err)
	}
	// c, _ := result
	defer rows.Close()
	for rows.Next() {
		var res string
		rows.Scan(&res)
		logrus.Info(res)

	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
	}

}
func RegisterUser(user types.User) (id int, err error) {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return 0, errors.New("error nil Conn")
	}
	// checkExist := fmt.Sprintf("SELECT users (SELECT 1 FROM users WHERE Login = %s LIMIT 1);", user.Login)
	insert := `INSERT INTO users(Login, Password) values($1, $2) RETURNING id`

	err = Conn.QueryRow(context.Background(), insert, user.Login, user.Password).Scan(&id)

	return id, err
}
func LoginUser(user types.User) (id int, cookie string, err error) {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return 0, "", errors.New("error nil Conn")
	}
	loginReq := fmt.Sprintf("SELECT id, login, password FROM users WHERE login = '%s' ;", user.Login)

	ctx := context.Background()
	rowsC, err := Conn.Query(ctx, loginReq)
	if err != nil {
		logrus.Error(err)
		return 0, "", err
	}
	defer rowsC.Close()
	var login string
	var password string
	for rowsC.Next() {
		rowsC.Scan(&id, &login, &password)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return 0, "", err
	}
	if login == user.Login && password == user.Password {
		logrus.Infoln(user.IP, user.UserAgent, user.Login)
		return id, crypto.CookieHash(user.IP, user.UserAgent, user.Login), nil
	} else {
		return 0, "", errors.New("incorrect login or password")
	}
}

func WriteUserCookie(user types.User) error {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	CookieForDelReq := fmt.Sprintf("SELECT cookie FROM sessions WHERE user_id = '%d' ;", user.ID)
	ctx := context.Background()
	rowsC, err := Conn.Query(ctx, CookieForDelReq)
	if err != nil {
		logrus.Error(err)
		return err
	}
	var cookie string
	defer rowsC.Close()
	for rowsC.Next() {
		rowsC.Scan(&cookie)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if cookie != user.Cookie {
		insert := `INSERT INTO sessions(cookie, ttl, user_id) values($1, $2, $3);`
		_, err = Conn.Exec(context.Background(), insert, user.Cookie, time.Now().Add(time.Hour*1), user.ID)
	} else {
		update := `UPDATE sessions SET ttl = $1 WHERE user_id = $2;`
		_, err = Conn.Exec(context.Background(), update, time.Now().Add(time.Hour*1), user.ID)
		if err != nil {
			logrus.Error(err)
		}
	}
	return err
}

func DeleteCookie(cookie string) error {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	deleteCookie := `DELETE FROM sessions WHERE cookie = $1;`

	// logrus.Info(deleteCookie)
	_, err := Conn.Exec(context.Background(), deleteCookie, cookie)
	return err
}

func CheckCookie(cookie, ip, userAgent string) error {
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	userIDReq := fmt.Sprintf("SELECT user_id, ttl FROM sessions WHERE cookie = '%s' ;", cookie)

	ctx := context.Background()
	rowsC, err := Conn.Query(ctx, userIDReq)
	if err != nil {
		logrus.Error(err)
		return err
	}
	var userID int
	var ttl time.Time
	defer rowsC.Close()
	for rowsC.Next() {
		rowsC.Scan(&userID, &ttl)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if !ttl.After(time.Now()) {
		logrus.Warn(ttl)
		DeleteCookie(cookie)
		return errors.New("cookie is expired")
	}

	loginReq := fmt.Sprintf("SELECT login FROM users WHERE id = '%d' ;", userID)

	rowsC, err = Conn.Query(ctx, loginReq)
	if err != nil {
		logrus.Error(err)
		return err
	}
	defer rowsC.Close()
	var login string
	for rowsC.Next() {
		rowsC.Scan(&login)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if cookie != crypto.CookieHash(ip, userAgent, login) {
		// logrus.Warn(cookie, "  ", crypto.CookieHash(ip, userAgent, login), " ", ip, " ", userAgent, " ", login)
		return errors.New("cookie not equel")
	}
	// } else {
	// 	return errors.New("incorrect login or password")
	// }
	return nil
}

func LoadNewOrder(cookie string, order int) error {
	//select user_id from sessions where cookie = cookie
	if Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	userIDReq := fmt.Sprintf("SELECT user_id FROM sessions WHERE cookie = '%s' ;", cookie)

	ctx := context.Background()
	rowsC, err := Conn.Query(ctx, userIDReq)
	if err != nil {
		logrus.Error(err)
		return err
	}
	var userID int
	defer rowsC.Close()
	for rowsC.Next() {
		rowsC.Scan(&userID)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return err
	}

	seluserIDReq := fmt.Sprintf("SELECT user_id FROM orders WHERE number = '%d' ;", order)

	rows, err := Conn.Query(ctx, seluserIDReq)
	if err != nil {
		logrus.Error(err)
		return err
	}
	selUserID := -1 // tochno ne sovpadet
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&selUserID)
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return err
	}
	if selUserID != userID && selUserID!=-1 {
		logrus.Error(selUserID, " ",userID)
		return errors.New("order belongs other user")
	}
	status := "NEW"
	insert := `INSERT INTO orders(user_id, number, status, uploaded_at) values($1, $2, $3, $4);`
	logrus.Info(insert, userID, order, status, time.Now().Format(time.RFC3339))
	_, err = Conn.Exec(context.Background(), insert, userID, order, status, time.Now().Format(time.RFC3339))
	return err
}

// func SaveDataToDB(sm *InMemory) {
// 	if Conn == nil {
// 		logrus.Error("Error nil Conn")
// 		return
// 	}
// 	sm.Lock()
// 	defer sm.Unlock()
// 	// TRUNCATE TABLE COMPANY
// 	// metrics := make([]types.Metrics, len(sm.MapCounter)+len(sm.MapGauge))

// 	keysCounter := make([]string, 0, len(sm.MapCounter))
// 	keysGauge := make([]string, 0, len(sm.MapGauge))

// 	for k := range sm.MapCounter {
// 		keysCounter = append(keysCounter, k)

// 	}

// 	for k := range sm.MapGauge {
// 		keysGauge = append(keysGauge, k)

// 	}
// 	for i := 0; i < len(keysCounter); i++ {
// 		insertCounter := `INSERT INTO praktikum(ID, Type, Delta) values($1, $2, $3)`
// 		_, err := Conn.Exec(context.Background(), insertCounter, keysCounter[i], "counter", int(sm.MapCounter[keysCounter[i]]))
// 		if err != nil {
// 			updateCounter := `UPDATE praktikum SET Type = $1, Delta = $2 WHERE ID = $3;`
// 			_, err := Conn.Exec(context.Background(), updateCounter, "counter", int(sm.MapCounter[keysCounter[i]]), keysCounter[i])
// 			if err != nil {
// 				logrus.Error(err)
// 			}
// 		}
// 	}
// 	for i := 0; i < len(keysGauge); i++ {
// 		insertGauge := `INSERT INTO praktikum(ID, Type, Value) values($1, $2, $3)`
// 		_, err := Conn.Exec(context.Background(), insertGauge, keysGauge[i], "gauge", float64(sm.MapGauge[keysGauge[i]]))
// 		if err != nil {
// 			updateGauge := `UPDATE praktikum SET Type = $1, Value = $2 WHERE ID = $3;`
// 			_, err := Conn.Exec(context.Background(), updateGauge, "gauge", float64(sm.MapGauge[keysGauge[i]]), keysGauge[i])
// 			if err != nil {
// 				logrus.Error(err)
// 			}
// 		}
// 	}

// }
// func RestoreDataFromDB(sm *InMemory) {
// 	if Conn == nil {
// 		logrus.Error("Error nil Conn")
// 		return
// 	}
// 	sm.Lock()
// 	defer sm.Unlock()
// 	// err := Conn.Ping()
// 	// if err != nil {
// 	// 	logrus.Error(err)
// 	// 	return
// 	// }
// 	ctx := context.Background()
// 	listCounter := "SELECT ID, Delta FROM praktikum WHERE Type='counter';"
// 	rowsC, err := Conn.Query(ctx, listCounter)
// 	if err != nil {
// 		logrus.Error(err)
// 		return
// 	}
// 	defer rowsC.Close()
// 	for rowsC.Next() {
// 		var id string
// 		var delta int64
// 		rowsC.Scan(&id, &delta)
// 		logrus.Info(id)
// 		logrus.Info(delta)
// 		sm.MapCounter[id] = types.Counter(delta)
// 	}

// 	err = rowsC.Err()
// 	if err != nil {
// 		logrus.Error(err)
// 	}
// 	listGauge := `SELECT ID, Value FROM praktikum WHERE Type='gauge';`
// 	rowsG, err := Conn.Query(ctx, listGauge)
// 	if err != nil {
// 		logrus.Error(err)
// 	}
// 	// c, _ := result
// 	defer rowsG.Close()
// 	for rowsG.Next() {
// 		var id string
// 		var value float64
// 		rowsG.Scan(&id, &value)
// 		logrus.Info(id)
// 		logrus.Info(value)
// 		sm.MapGauge[id] = types.Gauge(value)
// 	}
// 	err = rowsG.Err()
// 	if err != nil {
// 		logrus.Error(err)
// 	}
// }
