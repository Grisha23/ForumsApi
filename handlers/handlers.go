package handlers

import (
	"ForumsApi/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)


const (
	DbUser     = "docker"
	DbPassword = "docker"
	DbName     = "docker"
	//DbUser     = "tpforumsapi"
	//DbPassword = "222"
	//DbName = "forums_func"
)

var db *sql.DB

func InitDb() (*sql.DB, error) {
	var err error
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		DbUser, DbPassword, DbName)
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}

	init, err := ioutil.ReadFile("./forum.sql")
	_, err = db.Exec(string(init))

	if err != nil {
		panic(err)
	}

	fmt.Println("You connected to your database.")

	return db, nil
}

func getUser(nickname string) (*models.User, error) {
	if nickname == "" {
		return nil, nil
	}

	row := db.QueryRow("SELECT about,email,fullname,nickname FROM users WHERE nickname=$1", nickname)

	user := models.User{}

	err := row.Scan(&user.About, &user.Email, &user.FullName, &user.NickName)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func sendError(errText string, statusCode int, w *http.ResponseWriter) ([]byte, error){
	e := new(models.Error)
	e.Message = errText
	resp, _ := json.Marshal(e)

	// Проверка err json

	(*w).Header().Set("content-type", "application/json")
	(*w).WriteHeader(statusCode)
	(*w).Write(resp)

	return resp, nil
}

func UserProfile(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	if r.Method == http.MethodGet{
		user, err := getUser(nickname)

		if err != nil {
			sendError("Can't find user with nickname " + nickname + "\n", 404, &w)
			return
		}

		resp, _ := json.Marshal(user)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userUpdate := models.User{}

	err = json.Unmarshal(body, &userUpdate)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//aboutAdditional := ""
	//fullnameAdditional := ""
	//emailAdditional := ""

	about := false
	fullname := false
	email := false

	//separator := ""

	if userUpdate.About != ""{
		about = true
		//aboutAdditional = "about='"+userUpdate.About+"'"
	}
	if userUpdate.FullName != ""{
		fullname = true
		//if about {
		//	separator = ","
		//}
		//fullnameAdditional = separator + "fullname='"+userUpdate.FullName+"'"
		//separator = ""
	}
	if userUpdate.Email != ""{
		email = true
		//if about || fullname {
		//	separator = ","
		//}
		//emailAdditional = separator+"email='"+userUpdate.Email+"'"
	}

	if !email && !fullname && !about {
		user, err := getUser(nickname)

		if err != nil {
			sendError("Can't find prifile with id " + nickname + "\n", 404, &w)
		}

		resp, _ := json.Marshal(user)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}
	var query string
	var row *sql.Row
	//query := "UPDATE users SET " + aboutAdditional + fullnameAdditional + emailAdditional + " WHERE nickname=$1 RETURNING " +
	//	"about,email,fullname,nickname"
	//
	//row := db.QueryRow(query,nickname)

	if about && fullname && email {
		query = "UPDATE users SET about=$1, fullname=$2, email=$3 WHERE nickname=$4 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.About, userUpdate.FullName, userUpdate.Email, nickname)
	} else if about && fullname && !email {
		query = "UPDATE users SET about=$1, fullname=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.About, userUpdate.FullName, nickname)
	} else if about && !fullname && email {
		query = "UPDATE users SET about=$1, email=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.About, userUpdate.Email, nickname)
	} else if about && !fullname && !email {
		query = "UPDATE users SET about=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.About, nickname)
	} else if !about && fullname && email {
		query = "UPDATE users SET fullname=$1, email=$2 WHERE nickname=$3 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.FullName, userUpdate.Email, nickname)
	} else if !about && fullname && !email {
		query = "UPDATE users SET fullname=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.FullName, nickname)
	} else if !about && !fullname && email {
		query = "UPDATE users SET email=$1 WHERE nickname=$2 RETURNING about,email,fullname,nickname"
		row = db.QueryRow(query, userUpdate.Email, nickname)
	}

	err = row.Scan(&userUpdate.About, &userUpdate.Email, &userUpdate.FullName, &userUpdate.NickName)

	if err != nil {
		fmt.Println(err.Error())
		if err == sql.ErrNoRows {
			sendError("Can't find prifile with id " + nickname + "\n", 404, &w)
			return
		}

		errorName := err.(*pq.Error).Code.Name()

		if errorName == "unique_violation"{
			sendError("Can't change prifile with id " + nickname + "\n", 409, &w)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, _ := json.Marshal(userUpdate)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)

	return
}

/*
curl -i --header "Content-Type: application/json" --request GET http://127.0.0.1:8080/user/grisha23/details
curl -i --header "Content-Type: application/json" --request POST --data '{"about":"text about user" , "email": "myemail@ddf.ru", "fullname": "Grigory"}' http://127.0.0.1:8080/user/grisha23/profile

*/

func UserCreate(w http.ResponseWriter, r *http.Request)  {
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}

	user := models.User{}
	err := json.Unmarshal(body, &user)
	user.NickName = nickname

	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}

	if user.NickName == "" || user.About == "" || user.Email == "" || user.FullName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t, _ := db.Begin()

	defer t.Rollback()

	t.Exec("SET LOCAL synchronous_commit TO OFF")

	query := "INSERT INTO users(about, email, fullname, nickname) VALUES ($1,$2,$3,$4) RETURNING *"

	err = t.QueryRow(query, user.About, user.Email, user.FullName, user.NickName).Scan(&user.About,
		&user.Email, &user.FullName, &user.NickName)

	if err != nil {
		errorName := err.(*pq.Error).Code.Name()

		if errorName == "unique_violation"{
			users := make([]models.User, 0)

			rows, err := db.Query("SELECT * FROM users WHERE nickname=$1 OR email=$2", user.NickName, user.Email)

			if err != nil{
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			for rows.Next() {
				usr := models.User{}

				err := rows.Scan(&usr.About, &usr.Email, &usr.FullName, &usr.NickName)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				users = append(users, usr)
			}

			resp, _ := json.Marshal(users)
			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusConflict)
			w.Write(resp)

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return

	}


	resp, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Commit()

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
	return

}

/*
curl -i --header "Content-Type: application/json" --request POST --data '{"about":"text about user" , "email": "myemail@ddf.ru", "fullname": "Grigory"}' http://127.0.0.1:8080/user/grisha23/create

*/
func ThreadVote(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodPost{
		return
	}

	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	vote := models.Vote{}

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &vote)

	//thrId, err := strconv.Atoi(slugOrId)
	//if err != nil {
	//	id = slugOrId
	//} else {
	//	id = strconv.Itoa(thrId)
	//}

	//thr, err := getThread(slugOrId)
	//
	//if err != nil {
	//	//errorName := err.(*pq.Error).Code.Name()
	//	//if errorName ==
	//	sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
	//	return
	//}
	//oldVote := models.Vote{}

	//errGetVote := db.QueryRow("SELECT voice FROM votes WHERE nickname=$1 AND thread=$2", vote.Nickname, thr.Id).Scan(&oldVote.Voice)


	//var id string
	thrId, err := strconv.Atoi(slugOrId)
	//var row *sql.Row

	if err != nil {
		_,err = db.Exec("INSERT INTO votes(nickname, voice, thread) VALUES ($1,$2, (SELECT id FROM threads WHERE slug=$3)) " +
			"ON CONFLICT (nickname, thread) DO " +
			"UPDATE SET voice=$2",
			vote.Nickname, vote.Voice, slugOrId)
	} else {
		_,err = db.Exec("INSERT INTO votes(nickname, voice, thread) VALUES ($1,$2,$3) " +
			"ON CONFLICT (nickname, thread) DO " +
			"UPDATE SET voice=$2",
			vote.Nickname, vote.Voice, thrId)
	}
	// Прблема: id может быть slug, а может id. В случае slug нужно тащить отдельным запросом id треда.

	//newVote := models.Vote{}


	//err = row.Scan(&oldVote.Nickname, &oldVote.Voice, &oldVote.Thread)

	if err != nil {
		fmt.Println(err.Error())
		if err.(*pq.Error).Code.Name() == "foreign_key_violation" {
			sendError("Can't find user with id " + slugOrId + "\n", 404, &w)
			return
		}
	}

	thr, err := getThread(slugOrId)

	if err != nil {
		//errorName := err.(*pq.Error).Code.Name()
		//if errorName ==
		sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
		return
	}

	resp, _ := json.Marshal(thr)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)
	return

}

