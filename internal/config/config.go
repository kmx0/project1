package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Address    string `env:"ADDRESS" envDefault:"127.0.0.1:8081"`
	Key        string `env:"KEY" `
	DBURI      string `env:"DATABASE_URI" envDefault:"postgres://postgres:postgres@localhost:5432/gophermart"`
	AccSysSddr string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://127.0.0.1:8080"`
	// -accrual-database-uri="***postgres/praktikum?sslmode=disable"
	// "postgres://postgres:postgres@localhost:5432/metrics"
}

func LoadConfig() Config {
	logrus.SetReportCaller(true)
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		logrus.Error(err)
	}
	return cfg
}

// func ReplaceUnusedInAgent(cfg *Config) {
// 	address := flag.String("a", "127.0.0.1:8080", "Address on server for Sending Metrics ")
// 	reportInterval := flag.Duration("r", 10000000000, "REPORT_INTERVAL")
// 	pollInterval := flag.Duration("p", 5000000000, "POLL_INTERVAL")
// 	key := flag.String("k", "", "KEY for hash")
// 	flag.Parse()
// 	if _, ok := os.LookupEnv("ADDRESS"); !ok {
// 		cfg.Address = *address
// 	}

// 	if _, ok := os.LookupEnv("REPORT_INTERVAL"); !ok {
// 		cfg.ReportInterval = *reportInterval
// 	}
// 	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
// 		cfg.PollInterval = *pollInterval
// 	}
// 	if _, ok := os.LookupEnv("KEY"); !ok {
// 		cfg.Key = *key
// 	}
// }

// func ReplaceUnusedInServer(cfg *Config) {
// 	//    = flag.Flag("aa", "Address on Listen").Short('a').Default("127.0.0.1:8080").String()
// 	address := flag.String("a", "127.0.0.1:8080", "Address on Listen")
// 	restore := flag.Bool("r", true, "restore from file or not")
// 	storeInterval := flag.Duration("i", 300000000000, "STORE_INTERVAL")
// 	storeFile := flag.String("f", "/tmp/devops-metrics-db.json", "STORE_FILE")
// 	dbDSN := flag.String("d", "", "database URI")
// 	key := flag.String("k", "", "KEY for hash")

// 	flag.Parse()

// 	if _, ok := os.LookupEnv("ADDRESS"); !ok {
// 		cfg.Address = *address
// 	}
// 	if _, ok := os.LookupEnv("RESTORE"); !ok {

// 		cfg.Restore = *restore
// 	}
// 	if _, ok := os.LookupEnv("STORE_INTERVAL"); !ok {

// 		cfg.StoreInterval = *storeInterval
// 	}
// 	if _, ok := os.LookupEnv("STORE_FILE"); !ok {
// 		cfg.StoreFile = *storeFile
// 	}
// 	if _, ok := os.LookupEnv("KEY"); !ok {
// 		cfg.Key = *key
// 	}
// 	logrus.Info(cfg.DBDSN)
// 	logrus.Info(*dbDSN)
// 	if _, ok := os.LookupEnv("DATABASE_DSN"); !ok {
// 		// if !strings.Contains(cfg.DBDSN, "incorr") {
// 		cfg.DBDSN = *dbDSN
// 	}
// 	// }
// }
