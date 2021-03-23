package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// initialize logging
func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	log.SetOutput(os.Stdout)
}

// AppName to store application name
var AppName string = "restockbot"

// AppVersion to set version at compilation time
var AppVersion string = "9999"

// GitCommit to set git commit at compilation time (can be empty)
var GitCommit string

// GoVersion to set Go version at compilation time
var GoVersion string

func main() {

	rand.Seed(time.Now().UnixNano())

	var err error
	config := NewConfig()

	version := flag.Bool("version", false, "Print version and exit")
	quiet := flag.Bool("quiet", false, "Log errors only")
	verbose := flag.Bool("verbose", false, "Print more logs")
	debug := flag.Bool("debug", false, "Print even more logs")
	databaseFileName := flag.String("database", AppName+".db", "Database file name")
	configFileName := flag.String("config", AppName+".json", "Configuration file name")
	logFileName := flag.String("log-file", "", "Log file name")
	disableNotifications := flag.Bool("disable-notifications", false, "Do not send notifications")
	workers := flag.Int("workers", 1, "number of workers for parsing shops")
	pidFile := flag.String("pid-file", "", "write process ID to this file to disable concurrent executions")
	pidWaitTimeout := flag.Int("pid-wait-timeout", 0, "seconds to wait before giving up when another instance is running")

	flag.Parse()

	if *version {
		showVersion()
		return
	}

	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *verbose {
		log.SetLevel(log.InfoLevel)
	}
	if *quiet {
		log.SetLevel(log.ErrorLevel)
	}

	if *logFileName != "" {
		fd, err := os.OpenFile(*logFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Printf("cannot open file for logging: %s\n", err)
		}
		log.SetOutput(fd)
	}

	if *configFileName != "" {
		err = config.Read(*configFileName)
		if err != nil {
			log.Fatalf("cannot parse configuration file: %s", err)
		}
	}
	log.Debugf("configuration file %s parsed", *configFileName)

	// handle PID file
	if *pidFile != "" {
		if err := waitPid(*pidFile, *pidWaitTimeout); err != nil {
			log.Warnf("%s", err)
			return
		}
		if err := writePid(*pidFile); err != nil {
			log.Fatalf("cannot write PID file: %s", err)
		}
		defer removePid(*pidFile)
	}

	// connect to the database
	db, err := gorm.Open(sqlite.Open(*databaseFileName), &gorm.Config{})
	if err != nil {
		log.Fatalf("cannot connect to database: %s", err)
	}
	log.Debugf("connected to database %s", *databaseFileName)

	// create tables
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatalf("cannot create products table")
	}
	if err := db.AutoMigrate(&Shop{}); err != nil {
		log.Fatalf("cannot create shops table")
	}

	// register notifiers
	notifiers := []Notifier{}

	if !*disableNotifications {
		if config.HasTwitter() {
			twitterNotifier, err := NewTwitterNotifier(&config.TwitterConfig, db)
			if err != nil {
				log.Fatalf("cannot create twitter client: %s", err)
			}
			notifiers = append(notifiers, twitterNotifier)
		}
	}

	// Group links by shop
	ShopsMap := make(map[string][]string)

	for _, link := range config.URLs {
		name, err := ExtractShopName(link)
		if err != nil {
			log.Warnf("cannot extract shop name from %s: %s", link, err)
		} else {
			ShopsMap[name] = append(ShopsMap[name], link)
		}
	}

	// parse asynchronously
	var wg sync.WaitGroup
	jobsCount := 0

	// start with URLs
	for shopName, shopLinks := range ShopsMap {

		// read shop from database or create it
		var shop Shop
		trx := db.Where(Shop{Name: shopName}).FirstOrCreate(&shop)
		if trx.Error != nil {
			log.Errorf("cannot create or select shop %s to/from database: %s", shopName, trx.Error)
			continue
		}

		for _, link := range shopLinks {
			if jobsCount < *workers {
				// create parser
				parser, err := NewURLParser(link, config.BrowserAddress, config.IncludeRegex, config.ExcludeRegex)
				if err != nil {
					log.Warnf("could not create URL parser for %s", link)
					continue
				}
				wg.Add(1)
				jobsCount++
				go handleProducts(shop, parser, notifiers, db, &wg)
			} else {
				log.Debugf("waiting for intermediate jobs to end")
				wg.Wait()
				jobsCount = 0
			}
		}
	}

	log.Debugf("waiting for all jobs to end")
	wg.Wait()
}

