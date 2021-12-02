//
// REST
// ====
// This example demonstrates a HTTP REST web service with some fixture data.
// Follow along the example and patterns.
//
// Also check routes.json for the generated docs from passing the -routes flag,
// to run yourself do: `go run . -routes`
//
// Boot the server:
// ----------------
// $ go run main.go
//
// Client requests:
// ----------------
// $ curl http://localhost:3333/
// root.
//
// $ curl http://localhost:3333/articles
// [{"id":"1","title":"Hi"},{"id":"2","title":"sup"}]
//
// $ curl http://localhost:3333/articles/1
// {"id":"1","title":"Hi"}
//
// $ curl -X DELETE http://localhost:3333/articles/1
// {"id":"1","title":"Hi"}
//
// $ curl http://localhost:3333/articles/1
// "Not Found"
//
// $ curl -X POST -d '{"id":"will-be-omitted","title":"awesomeness"}' http://localhost:3333/articles
// {"id":"97","title":"awesomeness"}
//
// $ curl http://localhost:3333/articles/97
// {"id":"97","title":"awesomeness"}
//
// $ curl http://localhost:3333/articles
// [{"id":"2","title":"sup"},{"id":"97","title":"awesomeness"}]
//
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/docgen"
	"github.com/go-chi/render"
)

var routes = flag.Bool("routes", false, "Generate router documentation")

func main() {
	flag.Parse()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("root."))
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	})

	// $ curl http://localhost:3333/	// root.
	// $ curl http://localhost:3333/articles	// [{"id":"1","title":"Hi"},{"id":"2","title":"sup"}]
	// $ curl http://localhost:3333/articles/1	// {"id":"1","title":"Hi"}
	// $ curl -X DELETE http://localhost:3333/articles/1	// {"id":"1","title":"Hi"}
	// $ curl http://localhost:3333/articles/1	// "Not Found"
	// $ curl -X POST -d '{"id":"will-be-omitted","title":"awesomeness"}' http://localhost:3333/articles	// {"id":"97","title":"awesomeness"}
	// $ curl http://localhost:3333/articles/97	// {"id":"97","title":"awesomeness"}
	// $ curl http://localhost:3333/articles	// [{"id":"2","title":"sup"},{"id":"97","title":"awesomeness"}]

	/* $ curl http://bangkokguy.ddns.net/rest/v1/device // {"ip":"192.168.1.1","ssid":"MrWhite","passphrase":"f","currenttime":"08:00"}
	 * $ curl http://bangkokguy.ddns.net/rest/v1/temp // {"currenttemp":"24.00","nighttemp":"18.00","daytemp":"24.00","thereshold":"0.20"}
	 * $ curl http://bangkokguy.ddns.net/rest/v1/time // {"day":"06:00","night":"22:00"}
	 * $ curl http://bangkokguy.ddns.net/rest/v1/mode // {"mode":{"night|day" "auto|manual"},"heating":{"off":"manual|auto"}}
	 * $ curl -X PUT -d '{"day":"24.00","night":"18.00"}' http://bangkokguy.ddns.net/rest/v1/temp
	 * $ curl -X PUT -d '{"day":"06:00","night":"22:00"}' http://bangkokguy.ddns.net/rest/v1/time
	 * $ curl -X PUT -d '{"ssid":"Faszom","passphrase":"f"}' http://bangkokguy.ddns.net/rest/v1/device
	 * $ curl -X PUT -d '{"mode":"night|day|auto","heating":"on|off|auto"}' http://bangkokguy.ddns.net/rest/v1/mode
	 */

	// RESTy routes for "articles" resource
	r.Route("/rest/v1",
		func(r chi.Router) {
			r.With(paginate).Get("/", ListArticles)
			r.Post("/", CreateArticle)  // POST /articles
			r.Get("/device", GetDevice) // GET /rest/v1/device

			r.Route("/time",
				func(r chi.Router) {
					r.Get("/", GetTime)    // GET /temp
					r.Put("/", UpdateTime) // PUT /temp
				},
			)
			r.Route("/temp",
				func(r chi.Router) {
					r.Get("/", GetTemp)    // GET /temp
					r.Put("/", UpdateTemp) // PUT /temp
				},
			)
			r.Route("/mode",
				func(r chi.Router) {
					r.Get("/", GetMode)    // GET /temp
					r.Put("/", UpdateMode) // PUT /temp
				},
			)
			r.Route("/{articleID}",
				func(r chi.Router) {
					r.Use(ArticleCtx)            // Load the *Article on the request context
					r.Get("/", GetArticle)       // GET /articles/123
					r.Put("/", UpdateArticle)    // PUT /articles/123
					r.Delete("/", DeleteArticle) // DELETE /articles/123
				},
			)

			// GET /articles/whats-up
			r.With(ArticleCtx).Get("/{articleSlug:[a-z-]+}", GetArticle)
		})

	// Mount the admin sub-router, which btw is the same as:
	// r.Route("/admin", func(r chi.Router) { admin routes here })
	r.Mount("/admin", adminRouter())

	// Passing -routes to the program will generate docs for the above
	// router definition. See the `routes.json` file in this folder for
	// the output.
	if *routes {
		// fmt.Println(docgen.JSONRoutesDoc(r))
		fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
			ProjectPath: "github.com/go-chi/chi/v5",
			Intro:       "Welcome to the chi/_examples/rest generated docs.",
		}))
		return
	}

	http.ListenAndServe(":3333", r)
}

