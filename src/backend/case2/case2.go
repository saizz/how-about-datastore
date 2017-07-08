package case2

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"golang.org/x/net/context"

	"strconv"

	"fmt"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func init() {

	http.HandleFunc("/case2", handleCase2)

	rand.Seed(time.Now().UnixNano())
}

type entity struct {
	Value int
}

type query struct {
	concurrent int
	parent     int
	child      int
	sleep      int
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

func parseQuery(r *http.Request) (*query, error) {

	p := &parser{}

	concurrent := p.ParseInt(defaultValue(r.FormValue("concurrent"), "1"))
	parent := p.ParseInt(defaultValue(r.FormValue("parent"), "1"))
	child := p.ParseInt(defaultValue(r.FormValue("child"), "1"))
	sleep := p.ParseInt(defaultValue(r.FormValue("sleep"), "0"))

	err := p.Err()
	if err != nil {
		return nil, errors.Wrap(err, "at parseQuery")
	}

	return &query{
		concurrent: concurrent,
		parent:     parent,
		child:      child,
		sleep:      sleep,
	}, nil

}

func put(ctx context.Context, q *query, c int) error {

	me, any := make(appengine.MultiError, q.parent*q.child), false

	for j := 0; j < q.parent; j++ {
		pid := fmt.Sprintf("%04d", j)
		pkey := datastore.NewKey(ctx, "case2-parent", pid, 0, nil)

		for i := 0; i < q.child; i++ {

			id := fmt.Sprintf("%04d", i)
			k := datastore.NewKey(ctx, "case2-child", id, 0, pkey)
			e := &entity{
				// Value: rand.Intn(q.count),
				Value: c,
			}

			_, err := datastore.Put(ctx, k, e)
			if err != nil {
				log.Errorf(ctx, "put, concurrent=%v, parent=%v, child=%v: %v", c, j, i, err.Error())
				any = true
			} else {
				log.Infof(ctx, "put, concurrent=%v, parent=%v, child=%v: OK", c, j, i)
			}
			me[i] = err

			time.Sleep(time.Duration(q.sleep) * time.Second)
		}
	}

	if any {
		return me
	}
	return nil
}

func putInTransaction(ctx context.Context, q *query, c int) error {

	o := &datastore.TransactionOptions{
		XG: true,
	}
	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {

		return put(ctx, q, c)
	}, o)

}

func handleCase2(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	q, err := parseQuery(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "at handleCase2").Error(), http.StatusBadRequest)
		return
	}

	ch := make(chan error, q.concurrent)
	var wg sync.WaitGroup
	go func() {
		for c := 0; c < q.concurrent; c++ {
			wg.Add(1)
			go func(c int) {
				ch <- putInTransaction(ctx, q, c)
				wg.Done()
			}(c)
		}
		wg.Wait()
		close(ch)
	}()

	me, any := make(appengine.MultiError, 0, q.concurrent), false
	for {
		err, ok := <-ch
		if !ok {
			break
		}

		if err != nil {
			me = append(me, err)
			any = true
		}
	}

	w.Header().Set("Context-Type", "application/json")
	m := &struct {
		Message string
	}{
		Message: "OK",
	}
	if any {
		m.Message = me.Error()
		//log.Errorf(ctx, me.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(m)
	w.Write(b)
}