/*
curl -i --header "Content-Type: application/json" --request POST --data '{"nickname": "Grisha23", "voice": -1}' http://127.0.0.1:8080/thread/19/vote

*/

func ThreadPosts(w http.ResponseWriter, r *http.Request) {
	//if r.Method != http.MethodGet {
	//	return
	//}

	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	thr, err := getThread(slugOrId)

	if err != nil {
		sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
		return
	}

	limitVal := r.URL.Query().Get("limit")
	sinceVal := r.URL.Query().Get("since")
	descVal := r.URL.Query().Get("desc")
	sortVal := r.URL.Query().Get("sort")

	var since= false
	var desc= false
	var limit= false

	if limitVal == "" {
		limitVal = " ALL"
	} else {
		limit = true
	}
	if sinceVal != "" {
		since = true
	}
	if descVal == "true" {
		desc = true
	}
	if sortVal != "flat" && sortVal != "tree" && sortVal != "parent_tree" {
		sortVal = "flat"
	}

	var rows *sql.Rows

	if sortVal == "flat" {
		if desc {

			if since {

				rows, err = db.Query("SELECT * FROM posts WHERE thread = $1 AND id < $3 ORDER BY created DESC, id DESC LIMIT $2", thr.Id, limitVal, sinceVal)

			} else {

				rows, err = db.Query("SELECT * FROM posts WHERE thread = $1 ORDER BY id DESC LIMIT $2", thr.Id, limitVal)

			}

		} else {

			if since {

				rows, err = db.Query("SELECT * FROM posts WHERE thread = $1 AND id > $3 ORDER BY id ASC LIMIT $2", thr.Id, limitVal, sinceVal)

			} else {
				query := "SELECT * FROM posts WHERE thread = $1 ORDER BY id ASC LIMIT " + limitVal
				rows, err = db.Query(query, thr.Id)

			}

		}
	} else if sortVal == "tree" {
		sinceAddition := ""
		sortAddition := ""
		limitAddition := ""
		if desc == true {
			sortAddition = " order by path[0],path DESC "
			if since != false {
				sinceAddition = " where path < (select path from post_tree where id = " + sinceVal + " ) "
			}
		} else {
			sortAddition = " order by path[0],path ASC"
			if since != false {
				sinceAddition = " where path > (select path from post_tree where id = " + sinceVal + " ) "
			}
		}

		if limit != false {
			limitAddition = "limit " + limitVal
		}
		query := "WITH recursive post_tree(id,path) as(select p.id,array_append('{}'::bigint[], id) as arr_id from posts p " +
			"where p.parent = 0 and p.thread=$1 union all " +
			"select p.id, array_append(path, p.id) from posts p join post_tree pt on p.parent = pt.id) " +
			"select p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread from post_tree pt join " +
			"posts p on p.id = pt.id " + sinceAddition + " " + sortAddition + " " + limitAddition
		rows, err = db.Query(query, thr.Id)
	} else if sortVal == "parent_tree" {
		descflag := ""
		sinceAddition := ""
		sortAddition := ""
		limitAddition := ""
		if desc == true {
			descflag = " desc "
			sortAddition = "order by path[1] DESC,path"
			if since != false {
				sinceAddition = " where path[1] < (select path[1] from post_tree where id = " + sinceVal + " ) "
			}
		} else {
			descflag = " asc "
			sortAddition = " order by path[1] ,path ASC"
			if since != false {
				sinceAddition = " where path[1] > (select path[1] from post_tree where id = " + sinceVal + " ) "
			}
		}

		if limit != false {
			limitAddition = " where r <= " + limitVal
		}

		query :="select p.author,p.created,p.forum,p.id,p.isedited,p.message,p.parent,p.thread from (with recursive post_tree(id,path) as( "+
			"select p.id,array_append('{}'::bigint[], p.id) as arr_id "+
			"from posts p "+
			"where p.parent = 0 and p.thread = $1 "+

			"union all "+

			"select p.id, array_append(path, p.id) from posts p "+
			"join post_tree pt on p.parent = pt.id "+
			") "+
			"select post_tree.id as id,path, dense_rank() over (order by path[1] " + descflag + " ) as " +
			"r from post_tree " + sinceAddition + " ) as pt join posts p on p.id = pt.id " + limitAddition + " " + sortAddition + ";"
		rows, err = db.Query(query, thr.Id)
	}

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer rows.Close()
	posts := make([]models.Post, 0)
	var i = 0
	for rows.Next(){
		i++
		post := models.Post{}

		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			fmt.Println(err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//err = arr.Scan(&post.Childs)
		if err != nil {
			fmt.Println(err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		posts = append(posts, post)

	}
	w.Header().Set("content-type", "application/json")

	resp, _ := json.Marshal(posts)

	w.Write(resp)

	return
}

func ThreadDetails(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	if r.Method == http.MethodPost{

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		thr := models.Thread{}

		err = json.Unmarshal(body, &thr)

		if thr.Title == "" && thr.Message == ""{
			existThr, err := getThread(slugOrId)

			if err != nil {
				sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
				return
			}

			resp, err := json.Marshal(existThr)
			w.Header().Set("content-type", "application/json")

			w.Write(resp)
			return
		}

		var many = " "

		var messageAddition = ""

		var titleAddition = ""

		if thr.Message != "" {
			messageAddition = " message='" + thr.Message + "' "
		}

		if thr.Title != "" {
			titleAddition = " title='" + thr.Title + "' "
		}

		if thr.Title != "" && thr.Message != "" {
			many = ","
		}

		var row *sql.Row

		thrId, err := strconv.Atoi(slugOrId)

		var idenAdditional string

		if err != nil {
			idenAdditional = "slug='" + slugOrId + "' "

		} else {
			idenAdditional = "id=" + strconv.Itoa(thrId)
		}

		query := "UPDATE threads SET " + messageAddition + many + titleAddition + " WHERE " + idenAdditional + " RETURNING *"
		row = db.QueryRow(query)

		err = row.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &thr.Slug, &thr.Title, &thr.Votes)

		if err != nil {
			sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
			return
		}

		resp, err := json.Marshal(thr)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}

	thr, err := getThread(slugOrId)

	if err != nil {
		sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
		return
	}

	resp, _ := json.Marshal(thr)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)

	return
}

/*
curl -i --header "Content-Type: application/json" --request GET http://127.0.0.1:8080/thread/2/details
curl -i --header "Content-Type: application/json" --request POST --data '{"message": "Message test method change thread", "title": "Title change"}' http://127.0.0.1:8080/thread/14/details

*/

func getThread(slug string) (*models.Thread, error) {
	thrId, err := strconv.Atoi(slug)
	var row *sql.Row
	if err != nil {
		row = db.QueryRow("SELECT * FROM threads WHERE slug=$1;", slug)
	} else {
		row = db.QueryRow("SELECT * FROM threads WHERE id=$1;", thrId)
	}

	var sqlSlug sql.NullString

	thr := new(models.Thread)
	err = row.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum, &thr.Message, &sqlSlug, &thr.Title, &thr.Votes)

	if !sqlSlug.Valid {
		thr.Slug = ""
	} else {
		thr.Slug = sqlSlug.String
	}

	if err != nil {
		//fmt.Println("Error: ", err.Error())
		return nil, err
	}

	return thr, nil
}

