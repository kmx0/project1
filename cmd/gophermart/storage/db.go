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

type PostgresStorage struct {
	DB     *sql.DB
	Conn   *pgx.Conn
	DBName string
}

func NewDB() Storage {
	return &PostgresStorage{DBName: "gophermart"}
}

func (ps *PostgresStorage) PingDB(ctx context.Context, urlExample string) bool {
	// urlExample = "postgres://postgres:postgres@localhost:5432/gophermart"
	logrus.Info(urlExample)
	var err error
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	ps.Conn, err = pgx.Connect(context.Background(), urlExample)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	// defer conn.Close(context.Background())

	err = ps.Conn.Ping(context.Background())
	if err != nil {
		logrus.Error(err)
		return false
	}

	logrus.Info("Successfully connected!")
	logrus.Info(checkDBExist(ps.Conn, ps.DBName))
	// if !CheckDBExist() {

	// }
	TableName := "users"
	if !checkTableExist(ps.Conn, TableName) {
		req := fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			login varchar(255) UNIQUE,
			password varchar(255),
			balance real
			);`, TableName)
		addTabletoDB(ps.Conn, req)
	}
	TableName = "sessions"

	if !checkTableExist(ps.Conn, TableName) {
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
		addTabletoDB(ps.Conn, req)
	}
	TableName = "orders"

	if !checkTableExist(ps.Conn, TableName) {
		req := fmt.Sprintf(`CREATE TABLE %s (
					id SERIAL PRIMARY KEY,
					user_id  integer,
					number varchar(255) UNIQUE,
					status varchar(255),
					accrual real,
					uploaded_at timestamp with time zone,
					CONSTRAINT fk_user
					FOREIGN KEY(user_id)
					REFERENCES users(id)
					ON DELETE CASCADE
					);`, TableName)
		addTabletoDB(ps.Conn, req)
	}
	TableName = "withdrawals"

	if !checkTableExist(ps.Conn, TableName) {
		req := fmt.Sprintf(`CREATE TABLE %s (
					id SERIAL PRIMARY KEY,
					user_id  integer,
					order_id varchar(255) UNIQUE,
					withdraw real,
					processed_at timestamp with time zone,
					CONSTRAINT fk_user
					FOREIGN KEY(user_id)
					REFERENCES users(id)
					ON DELETE CASCADE
					);`, TableName)
		addTabletoDB(ps.Conn, req)
	}

	// logrus.Info(CheckTableExist())
	return true
}

func checkDBExist(Conn *pgx.Conn, DBName string) bool {
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

func checkTableExist(Conn *pgx.Conn, TableName string) bool {
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

func addTabletoDB(Conn *pgx.Conn, req string) {
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
func (ps *PostgresStorage) RegisterUser(user types.User) (id int, err error) {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return 0, errors.New("error nil Conn")
	}
	// checkExist := fmt.Sprintf("SELECT users (SELECT 1 FROM users WHERE Login = %s LIMIT 1);", user.Login)
	insert := `INSERT INTO users(login, password, balance) values($1, $2, $3) RETURNING id`

	err = ps.Conn.QueryRow(context.Background(), insert, user.Login, user.Password, user.Balance).Scan(&id)

	return id, err
}

func (ps *PostgresStorage) LoginUser(user types.User) (id int, cookie string, err error) {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return 0, "", errors.New("error nil Conn")
	}
	loginReq := fmt.Sprintf("SELECT id, login, password FROM users WHERE login = '%s' ;", user.Login)

	ctx := context.Background()
	rowsC, err := ps.Conn.Query(ctx, loginReq)
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

func (ps *PostgresStorage) WriteUserCookie(user types.User, id int) error {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	CookieForDelReq := fmt.Sprintf("SELECT cookie FROM sessions WHERE user_id = '%d' ;", user.ID)
	ctx := context.Background()
	rowsC, err := ps.Conn.Query(ctx, CookieForDelReq)
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
		_, err = ps.Conn.Exec(context.Background(), insert, user.Cookie, time.Now().Add(time.Hour*1), user.ID)
	} else {
		update := `UPDATE sessions SET ttl = $1 WHERE user_id = $2;`
		_, err = ps.Conn.Exec(context.Background(), update, time.Now().Add(time.Hour*1), user.ID)
		if err != nil {
			logrus.Error(err)
		}
	}
	return err
}

func (ps *PostgresStorage) DeleteCookie(cookie string) error {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	deleteCookie := `DELETE FROM sessions WHERE cookie = $1;`

	// logrus.Info(deleteCookie)
	_, err := ps.Conn.Exec(context.Background(), deleteCookie, cookie)
	return err
}

func (ps *PostgresStorage) CheckCookie(cookie, ip, userAgent string) error {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	userIDReq := fmt.Sprintf("SELECT user_id, ttl FROM sessions WHERE cookie = '%s' ;", cookie)

	ctx := context.Background()
	rowsC, err := ps.Conn.Query(ctx, userIDReq)
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
		ps.DeleteCookie(cookie)
		return errors.New("cookie is expired")
	}

	loginReq := fmt.Sprintf("SELECT login FROM users WHERE id = '%d' ;", userID)

	rowsC, err = ps.Conn.Query(ctx, loginReq)
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

func (ps *PostgresStorage) LoadNewOrder(cookie string, order string) error {
	//select user_id from sessions where cookie = cookie
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	userIDReq := fmt.Sprintf("SELECT user_id FROM sessions WHERE cookie = '%s' ;", cookie)

	ctx := context.Background()
	rowsC, err := ps.Conn.Query(ctx, userIDReq)
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

	seluserIDReq := fmt.Sprintf("SELECT user_id FROM orders WHERE number = '%s' ;", order)

	rows, err := ps.Conn.Query(ctx, seluserIDReq)
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
	if selUserID != userID && selUserID != -1 {
		logrus.Error(selUserID, " ", userID)
		return errors.New("order belongs other user")
	}
	status := "NEW"
	insert := `INSERT INTO orders(user_id, number, status, uploaded_at) values($1, $2, $3, $4);`
	// logrus.Info(insert, userID, order, status, time.Now().Format(time.RFC3339))
	_, err = ps.Conn.Exec(context.Background(), insert, userID, order, status, time.Now().Format(time.RFC3339))
	return err
}

func (ps *PostgresStorage) GetOrdersList(cookie string) ([]types.Order, error) {
	//select user_id from sessions where cookie = cookie
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return nil, errors.New("error nil Conn")
	}
	userIDReq := fmt.Sprintf("SELECT user_id FROM sessions WHERE cookie = '%s' ;", cookie)

	ctx := context.Background()
	rowsC, err := ps.Conn.Query(ctx, userIDReq)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	var userID int
	defer rowsC.Close()
	for rowsC.Next() {
		rowsC.Scan(&userID)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	seluserIDReq := fmt.Sprintf("SELECT number,status,accrual,uploaded_at FROM orders WHERE user_id = '%d' ORDER BY  uploaded_at DESC;", userID)

	rows, err := ps.Conn.Query(ctx, seluserIDReq)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	var order types.Order
	ordersList := make([]types.Order, 0)
	defer rows.Close()
	for rows.Next() {
		var nullInt sql.NullInt32
		rows.Scan(&order.Number, &order.Status, &nullInt, &order.UploadedAt)
		if nullInt.Valid {
			order.Accrual = int(nullInt.Int32)
		}
		ordersList = append(ordersList, order)
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return ordersList, err
	}
	return ordersList, err
}
func (ps *PostgresStorage) WriteAccrual(accrual types.AccrualO) error {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}

	// err = ps.Conn.QueryRow(context.Background(), insert, user.Login, user.Password, user.Balance).Scan(&id)

	var id int
	update := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING user_id;`
	err := ps.Conn.QueryRow(context.Background(), update, accrual.Status, accrual.Accrual, accrual.Order).Scan(&id)
	if err != nil {
		logrus.Error(err)
		return err
	}
	if accrual.Accrual != 0 {
		return ps.ChangeBalanceValue(accrual.Accrual, "+", id)
	} else {
		return nil
	}
}
func (ps *PostgresStorage) ChangeBalanceValue(value float64, action string, userID int) error {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}
	updateBal := fmt.Sprintf(`UPDATE users SET balance = balance %s $1 WHERE id = $2;`, action)
	// logrus.Info(insert, userID, order, status, time.Now().Format(time.RFC3339))
	_, err := ps.Conn.Exec(context.Background(), updateBal, value, userID)
	return err
}

