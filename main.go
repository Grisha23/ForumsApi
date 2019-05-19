package main

import (
	"github.com/Grisha23/ForumsApi/handlers"
	// "ForumsApi/handlers"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func AccessLogMiddleware (mux *mux.Router,) http.HandlerFunc   {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()

		mux.ServeHTTP(w, r)

		sortVal := r.URL.Query().Get("sort")
		if sortVal != "" {
			fmt.Println("method ", r.Method, "; url", r.URL.Path,  " Sort: ", sortVal,
				"Time work: ", time.Since(begin))
		} else {
			fmt.Println("method ", r.Method, "; url", r.URL.Path,
				"Time work: ", time.Since(begin))

		}

		HitStat.With(prometheus.Labels{
			"url":    r.URL.Path,
			"method": r.Method,
			"code":   w.Header().Get("Status-Code"),
		}).Inc()


		rps.Add(1)
		
		// if error != nil {
		// 	fmt.Println("error here")
		// 	return
		// }
		// fmt.Println(cpu_info)

		//if sortVal != "" {
		//	fmt.Println("END method ", r.Method, " Sort: ", sortVal, "; url", r.URL.Path,
		//		"Time work: ", time.Since(begin))
		//} else {
		//	fmt.Println("END method ", r.Method, "; url", r.URL.Path,
		//		"Time work: ", time.Since(begin))
		//}



	})
}

var (
	HitStat = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "ForumsApi",
			Subsystem: "hit_stat",
			Name:      "HitStat",
			Help:      "Hit info.",
		},
		[]string{
			"url",
			"method",
			"code",
		},
	)

	rps =prometheus.NewCounter(
		prometheus.CounterOpts{
		  Name: "rps_total",
		})
	)
	

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(HitStat)
	prometheus.MustRegister(rps)

}


func main(){

	// The Handler function provides a default handler to expose metrics
	// via an HTTP server. "/metrics" is the usual endpoint for that.


	db, _ := handlers.InitDb()

	router := mux.NewRouter()

	http.Handle("/metrics", promhttp.Handler())

	router.HandleFunc("/api/forum/create", handlers.ForumCreate)
	router.HandleFunc(`/api/forum/{slug}/create`, handlers.ThreadCreate)
	router.HandleFunc(`/api/forum/{slug}/details`, handlers.ForumDetails) // +
	router.HandleFunc(`/api/forum/{slug}/threads`, handlers.ForumThreads) // - не оч
	router.HandleFunc(`/api/forum/{slug}/users`, handlers.ForumUsers) // +

	router.HandleFunc(`/api/post/{id}/details`, handlers.PostDetails) // +

	router.HandleFunc(`/api/service/clear`, handlers.ServiceClear)
	router.HandleFunc(`/api/service/status`, handlers.ServiceStatus) // -

	router.HandleFunc(`/api/thread/{slug_or_id}/create`, handlers.PostCreate)
	router.HandleFunc(`/api/thread/{slug_or_id}/details`, handlers.ThreadDetails) // +
	router.HandleFunc(`/api/thread/{slug_or_id}/posts`, handlers.ThreadPosts) // +
	router.HandleFunc(`/api/thread/{slug_or_id}/vote`, handlers.ThreadVote)

	router.HandleFunc(`/api/user/{nickname}/create`, handlers.UserCreate)
	router.HandleFunc(`/api/user/{nickname}/profile`, handlers.UserProfile)  // + быстро

	siteHandler := AccessLogMiddleware(router)

	http.Handle("/", siteHandler)
	http.ListenAndServe(":5000", nil)

	defer db.Close()

	return
}