func PostCreate(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	posts := make([]models.Post, 0)

	err = json.Unmarshal(body, &posts)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := make([]models.Post,0)

	//thrId, err := strconv.Atoi(slugOrId)
	//var id string
	//if err != nil {
	//	id = strconv.Itoa(thrId)
	//} else {
	//	id = slugOrId
	//}
	t, err := db.Begin()
	_, err = t.Exec("SET LOCAL synchronous_commit = OFF")

	thr, err := getThread(slugOrId)

	if err != nil{
		sendError("Can't find thread with id " + slugOrId + "\n", 404, &w)
		return
	}


	defer t.Rollback()

	var firstCreated time.Time
	var count = 0
	//var err error
	for _, p := range posts{

		newPost := models.Post{}
		if count == 0 { // Для того, чтобы все последующие добавления постов происхдили с той же датой и временем.
			row := t.QueryRow("INSERT INTO posts(author, forum, message, parent, thread) VALUES ($1,$2,$3,$4,$5) RETURNING *",
				p.Author, thr.Forum, p.Message, p.Parent, thr.Id)
			err = row.Scan(&newPost.Author, &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
				&newPost.Parent, &newPost.Thread)

			firstCreated = newPost.Created
		} else {
			row := t.QueryRow("INSERT INTO posts(author, forum, message, parent, thread, created) VALUES ($1,$2,$3,$4,$5,$6) RETURNING *",
				p.Author, thr.Forum, p.Message, p.Parent, thr.Id, firstCreated)
			err = row.Scan(&newPost.Author, &newPost.Created, &newPost.Forum, &newPost.Id, &newPost.IsEdited, &newPost.Message,
				&newPost.Parent, &newPost.Thread)
		}


		if err != nil {
			errorName := err.(*pq.Error).Code.Name()
			if err.Error() == "pq: Parent post exc" {
				sendError("Parent post was created in another thread \n", 409, &w)
				return
			}

			if errorName == "foreign_key_violation" {
				sendError("Can't find parent post \n", 404, &w)
				return
			}

			sendError("Can't find parent post \n", 404, &w)
			return
		}

		data = append(data, newPost)

		count++

	}

	resp, err := json.Marshal(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	t.Commit()

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)

	defer r.Body.Close()

	return
}

