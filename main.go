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
	"git.sr.ht/~kota/kudoer/application/mail"
	"git.sr.ht/~kota/kudoer/application/media"
	"git.sr.ht/~kota/kudoer/config"
	"git.sr.ht/~kota/kudoer/db"
	"git.sr.ht/~kota/kudoer/db/models"
	"git.sr.ht/~kota/kudoer/ui"
	"git.sr.ht/~kota/zqlsession"
	"github.com/alexedwards/scs/v2"
	"github.com/throttled/throttled/v2"
	throttledstore "github.com/throttled/throttled/v2/store/memstore"
)

func main() {
	cfgPath := flag.String(
		"config",
		"/etc/kudoer/config.toml",
		"Path to configuration file",
	)
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stdout, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		errLog.Fatal(err)
	}

	infoLog.Println("opening database:", cfg.DSN)
	db, err := db.Open(cfg.DSN)
	if err != nil {
		errLog.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			errLog.Fatal(err)
		}
	}()

	mailer := mail.New(
		cfg.MailHost,
		cfg.MailPort,
		cfg.MailUsername,
		cfg.MailPassword,
		cfg.MailSender,
	)

	mediaStore, err := media.Open(cfg.MSN)
	if err != nil {
		errLog.Fatal(err)
	}

	templates, err := ui.Templates()
	if err != nil {
		errLog.Fatal(err)
	}

	// Setup session storage.
	sessionManager := scs.New()
	sessionManager.Store = zqlsession.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.ErrorFunc = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
		_ = errLog.Output(2, trace) // Ignore failed error logging.
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
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
		&models.ProfilePictureModel{DB: db},
	)

	err = app.Serve(cfg.Addr)
	if err != nil {
		errLog.Fatalln(err)
	}
}
