package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	osuapi "github.com/thehowl/go-osuapi"

	"github.com/mempler/cheesegull/api"
	"github.com/mempler/cheesegull/config"
	"github.com/mempler/cheesegull/dbmirror"
	"github.com/mempler/cheesegull/downloader"
	"github.com/mempler/cheesegull/housekeeper"
	"github.com/mempler/cheesegull/models"

	// Components of the API we want to use
	_ "github.com/mempler/cheesegull/api/download"
	_ "github.com/mempler/cheesegull/api/metadata"
)

const searchDSNDocs = `"DSN to use for fulltext searches. ` +
	`This should be a SphinxQL server. Follow the format of the MySQL DSN. ` +
	`This can be the same as MYSQL_DSN, and cheesegull will still run ` +
	`successfully, however what happens when search is tried is undefined ` +
	`behaviour and you should definetely bother to set it up (follow the README).`

var (
	conf = config.Parse()
)

func addTimeParsing(dsn string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	dsn += sep + "parseTime=true&multiStatements=true"
	return dsn
}

func main() {
	fmt.Println("CheeseGull", Version)
	api.Version = Version

	// set up osuapi client
	c := osuapi.NewClient(conf.Osu.APIKey)

	// set up downloader
	d, err := downloader.LogIn(conf.Osu.Username, conf.Osu.Password, conf.Osu.DownloadHostname)
	if err != nil {
		fmt.Println("Can't log in into osu!:", err)
		os.Exit(1)
	}
	dbmirror.SetHasVideo(d.HasVideo)

	// set up mysql
	db, err := sql.Open("mysql", addTimeParsing(conf.MySQL.Username+":"+conf.MySQL.Password+"@tcp("+conf.MySQL.Hostname+":"+strconv.Itoa(conf.MySQL.Port)+")/"+conf.MySQL.Database))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set up search
	db2, err := sql.Open("mysql", conf.SphinxQL.Username+":"+conf.SphinxQL.Password+"@"+conf.SphinxQL.Hostname+":"+strconv.Itoa(conf.SphinxQL.Port)+"/"+conf.SphinxQL.Database)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// set up housekeeper
	house := housekeeper.New()
	err = house.LoadState()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	house.MaxSize = uint64(float64(1024*1024*1024) * (conf.Server.BMCacheSize))
	house.StartCleaner()

	// run mysql migrations
	err = models.RunMigrations(db)
	if err != nil {
		fmt.Println("Error running migrations", err)
	}

	// start running components of cheesegull
	if conf.Server.ShouldDiscover {
		go dbmirror.StartSetUpdater(c, db)
		go dbmirror.DiscoverEvery(c, db, time.Hour*6, time.Second*20)
	}

	// create request handler
	panic(http.ListenAndServe(conf.Server.Hostname+":"+strconv.Itoa(conf.Server.Port), api.CreateHandler(db, db2, house, d)))
}
