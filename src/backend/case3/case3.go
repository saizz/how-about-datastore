package case3

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

func init() {

	http.HandleFunc("/case3", handleCase3)
	http.HandleFunc("/_ah/tq/long-tx", handleLongTx)
}

type entity struct {
	Value int
}

type parser struct {
	err error
}

// ParseInt parse string to int.
func (p *parser) ParseInt(v string) int {

	if p.err != nil {
		return 0
	}

	value, err := strconv.Atoi(v)
	if err != nil {
		p.err = err
		return 0
	}

	return value
}

// Err return error, when parsing.
func (p *parser) Err() error {

	return p.err
}

func defaultValue(v, dv string) string {

	if v == "" {
		return dv
	}
	return v
}

func handleCase3(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	p := &parser{}
	n := p.ParseInt(defaultValue(r.FormValue("n"), "1"))
	if p.Err() != nil {
		log.Errorf(ctx, p.Err().Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	task := taskqueue.NewPOSTTask("/_ah/tq/long-tx", url.Values{"n": {strconv.Itoa(n)}})

	if _, err := taskqueue.Add(ctx, task, "default"); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func handleLongTx(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	p := &parser{}
	n := p.ParseInt(defaultValue(r.FormValue("n"), "1"))

	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {

		pid := "0000"
		pkey := datastore.NewKey(ctx, "case3-parent", pid, 0, nil)
		id := "0000"
		k := datastore.NewKey(ctx, "case3-child", id, 0, pkey)

		e := &entity{
			// Value: rand.Intn(q.count),
			Value: n,
		}
		if _, err := datastore.Put(ctx, k, e); err != nil {
			return err
		}

		for i := 0; i < n; i++ {
			log.Infof(ctx, "count: %v", i)
			time.Sleep(1 * time.Second)
		}
		return nil
	}, nil)

	if err != nil {
		log.Errorf(ctx, "error: %v", err.Error())
	} else {
		log.Infof(ctx, "count done")
	}
}
