// Package rdb это пакет для decode/encode rdb файлов
//
// Описание формата http://rdb.fnordig.de/file_format.html
//
// Типы данных:
//   String
//   List
//   Set
//   Sorted Set
//   Hash
//   ZipMap
//   ZipList
//   IntSet
//   Sorted Set in Ziplist
//   HashMap in Ziplist (добавлен в RDB version 4)
//   List in QuickList (добавлен в RDB version 7)
package rdb