func ListArticles(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, NewArticleListResponse(articles)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// ArticleCtx middleware is used to load an Article object from
// the URL parameters passed through as the request. In case
// the Article could not be found, we stop here and return a 404.
func ArticleCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var article *Article
		var err error

		if articleID := chi.URLParam(r, "articleID"); articleID != "" {
			article, err = dbGetArticle(articleID)
		} else if articleSlug := chi.URLParam(r, "articleSlug"); articleSlug != "" {
			article, err = dbGetArticleBySlug(articleSlug)
		} else {
			render.Render(w, r, ErrNotFound)
			return
		}
		if err != nil {
			render.Render(w, r, ErrNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "article", article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SearchArticles searches the Articles data for a matching article.
// It's just a stub, but you get the idea.
func SearchArticles(w http.ResponseWriter, r *http.Request) {
	//render.RenderList(w, r, NewArticleListResponse(articles))
	render.Render(w, r,
		&ErrResponse{
			HTTPStatusCode: 400,
			StatusText:     "Invalid request.",
		})
}

// CreateArticle persists the posted Article and returns it
// back to the client as an acknowledgement.
func CreateArticle(w http.ResponseWriter, r *http.Request) {
	data := &ArticleRequest{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	article := data.Article
	dbNewArticle(article)

	render.Status(r, http.StatusCreated)
	render.Render(w, r, NewArticleResponse(article))
}

var IP string = "192.168.1.123"
var SSID = "MrWhite"
var PassPhrase = "F"
var CurrentTime = time.Now().Format("2006-01-02 15:04:05")
var CurrentTemp = "22.00"
var NightTemp = "18.00"
var DayTemp = "24.00"
var Thereshold = "0.20"
var Day string = "06:00"
var Night string = "22:00"
var CurrentStateOfMode = "night"
var CurrentStateOfHeating = "off"
var Mode [2]string = [2]string{CurrentStateOfMode, "auto"}
var Heating [2]string = [2]string{CurrentStateOfHeating, "auto"}

/**-----------------------------------------------------------------------------------
 * get device
 * ==========
 * $ curl http://bangkokguy.ddns.net/rest/v1/device // {"ip":"192.168.1.1","ssid":"MrWhite","passphrase":"f","currenttime":"08:00"}
 *------------------------------------------------------------------------------------*/
type Device struct {
	IP          string `json:"ip"`
	SSID        string `json:"ssid"`
	PassPhrase  string `json:"passphrase"`
	CurrentTime string `json:"currenttime"`
}

func GetDevice(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, dbGetDevice()); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
func (rd *Device) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}
func dbGetDevice() *Device {
	var u Device
	u.CurrentTime = CurrentTime //"20:00:00"
	u.IP = IP                   //"192.168.1.123"
	u.PassPhrase = PassPhrase   //"F"
	u.SSID = SSID               //"MrWhite"
	return &u
	//return nil, errors.New("user not found.")
}

/**-----------------------------------------------------------------------------------
* get temp
* ==========
* $ curl http://bangkokguy.ddns.net/rest/v1/temp //
		  {"currenttemp":"24.00","nighttemp":"18.00","daytemp":"24.00","thereshold":"0.20"}
*------------------------------------------------------------------------------------*/
type Temp struct {
	CurrentTemp string `json:"currenttemp"`
	NightTemp   string `json:"nighttemp"`
	DayTemp     string `json:"daytemp"`
	Thereshold  string `json:"thereshold"`
}

const min = -10
const max = 40

func GetTemp(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, dbGetTemp()); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
func (rd *Temp) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}
func dbGetTemp() *Temp {
	var t Temp
	//t.CurrentTemp = CurrentTemp //"20.00"
	t.CurrentTemp = fmt.Sprintf("%f", min+rand.Float64()*(max-min))
	t.DayTemp = DayTemp       //"23.00"
	t.NightTemp = NightTemp   //"18.00"
	t.Thereshold = Thereshold //"0.20"
	return &t
	//return nil, errors.New("user not found.")
}

/**-----------------------------------------------------------------------------------
 * get time
 * ========
 * $ curl http://bangkokguy.ddns.net/rest/v1/time // {"day":"06:00","night":"22:00"}
 *------------------------------------------------------------------------------------*/
type Times struct {
	Day   string `json:"day"`
	Night string `json:"night"`
}

func GetTime(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, dbGetTime()); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
func (rd *Times) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}
func dbGetTime() *Times {
	var t Times
	t.Day = Day
	t.Night = Night
	return &t
}

/**-----------------------------------------------------------------------------------
 * get mode
 * ========
 * $ curl http://bangkokguy.ddns.net/rest/v1/mode // {"mode":{"night|day" "auto|manual"},"heating":{"off":"manual|auto"}}
 *------------------------------------------------------------------------------------*/

type Modes struct {
	Mode    [2]string `json:["night", "day"] ["auto", "manual"]`
	Heating [2]string `json:["on", "off] ["auto", "manual"]`
}
type ModesIn struct {
	Mode    string `json:"mode"`
	Heating string `json:"heting"`
}

func GetMode(w http.ResponseWriter, r *http.Request) {
	if err := render.Render(w, r, dbGetMode()); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}
func (rd *Modes) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	return nil
}
func dbGetMode() *Modes {
	var m Modes
	m.Mode = Mode
	m.Heating = Heating
	return &m
}