/*
curl -i --header "Content-Type: application/json" --request POST --data '[{"author":"Grisha23", "message":"NEW", "parent":0},{"author":"Grisha23", "message":"NEW", "parent":2}, {"author":"Grisha23", "message":"NEW NEW NEW NEW !!!!", "parent":0}]' http://127.0.0.1:8080/thread/14/create

*/


func ServiceStatus(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodGet{
		return
	}

	row := db.QueryRow("SELECT t1.cnt c1, t2.cnt c2, t3.cnt c3, t4.cnt c4 FROM (SELECT count(*) cnt FROM users) t1, (SELECT COUNT(*) cnt FROM forums) t2, (SELECT COUNT(*) cnt FROM posts) t3, (SELECT COUNT(*) cnt FROM threads) t4;")
//	row := db.QueryRow("SELECT reltuples::bigint AS estimate FROM pg_class WHERE oid='users'::regclass")

	status := models.Status{}

	err := row.Scan(&status.User, &status.Forum, &status.Post, &status.Thread)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, _ := json.Marshal(status)
	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusOK)

	w.Write(resp)

	return
}

func ServiceClear(w http.ResponseWriter, r *http.Request)  {
	if r.Method != http.MethodPost{
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	db.Query("TRUNCATE TABLE votes, users, posts, threads, forums")

	w.WriteHeader(http.StatusOK)

	return
}

func PostDetails(w http.ResponseWriter, r *http.Request){

	vars := mux.Vars(r)
	id := vars["id"]

	related := r.URL.Query().Get("related")

	if r.Method == http.MethodPost {
		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer r.Body.Close()

		post := new(models.Post)

		err = json.Unmarshal(body, post)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err != nil{
			sendError( "Can't find post with id " + id + "\n", 404, &w)
			return
		}

		if post.Message == "" {
			row := db.QueryRow("SELECT * FROM posts WHERE id=$1", id)

			err = row.Scan(&post.Author,&post.Created,&post.Forum,&post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

			if err != nil{
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			resp, _ := json.Marshal(post)
			w.Header().Set("content-type", "application/json")

			w.Write(resp)

			return
		}
		row := db.QueryRow("UPDATE posts SET message=$1, isedited=true WHERE id=$2 RETURNING *", post.Message, id)
		err = row.Scan(&post.Author,&post.Created,&post.Forum,&post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

		if err != nil {
			sendError("Can't find post with id "+id+"\n", 404, &w)
			return
		}

		resp, _ := json.Marshal(post)
		w.Header().Set("content-type", "application/json")

		w.Write(resp)

		return
	}

	row := db.QueryRow("SELECT * FROM posts WHERE id=$1;", id)

	post := models.Post{}

	err := row.Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)

	if err != nil {
		sendError( "Can't find post with id " + id + "\n", 404, &w)
		return
	}

	postDetail := models.PostDetail{}

	postDetail.Post = &post

	var relatedObj []string
	pathItems := strings.Split(related, ",")
	for index := range pathItems  {
		item := pathItems[index]
		relatedObj = append(relatedObj, item)
	}
	for index := range relatedObj {
		if relatedObj[index] == "user" {
			author, _ := getUser(post.Author)
			postDetail.Author = author
		}
		if relatedObj[index] == "thread" {
			thread, err := getThread(strconv.Itoa(int(post.Thread)))

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			postDetail.Thread = thread
		}
		if relatedObj[index] == "forum" {
			forum, _ := getForum(post.Forum)
			postDetail.Forum = forum
		}
	}

	resp, _ := json.Marshal(postDetail)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)
	return
}

/*
curl -i --header "Content-Type: application/json" --request GET http://127.0.0.1:8080/post/2/details

curl -i --header "Content-Type: application/json" --request POST --data '{"message":"NEW NEW NEW"}' http://127.0.0.1:8080/post/2/details

*/

func ForumUsers(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		return
	}

	limitVal := r.URL.Query().Get("limit")
	sinceVal := r.URL.Query().Get("since")
	descVal := r.URL.Query().Get("desc")

	var limit = false
	var since = false
	var desc = false

	if limitVal != "" {
		limit = true
	}
	if sinceVal != "" {
		since = true
	}
	if descVal == "true" {
		desc = true
	}

	var rows *sql.Rows
	var err error

	vars := mux.Vars(r)
	slug := vars["slug"]

	frm, _ := getForum(slug)

	if frm == nil {
		sendError("Can't find forum with slug " + slug + "\n", 404, &w)
		return
	}
	//Исправить на JOIN
	if !limit && !since && !desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 GROUP BY author) ORDER BY nickname ASC;", slug)
	} else if !limit && !since && desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 GROUP BY author) ORDER BY nickname DESC;", slug)
	} else if !limit && since && !desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 AND author>$2 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 AND author>$2 GROUP BY author) AND nickname>$2 ORDER BY nickname ASC;", slug, sinceVal)
	} else if !limit && since && desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 AND author<$2 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 AND author<$2 GROUP BY author) AND nickname<$2 ORDER BY nickname DESC;", slug, sinceVal)
	} else if limit && !since && !desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 GROUP BY author) ORDER BY nickname ASC LIMIT $2;", slug, limitVal)
	} else if limit && !since && desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 GROUP BY author) ORDER BY nickname DESC LIMIT $2;", slug, limitVal)
	} else if limit && since && !desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 AND author>$2 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 AND author>$2 GROUP BY author) AND nickname>$2 ORDER BY nickname ASC LIMIT $3;", slug, sinceVal, limitVal)
	} else if limit && since && desc {
		rows, err = db.Query("SELECT * FROM users WHERE nickname IN (SELECT author FROM threads WHERE forum=$1 AND author<$2 GROUP BY author) OR nickname IN (SELECT author FROM posts WHERE forum=$1 AND author<$2 GROUP BY author) AND nickname<$2 ORDER BY nickname DESC LIMIT $3;", slug, sinceVal, limitVal)
	}

	if err != nil {
		fmt.Println(err.Error())
		sendError( "Can't find forum with slug " + slug + "\n", 404, &w)
		return
	}

	defer rows.Close()

	users := make([]models.User, 0)

	for rows.Next() {
		usr := models.User{}

		err := rows.Scan(&usr.About, &usr.Email, &usr.FullName, &usr.NickName)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		users = append(users, usr)
	}

	resp, _ := json.Marshal(users)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)

	return
}

