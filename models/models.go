package models

import "time"

type Error struct {
	Message string 			`json:"message"`
}

type Forum struct {
	Posts int64   			`json:"posts"`			// Кол-во сообщение в данном форуме
	Slug string  			`json:"slug"`			// Человеко понятный URL
	Threads int32 			`json:"threads"`		// Кол-во веток в данном форуме
	Title string  			`json:"title"`			// Название форума
	User string   			`json:"user"`			// Nickname создателя
}

type Post struct {
	Author string 			`json:"author"`			// Автор, написавший сообщение
	Created time.Time		`json:"created"`		// Дата создания сообщения на форуме
	Forum string 			`json:"forum"`			// Идентификатор форума
	Id int64 				`json:"id"`				// Идентификатор данного сообщения
	IsEdited bool 			`json:"isEdited"`		// Истина, если данное сообщение было изменено.
	Message string 			`json:"message"`		// Собственно сообщение форума.
	Parent int64 			`json:"parent"`			// Идентификатор родительского сообщения (0 - корневое сообщение обсуждения).
	Thread int32 			`json:"thread"`			// Идентификатор ветви (id) обсуждения данного сообещния.
}

type Status struct {
	Forum int32 			`json:"forum"` 			// Кол-во разделов в базе данных.
	Post int64 				`json:"post"`			// Кол-во сообщений в базе данных.
	Thread int32 			`json:"thread"`			// Кол-во веток обсуждения в базе данных.
	User int32 				`json:"user"`			// Кол-во пользователей в базе данных.
}

type Thread struct {
	Author string   		`json:"author"`			// Пользователь, создавший данную тему.
	Created time.Time 		`json:"created"` 		// Дата создания ветки на форуме.
	Forum string 			`json:"forum"` 			// Форум, в котором расположена данная ветка обсуждения.
	Id int32 				`json:"id"`				// Идентификатор ветки обсуждения.
	Message string 			`json:"message"`		// Описание ветки обсуждения.
	Slug string				`json:"slug"`			// Человекопонятный URL. В данной структуре slug опционален и не может быть числом.
	Title string 			`json:"title"`			// Заголовок ветки обсуждения.
	Votes int32 			`json:"votes"`			// Кол-во голосов непосредственно за данное сообщение форума.
}

type User struct {
	About string 			`json:"about"`			// Описание пользователя.
	Email string 			`json:"email"`			// Почтовый адрес пользователя (уникальное поле).
	FullName string 		`json:"fullname"`		// Полное имя пользователя.
	NickName string 		`json:"nickname"`		// Имя пользователя (уникальное поле). Данное поле допускает только латиницу, цифры и знак подчеркивания. Сравнение имени регистронезависимо.
}

type Vote struct {
	Nickname string 		`json:"nickname"`
	Voice int32 			`json:"voice"`
	Thread string 			`json:"-"`
}

type PostDetail struct {
	Author *User 			`json:"author"`
	Forum *Forum 			`json:"forum"`
	Post *Post 				`json:"post"`
	Thread *Thread			`json:"thread"`
}