/**-----------------------------------------------------------------------------------
 * put time
 * ========
 *
 *------------------------------------------------------------------------------------*/
func UpdateTime(w http.ResponseWriter, r *http.Request) {
	var time *Times

	data := &Times{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	time = data
	dbUpdateTime(time)

	render.Render(w, r, dbGetTime())
}
func (a *Times) Bind(r *http.Request) error {
	if a == nil {
		return errors.New("missing required Article fields")
	}
	a.Day = strings.ToLower(a.Day) // as an example, we down-case
	return nil
}
func dbUpdateTime(time *Times) (*Times, error) {
	Day = time.Day
	Night = time.Night
	return time, nil
}

/**-----------------------------------------------------------------------------------
* put temp
* ========
* $ curl -X PUT -d '{"daytemp":"24.00","nighttemp":"18.00", "thereshold":"0.20"}'
		http://bangkokguy.ddns.net/rest/v1/temp
*------------------------------------------------------------------------------------*/
func UpdateTemp(w http.ResponseWriter, r *http.Request) {
	var temp *Temp

	data := &Temp{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	temp = data
	println(temp.CurrentTemp + temp.DayTemp + temp.NightTemp + temp.Thereshold)
	dbUpdateTemp(temp)

	render.Render(w, r, dbGetTemp())
}
func (a *Temp) Bind(r *http.Request) error {
	if a == nil {
		return errors.New("missing required Temp fields")
	}
	//a.Day = strings.ToLower(a.Day) // as an example, we down-case
	return nil
}
func dbUpdateTemp(temp *Temp) (*Temp, error) {
	DayTemp = temp.DayTemp
	NightTemp = temp.NightTemp
	Thereshold = temp.Thereshold
	return temp, nil
}

/**-----------------------------------------------------------------------------------
 * put mode
 * ========
 * $ curl -X PUT -d '{"mode":"night|day|auto","heating":"on|off|auto"}' http://bangkokguy.ddns.net/rest/v1/mode
 *------------------------------------------------------------------------------------*/
func UpdateMode(w http.ResponseWriter, r *http.Request) {
	var mode *ModesIn

	data := &ModesIn{}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	mode = data
	dbUpdateMode(mode)

	render.Render(w, r, dbGetMode())
}

func (a *ModesIn) Bind(r *http.Request) error {
	if a == nil {
		return errors.New("missing required Mode fields")
	}
	return nil
}
func dbUpdateMode(mode *ModesIn) (*ModesIn, error) {
	Heating[1] = mode.Heating
	Heating[0] = CurrentStateOfHeating
	Mode[1] = mode.Mode
	Mode[0] = CurrentStateOfMode

	//var Mode [2]string = [2]string{"night", "auto"}
	//var Heating [2]string = [2]string{"off", "auto"}

	return mode, nil
}

/*------------------------------------------------------------------------------------*/
// GetArticle returns the specific Article. You'll notice it just
// fetches the Article right off the context, as its understood that
// if we made it this far, the Article must be on the context. In case
// its not due to a bug, then it will panic, and our Recoverer will save us.
func GetArticle(w http.ResponseWriter, r *http.Request) {
	// Assume if we've reach this far, we can access the article
	// context because this handler is a child of the ArticleCtx
	// middleware. The worst case, the recoverer middleware will save us.
	article := r.Context().Value("article").(*Article)

	if err := render.Render(w, r, NewArticleResponse(article)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// UpdateArticle updates an existing Article in our persistent store.
func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	println("UpdateArticle")
	article := r.Context().Value("article").(*Article)

	data := &ArticleRequest{Article: article}
	if err := render.Bind(r, data); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}
	article = data.Article
	dbUpdateArticle(article.ID, article)

	render.Render(w, r, NewArticleResponse(article))
}

// DeleteArticle removes an existing Article from our persistent store.
func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	var err error

	// Assume if we've reach this far, we can access the article
	// context because this handler is a child of the ArticleCtx
	// middleware. The worst case, the recoverer middleware will save us.
	article := r.Context().Value("article").(*Article)

	article, err = dbRemoveArticle(article.ID)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, NewArticleResponse(article))
}