/*
curl -i --header "Content-Type: application/json" --request GET http://127.0.0.1:8080/forum/stories-about/users?since=z

*/

func ForumThreads(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		return
	}

	limitVal := r.URL.Query().Get("limit")
	sinceVal := r.URL.Query().Get("since")
	descVal := r.URL.Query().Get("desc")

	var limit = false
	var since = false
	var desc = false

	if limitVal != "" {
		limit = true
	}
	if sinceVal != "" {
		since = true
	}
	if descVal == "true" {
		desc = true
	}

	vars := mux.Vars(r)
	slug := vars["slug"]
	frm, _ := getForum(slug)  // Исправить
	if frm == nil {
		sendError("Can't find forum with slug " + slug + "\n", 404, &w)
		return
	}

	var rows *sql.Rows
	var err error

	//query := " ...EXISTS SELECT 1 FROM forums WHERE slug=$1..."

	if limit && !since && !desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created LIMIT $2;", slug, limitVal)
	} else if since && !limit && !desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created;", slug, sinceVal)
	} else if limit && since && !desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", slug, sinceVal, limitVal)
	} else if limit && !since && desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created DESC LIMIT $2;", slug, limitVal)
	} else if since && !limit && desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC;", slug, sinceVal)
	} else if limit && since && desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created <= $2 ORDER BY created DESC LIMIT $3;", slug, sinceVal, limitVal)
	} else if limit && since && !desc{
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created LIMIT $3;", slug, sinceVal, limitVal)
	} else if !limit && !since && !desc {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created;", slug)
	} else {
		rows, err = db.Query("SELECT * FROM threads WHERE forum = $1 ORDER BY created;", slug)
	}

	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	thrs := make([]models.Thread, 0)
	var nullSlug sql.NullString
	for rows.Next() {
		thr := models.Thread{}
		err := rows.Scan(&thr.Id, &thr.Author, &thr.Created, &thr.Forum,  &thr.Message, &nullSlug, &thr.Title, &thr.Votes)

		if nullSlug.Valid {
			thr.Slug = nullSlug.String
		} else {
			thr.Slug = ""
		}

		if err != nil {
			fmt.Println(err.Error())

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		thrs = append(thrs, thr)
	}

	resp, _ := json.Marshal(thrs)
	w.Header().Set("content-type", "application/json")

	w.Write(resp)

	return
}