func (ps *PostgresStorage) GetUserID(cookie string) (id int, err error) {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return -1, errors.New("error nil Conn")
	}
	userIDReq := fmt.Sprintf("SELECT user_id FROM sessions WHERE cookie = '%s' ;", cookie)

	ctx := context.Background()
	rowsC, err := ps.Conn.Query(ctx, userIDReq)
	if err != nil {
		logrus.Error(err)
		return -1, err
	}
	defer rowsC.Close()
	for rowsC.Next() {
		rowsC.Scan(&id)
	}
	err = rowsC.Err()
	if err != nil {
		logrus.Error(err)
		return -1, err
	}
	return id, err
}

func (ps *PostgresStorage) GetBalance(cookie string) (balance float64, err error) {
	//request to table session return user_id
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return balance, errors.New("error nil Conn")
	}
	//request to table users return balance
	userID, err := ps.GetUserID(cookie)
	if err != nil {
		return balance, err
	}
	selbalanceReq := fmt.Sprintf("SELECT balance FROM users WHERE id = '%d';", userID)
	ctx := context.Background()
	rows, err := ps.Conn.Query(ctx, selbalanceReq)
	if err != nil {
		logrus.Error(err)
		return balance, err
	}
	defer rows.Close()
	for rows.Next() {
		var NullFloat64 sql.NullFloat64
		rows.Scan(&NullFloat64)
		if NullFloat64.Valid {
			balance = NullFloat64.Float64
		}
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return balance, err
	}
	return balance, err

}

