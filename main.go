package main

import (
	"ForumsApi/handlers"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func AccessLogMiddleware (mux *mux.Router,) http.HandlerFunc   {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		mux.ServeHTTP(w, r)

		fmt.Println("method", r.Method, "; url", r.URL.Path)

	})
}


func main(){
	db, _ := handlers.InitDb()

	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/api/forum/create", handlers.ForumCreate)
	router.HandleFunc(`/api/forum/{slug}/create`, handlers.ThreadCreate)
	router.HandleFunc(`/api/forum/{slug}/details`, handlers.ForumDetails)
	router.HandleFunc(`/api/forum/{slug}/threads`, handlers.ForumThreads)
	router.HandleFunc(`/api/forum/{slug}/users`, handlers.ForumUsers)

	router.HandleFunc(`/api/post/{id}/details`, handlers.PostDetails)

	router.HandleFunc(`/api/service/clear`, handlers.ServiceClear)
	router.HandleFunc(`/api/service/status`, handlers.ServiceStatus)

	router.HandleFunc(`/api/thread/{slug_or_id}/create`, handlers.PostCreate)
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.ThreadDetails)
	router.HandleFunc(`/api/thread/{slug_or_id}/posts`, handlers.ThreadPosts)
	router.HandleFunc(`/api/thread/{slug_or_id}/vote`, handlers.ThreadVote)

	router.HandleFunc(`/api/user/{nickname}/create`, handlers.UserCreate)
	router.HandleFunc(`/api/user/{nickname}/profile`, handlers.UserProfile)

	siteHandler := AccessLogMiddleware(router)

	http.Handle("/", router)
	http.ListenAndServe(":5000", siteHandler)
	return
}