func getForum(slugOrId string) (*models.Forum,error) {
	forum := models.Forum{}
	err := db.QueryRow("SELECT * FROM forums WHERE slug=$1", slugOrId).Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

	if err != nil {
		return nil, err
	}

	return &forum, nil
}

/*
FORUM THREADS

curl -i --header "Content-Type: application/json" --request GET http://127.0.0.1:8080/forum/stories-about/threads

*/

func ForumDetails(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		return
	}
	vars := mux.Vars(r)
	slug := vars["slug"]
	frm, err := getForum(slug)

	if err != nil { // Значит строка пустая.
		sendError( "Can't find user with slug " + slug + "\n", 404, &w)
		return
	}

	resp, err := json.Marshal(frm)

	if err != nil {
		return
	}
	w.Header().Set("content-type", "application/json")

	w.Write(resp)
	return
}

/*
FORUM DETAILS
curl -i --header "Content-Type: application/json" --request GET http://127.0.0.1:8080/forum/stories-about/details
*/

func ThreadCreate(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if readErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	thr := models.Thread{}

	json.Unmarshal(body, &thr)

	params := mux.Vars(r)
	slug := params["slug"]

	var err error
	var row *sql.Row
	if thr.Slug == "" {
		row = db.QueryRow("INSERT INTO threads(author, created, forum, message, title) VALUES ($1, $2, " +
			"(SELECT slug FROM forums WHERE slug=$3), $4, $5) RETURNING *", thr.Author, thr.Created, slug,
			thr.Message, thr.Title)
	} else {
		row = db.QueryRow("INSERT INTO threads(author, created, forum, message, title, slug) VALUES ($1, $2, " +
			"(SELECT slug FROM forums WHERE slug=$3), $4, $5, $6) RETURNING *", thr.Author, thr.Created, slug,
			thr.Message, thr.Title, thr.Slug)
	}

	newThr := models.Thread{}
	var sqlSlug sql.NullString
	err = row.Scan(&newThr.Id, &newThr.Author, &newThr.Created, &newThr.Forum, &newThr.Message, &sqlSlug, &newThr.Title, &newThr.Votes)

	if err != nil {
		errorName := err.(*pq.Error).Code.Name()

		if errorName == "foreign_key_violation" || errorName == "not_null_violation"{
			sendError( "Can't find user or forum \n", 404, &w)
			return
		}

		if errorName == "unique_violation"{
			existThr, _ := getThread(thr.Slug)

			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusConflict)
			resp, _ := json.Marshal(existThr)

			w.Write(resp)
			return
		}
		return
	}

	if !sqlSlug.Valid {
		newThr.Slug = ""
	} else {
		newThr.Slug = sqlSlug.String
	}

	//var forumSlug string
	//db.QueryRow("SELECT slug FROM forums WHERE slug=$1", thr.Forum).Scan(&forumSlug) // Неэффективно
	//newThr.Forum = forumSlug

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, _:= json.Marshal(newThr)
	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)
	w.Write(resp)

	return
}

