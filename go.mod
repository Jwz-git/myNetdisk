module personal-disk

go 1.25.6

require github.com/go-sql-driver/mysql v1.7.1

require filippo.io/edwards25519 v1.1.0 // indirect

replace github.com/go-sql-driver/mysql => ./third_party/go_mysql