// For a given shop, fetch and parse its URL, then eventually send notifications
func handleProducts(shop Shop, parser *URLParser, notifiers []Notifier, db *gorm.DB, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Debugf("parsing with %s", parser)
	products, err := parser.Parse()
	if err != nil {
		log.Warnf("cannot parse: %s", err)
		return
	}
	log.Debugf("parsed")

	// upsert products to database
	for _, product := range products {

		log.Debugf("detected product %+v", product)

		if !product.IsValid() {
			log.Warnf("parsed malformatted product: %+v", product)
			continue
		}

		// check if product is already in the database
		// sometimes new products are detected on the website, directly available, without reference in the database
		// the bot has to send a notification instead of blindly creating it in the database and check availability afterwards
		var count int64
		trx := db.Model(&Product{}).Where(Product{URL: product.URL}).Count(&count)
		if trx.Error != nil {
			log.Warnf("cannot see if product %s already exists in the database: %s", product.Name, trx.Error)
			continue
		}

		// fetch product from database or create it if it doesn't exist
		var dbProduct Product
		trx = db.Where(Product{URL: product.URL}).Attrs(Product{Name: product.Name, Shop: shop, Price: product.Price, PriceCurrency: product.PriceCurrency, Available: product.Available}).FirstOrCreate(&dbProduct)
		if trx.Error != nil {
			log.Warnf("cannot fetch product %s from database: %s", product.Name, trx.Error)
			continue
		}
		log.Debugf("product %s found in database", dbProduct.Name)

		// detect availability change
		duration := time.Now().Sub(dbProduct.UpdatedAt).Truncate(time.Second)
		createThread := false
		closeThread := false

		// non-existing product directly available
		if count == 0 && product.Available {
			log.Infof("product %s on %s is now available", product.Name, shop.Name)
			createThread = true
		}

		// existing product with availability change
		if count > 0 && (dbProduct.Available != product.Available) {
			if product.Available {
				log.Infof("product %s on %s is now available", product.Name, shop.Name)
				createThread = true
			} else {
				log.Infof("product %s on %s is not available anymore", product.Name, shop.Name)
				closeThread = true
			}
		}

		// update product in database before sending notification
		// if there is a database failure, we don't want the bot to send a notification at each run
		if dbProduct.ToMerge(product) {
			dbProduct.Merge(product)
			trx = db.Save(&dbProduct)
			if trx.Error != nil {
				log.Warnf("cannot save product %s to database: %s", dbProduct.Name, trx.Error)
				continue
			}
			log.Debugf("product %s updated in database", dbProduct.Name)
		}

		// send notifications
		if createThread {
			for _, notifier := range notifiers {
				if err := notifier.NotifyWhenAvailable(shop.Name, dbProduct.Name, dbProduct.Price, dbProduct.PriceCurrency, dbProduct.URL); err != nil {
					log.Errorf("%s", err)
				}
			}
		} else if closeThread {
			for _, notifier := range notifiers {
				if err := notifier.NotifyWhenNotAvailable(dbProduct.URL, duration); err != nil {
					log.Errorf("%s", err)
				}
			}
		}
	}
}

func showVersion() {
	if GitCommit != "" {
		AppVersion = fmt.Sprintf("%s-%s", AppVersion, GitCommit)
	}
	fmt.Printf("%s version %s (compiled with %s)\n", AppName, AppVersion, GoVersion)
}