// A completely separate router for administrator routes
func adminRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(AdminOnly)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("admin: index"))
	})
	r.Get("/accounts", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("admin: list accounts.."))
	})
	r.Get("/users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("admin: view user id %v", chi.URLParam(r, "userId"))))
	})
	return r
}

// AdminOnly middleware restricts access to just administrators.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value("acl.admin").(bool)
		if !ok || !isAdmin {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// paginate is a stub, but very possible to implement middleware logic
// to handle the request params for handling a paginated request.
func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		println("paginate inside " + r.URL.Host + " the end")
		next.ServeHTTP(w, r)
	})
}

// This is entirely optional, but I wanted to demonstrate how you could easily
// add your own logic to the render.Respond method.
func init() {
	render.Respond = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		if err, ok := v.(error); ok {

			// We set a default error status response code if one hasn't been set.
			if _, ok := r.Context().Value(render.StatusCtxKey).(int); !ok {
				w.WriteHeader(400)
			}

			// We log the error
			fmt.Printf("Logging err: %s\n", err.Error())

			// We change the response to not reveal the actual error message,
			// instead we can transform the message something more friendly or mapped
			// to some code / language, etc.
			render.DefaultResponder(w, r, render.M{"status": "error"})
			return
		}

		render.DefaultResponder(w, r, v)
	}
}

//--
// Request and Response payloads for the REST api.
//
// The payloads embed the data model objects an
//
// In a real-world project, it would make sense to put these payloads
// in another file, or another sub-package.
//--

type UserPayload struct {
	*User
	Role string `json:"role"`
}

func NewUserPayloadResponse(user *User) *UserPayload {
	return &UserPayload{User: user}
}

// Bind on UserPayload will run after the unmarshalling is complete, its
// a good time to focus some post-processing after a decoding.
func (u *UserPayload) Bind(r *http.Request) error {
	return nil
}

func (u *UserPayload) Render(w http.ResponseWriter, r *http.Request) error {
	u.Role = "collaborator"
	return nil
}

// ArticleRequest is the request payload for Article data model.
//
// NOTE: It's good practice to have well defined request and response payloads
// so you can manage the specific inputs and outputs for clients, and also gives
// you the opportunity to transform data on input or output, for example
// on request, we'd like to protect certain fields and on output perhaps
// we'd like to include a computed field based on other values that aren't
// in the data model. Also, check out this awesome blog post on struct composition:
// http://attilaolah.eu/2014/09/10/json-and-struct-composition-in-go/
type ArticleRequest struct {
	*Article

	User *UserPayload `json:"user,omitempty"`

	ProtectedID string `json:"id"` // override 'id' json to have more control
}

