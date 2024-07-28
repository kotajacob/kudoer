// License: AGPL-3.0-only
// (c) 2024 Dakota Walsh <kota@nilsu.org>
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"git.sr.ht/~kota/kudoer/application"
	"git.sr.ht/~kota/kudoer/db"
	"git.sr.ht/~kota/kudoer/litesession"
	"git.sr.ht/~kota/kudoer/mail"
	"git.sr.ht/~kota/kudoer/media"
	"git.sr.ht/~kota/kudoer/models"
	"git.sr.ht/~kota/kudoer/ui"
	"github.com/alexedwards/scs/v2"
	"github.com/throttled/throttled/v2"
	throttledstore "github.com/throttled/throttled/v2/store/memstore"
)

func main() {
	addr := flag.String("addr", ":2024", "HTTP Network Address")
	dsn := flag.String("dsn", "kudoer.db", "SQLite data source name")
	msn := flag.String("media", "media_store", "Media source name")
	mailHost := flag.String("mail-host", "", "Mail server host")
	mailPort := flag.Int("mail-port", 25, "Mail server port")
	mailUsername := flag.String("mail-username", "", "Mail server username")
	mailPassword := flag.String("mail-password", "", "Mail server password")
	mailSender := flag.String("mail-sender", "Kudoer <no-reply@kudoer.com>", "Mail server sender")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("opening database:", *dsn)
	db, err := db.Open(*dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	mailer := mail.New(
		*mailHost,
		*mailPort,
		*mailUsername,
		*mailPassword,
		*mailSender,
	)

	mediaStore, err := media.Open(*msn)
	if err != nil {
		log.Fatal(err)
	}

	templates, err := ui.Templates()
	if err != nil {
		errLog.Fatal(err)
	}

	// Setup session storage.
	sessionManager := scs.New()
	sessionManager.Store = litesession.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.ErrorFunc = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
		errLog.Output(2, trace)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	// Set up HTTP request throttling.
	tstore, err := throttledstore.NewCtx(65536)
	if err != nil {
		errLog.Fatal(err)
	}
	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(20),
		MaxBurst: 5,
	}
	throttler, err := throttled.NewGCRARateLimiterCtx(tstore, quota)
	if err != nil {
		errLog.Fatal(err)
	}
	rateLimiter := &throttled.HTTPRateLimiterCtx{
		RateLimiter: throttler,
		VaryBy: &throttled.VaryBy{
			Path:    true,
			Method:  true,
			Headers: []string{"X-Forwarded-For"},
		},
	}

	app := application.New(
		infoLog,
		errLog,
		templates,
		sessionManager,
		rateLimiter,
		mediaStore,
		mailer,
		&models.UserModel{DB: db},
		&models.ItemModel{DB: db},
		&models.KudoModel{DB: db},
		&models.SearchModel{DB: db},
		&models.PWResetModel{DB: db},
	)

	err = app.Serve(*addr)
	if err != nil {
		errLog.Fatalln(err)
	}
}
