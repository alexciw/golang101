package test

import (
	"encoding/json"
	"log"
	"testing"
)

type Person struct {
	Name  string
	Phone string
	Addr  string
}

type Book struct {
	Title   string
	Author  Person
	Pages   int            // 书的页数
	Indexes map[string]int // 书的索引
}

type BookAnonymous struct {
	Title string
	Person
	Pages   int            // 书的页数
	Indexes map[string]int // 书的索引
}

func TestStruct(t *testing.T) {
	book := Book{}
	b, _ := json.Marshal(book)
	log.Println("Book:", string(b))

	var bookNil Book
	bn, _ := json.Marshal(bookNil)
	log.Println("Book:", string(bn))

	bookAnonymous := BookAnonymous{}
	ba, _ := json.Marshal(bookAnonymous)
	log.Println("BookAnonymous:", string(ba))
}