/*
CREATE THREAD
curl -i --header "Content-Type: application/json" --request POST --data '{"author":"Grisha23","message":"DWjn waonda owadndn wa awn n3342", "title": "Thread1"}'   http://127.0.0.1:8080/forum/stories-about/create
*/

func ForumCreate(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodGet {
		return
	}

	body, err := ioutil.ReadAll(r.Body)

	t, _ := db.Begin()
	defer t.Rollback()

	t.Exec("SET LOCAL synchronous_commit TO OFF")

	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	forum := new(models.Forum)
	err = json.Unmarshal(body, forum)

	existUser, _ := getUser(forum.User)

	if existUser == nil {
		sendError( "Can't find user with name " + forum.User + "\n", 404, &w)
		return
	}

	row := t.QueryRow("INSERT INTO forums(slug, title, author) VALUES ($1, $2, $3) RETURNING *", forum.Slug, forum.Title, existUser.NickName)

	err = row.Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)

	if err != nil {
		errorName := err.(*pq.Error).Code.Name()
		if errorName == "foreign_key_violation" {
			sendError( "Can't find user with name " + forum.User + "\n", 404, &w)
			return
		}
		if errorName == "unique_violation" {
			row := db.QueryRow("SELECT * FROM forums WHERE slug=$1", forum.Slug)
			fr := models.Forum{}
			row.Scan(&fr.Posts, &fr.Slug, &fr.Threads, &fr.Title, &fr.User)
			w.Header().Set("content-type", "application/json")

			w.WriteHeader(http.StatusConflict)
			resp, _ := json.Marshal(fr)

			w.Write(resp)
			return
		}
	}

	t.Commit()

	resp, _ := json.Marshal(forum)

	w.Header().Set("content-type", "application/json")

	w.WriteHeader(http.StatusCreated)

	w.Write(resp)

	return
}

/*
CREATE FORUM
curl -i --header "Content-Type: application/json"   --request POST
--data '{"slug":"stori123es-eabout","title":"Stoewries about som12ewe3ething",
"user": "Gris21ha23"}'   http://127.0.0.1:8080/forum/create
*/

