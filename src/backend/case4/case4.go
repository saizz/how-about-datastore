package case4

import (
	"errors"
	"net/http"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func init() {

	http.HandleFunc("/case4", handleCase4)
	http.HandleFunc("/_ah/tq/hello", handleHello)
}

func handleCase4(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	e := false
	if r.FormValue("err") == "t" {
		e = true
	}

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {

		task := taskqueue.NewPOSTTask("/_ah/tq/hello", nil)
		if _, err := taskqueue.Add(ctx, task, "default"); err != nil {
			return err
		}

		if e {
			return errors.New("dummy error")
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf(ctx, "error: %v", err.Error())
	}
}

func handleHello(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	log.Infof(ctx, "hello")
}
