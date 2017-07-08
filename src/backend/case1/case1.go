package case1

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

	http.HandleFunc("/case1", handleCase1)
	http.HandleFunc("/case1-inc-value", handleCase1IncValue)

	rand.Seed(time.Now().UnixNano())
}

type entity struct {
	Value int
}

type query struct {
	concurrent int
	child      int
	sleep      int
	summary    bool
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

func (p *parser) ParseBool(v string) bool {

	if p.err != nil {
		return false
	}

	if v == "t" {
		return true
	}

	return false
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
	child := p.ParseInt(defaultValue(r.FormValue("child"), "1"))
	sleep := p.ParseInt(defaultValue(r.FormValue("sleep"), "0"))
	summary := p.ParseBool(r.FormValue("summary"))

	err := p.Err()
	if err != nil {
		return nil, errors.Wrap(err, "at parseQuery")
	}

	return &query{
		concurrent: concurrent,
		child:      child,
		sleep:      sleep,
		summary:    summary,
	}, nil

}

func put(ctx context.Context, q *query, c int) error {

	pid := "0000"
	pkey := datastore.NewKey(ctx, "case1-parent", pid, 0, nil)

	me, any := make(appengine.MultiError, q.child), false
	for i := 0; i < q.child; i++ {

		id := fmt.Sprintf("%04d", i)
		k := datastore.NewKey(ctx, "case1-child", id, 0, pkey)
		e := &entity{
			Value: rand.Intn(q.concurrent * q.child * 100),
			//Value: c,
		}

		_, err := datastore.Put(ctx, k, e)
		if !q.summary {
			if err != nil {
				log.Errorf(ctx, "put, concurrent=%v, child=%v: %v", c, i, err.Error())
				any = true
			} else {
				log.Infof(ctx, "put, concurrent=%v, child=%v: OK", c, i)
			}
		}
		me[i] = err

		time.Sleep(time.Duration(q.sleep) * time.Second)
	}

	if q.summary {
		if any {
			log.Errorf(ctx, "any put, concurrent=%v: %v", c, me.Error())
		} else {
			log.Infof(ctx, "all put, concurrent=%v: OK", c)
		}
	}

	if any {
		return me
	}
	return nil
}

func putInTransaction(ctx context.Context, q *query, c int) error {

	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {

		return put(ctx, q, c)
	}, nil)

}

func handleCase1(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	q, err := parseQuery(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "at handleCase1").Error(), http.StatusBadRequest)
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
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(m)
	w.Write(b)
}

func inc(ctx context.Context, q *query, c int) error {

	pid := "0000"
	pkey := datastore.NewKey(ctx, "case1-parent", pid, 0, nil)

	id := "0000"
	k := datastore.NewKey(ctx, "case1-child", id, 0, pkey)

	e := &entity{}

	err := datastore.Get(ctx, k, e)
	if err != nil {
		log.Errorf(ctx, "get, concurrent=%v: %v", c, err.Error())
		return err
	}

	me, any := make(appengine.MultiError, q.child), false
	for i := 0; i < q.child; i++ {

		e.Value++
		_, err = datastore.Put(ctx, k, e)
		if err != nil {
			log.Errorf(ctx, "put, concurrent=%v, child=%v, incremented value=%v : %v", c, i, e.Value, err.Error())
			any = true
		} else {
			log.Infof(ctx, "put, concurrent=%v, child=%v, incremented value=%v : OK", c, i, e.Value)
		}
		me[i] = err

		time.Sleep(time.Duration(q.sleep) * time.Second)
	}

	if any {
		return me
	}
	return nil
}

func incInTransaction(ctx context.Context, q *query, c int) error {

	return datastore.RunInTransaction(ctx, func(ctx context.Context) error {

		return inc(ctx, q, c)
	}, nil)

}

func handleCase1IncValue(w http.ResponseWriter, r *http.Request) {

	ctx := appengine.NewContext(r)

	q := &query{
		concurrent: 1,
		child:      1,
		sleep:      0,
	}
	put(ctx, q, 0)

	q, err := parseQuery(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "at handleCase1IncValue").Error(), http.StatusBadRequest)
		return
	}

	ch := make(chan error, q.concurrent)
	var wg sync.WaitGroup
	go func() {
		for c := 0; c < q.concurrent; c++ {
			wg.Add(1)
			go func(c int) {
				ch <- incInTransaction(ctx, q, c)
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
		w.WriteHeader(http.StatusInternalServerError)
	}

	b, _ := json.Marshal(m)
	w.Write(b)
}