func (a *ArticleRequest) Bind(r *http.Request) error {
	// a.Article is nil if no Article fields are sent in the request. Return an
	// error to avoid a nil pointer dereference.
	if a.Article == nil {
		return errors.New("missing required Article fields")
	}

	// a.User is nil if no Userpayload fields are sent in the request. In this app
	// this won't cause a panic, but checks in this Bind method may be required if
	// a.User or futher nested fields like a.User.Name are accessed elsewhere.

	// just a post-process after a decode..
	a.ProtectedID = ""                                 // unset the protected ID
	a.Article.Title = strings.ToLower(a.Article.Title) // as an example, we down-case
	return nil
}

// ArticleResponse is the response payload for the Article data model.
// See NOTE above in ArticleRequest as well.
//
// In the ArticleResponse object, first a Render() is called on itself,
// then the next field, and so on, all the way down the tree.
// Render is called in top-down order, like a http handler middleware chain.
type ArticleResponse struct {
	*Article

	User *UserPayload `json:"user,omitempty"`

	// We add an additional field to the response here.. such as this
	// elapsed computed property
	Elapsed int64 `json:"elapsed"`
}

func NewArticleResponse(article *Article) *ArticleResponse {
	resp := &ArticleResponse{Article: article}

	if resp.User == nil {
		if user, _ := dbGetUser(resp.UserID); user != nil {
			resp.User = NewUserPayloadResponse(user)
		}
	}

	return resp
}

func (rd *ArticleResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

func NewArticleListResponse(articles []*Article) []render.Renderer {
	list := []render.Renderer{}
	for _, article := range articles {
		list = append(list, NewArticleResponse(article))
	}
	return list
}

// NOTE: as a thought, the request and response payloads for an Article could be the
// same payload type, perhaps will do an example with it as well.
// type ArticlePayload struct {
//   *Article
// }

//--
// Error response payloads & renderers
//--

// ErrResponse renderer type for handling all sorts of errors.
//
// In the best case scenario, the excellent github.com/pkg/errors package
// helps reveal information on the error, setting it on Err, and in the Render()
// method, using it to set the application-specific error code in AppCode.
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status"`          // user-level status message
	AppCode    int64  `json:"code,omitempty"`  // application-specific error code
	ErrorText  string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrInvalidRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 400,
		StatusText:     "Invalid request.",
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}

var ErrNotFound = &ErrResponse{HTTPStatusCode: 404, StatusText: "Resource not found."}

//--
// Data model objects and persistence mocks:
//--

// User data model
type User struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Article data model. I suggest looking at https://upper.io for an easy
// and powerful data persistence adapter.
type Article struct {
	ID     string `json:"id"`
	UserID int64  `json:"user_id"` // the author
	Title  string `json:"title"`
	Slug   string `json:"slug"`
}

// Article fixture data
var articles = []*Article{
	{ID: "1", UserID: 100, Title: "Hi", Slug: "hi"},
	{ID: "2", UserID: 200, Title: "sup", Slug: "sup"},
	{ID: "3", UserID: 300, Title: "alo", Slug: "alo"},
	{ID: "4", UserID: 400, Title: "bonjour", Slug: "bonjour"},
	{ID: "5", UserID: 500, Title: "whats up", Slug: "whats-up"},
}

// User fixture data
var users = []*User{
	{ID: 100, Name: "Peter"},
	{ID: 200, Name: "Julia"},
}

func dbNewArticle(article *Article) (string, error) {
	article.ID = fmt.Sprintf("%d", rand.Intn(100)+10)
	articles = append(articles, article)
	return article.ID, nil
}

func dbGetArticle(id string) (*Article, error) {
	for _, a := range articles {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbGetArticleBySlug(slug string) (*Article, error) {
	for _, a := range articles {
		if a.Slug == slug {
			return a, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbUpdateArticle(id string, article *Article) (*Article, error) {
	for i, a := range articles {
		if a.ID == id {
			articles[i] = article
			return article, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbRemoveArticle(id string) (*Article, error) {
	for i, a := range articles {
		if a.ID == id {
			articles = append((articles)[:i], (articles)[i+1:]...)
			return a, nil
		}
	}
	return nil, errors.New("article not found")
}

func dbGetUser(id int64) (*User, error) {
	for _, u := range users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}