func (ps *PostgresStorage) GetSUMWithdraws(userID int) (withdrawals float64, err error) {
	//request to table session return user_id
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return withdrawals, errors.New("error nil Conn")
	}

	//request to table withdrawals return sum withdraws
	ctx := context.Background()
	sumWithdrawalsReq := fmt.Sprintf("SELECT SUM (withdraw) AS sumw FROM withdrawals WHERE user_id = '%d';", userID)
	rows, err := ps.Conn.Query(ctx, sumWithdrawalsReq)
	if err != nil {
		logrus.Error(err)
		return withdrawals, err
	}
	defer rows.Close()
	for rows.Next() {
		var NullFloat64 sql.NullFloat64
		rows.Scan(&NullFloat64)
		if NullFloat64.Valid {
			withdrawals = NullFloat64.Float64
		}
	}
	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return withdrawals, err
	}
	return withdrawals, err
}
func (ps *PostgresStorage) GetWithdrawals(userID int) ([]types.Withdraw, error) {
	//request to table session return user_id
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return nil, errors.New("error nil Conn")
	}

	//request to table withdrawals return sum withdraws
	ctx := context.Background()
	sumWithdrawalsReq := fmt.Sprintf("SELECT order_id, withdraw, processed_at FROM withdrawals WHERE user_id = '%d' ORDER BY  processed_at;", userID)
	rows, err := ps.Conn.Query(ctx, sumWithdrawalsReq)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	defer rows.Close()
	var withdraw types.Withdraw
	withdrawals := make([]types.Withdraw, 0)
	for rows.Next() {
		var nullFloat64 sql.NullFloat64
		rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAT)
		if nullFloat64.Valid {
			withdraw.Sum = nullFloat64.Float64
		}
		withdrawals = append(withdrawals, withdraw)
	}

	err = rows.Err()
	if err != nil {
		logrus.Error(err)
		return withdrawals, err
	}
	return withdrawals, err
}

func (ps *PostgresStorage) WriteWithdraw(withdraw types.Withdraw, userID int) error {
	if ps.Conn == nil {
		logrus.Error("Error nil Conn")
		return errors.New("error nil Conn")
	}

	update := `INSERT INTO withdrawals (user_id, order_id, withdraw, processed_at) values($1,$2,$3,$4);`
	_, err := ps.Conn.Exec(context.Background(), update, userID, withdraw.Order, withdraw.Sum, time.Now().Format(time.RFC3339))
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
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

// func (sm *Storage) GetCounter(metricType string, metric string) (types.Counter, error) {

// 	return sm.MapCounter[metric], nil
// }
