module fasttrack

go 1.18

replace github.com/mattn/go-sqlite3 v1.14.16 => github.com/jgiannuzzi/go-sqlite3 v1.14.17-0.20221110191122-34e1e855329d

require (
	github.com/google/uuid v1.3.0
	github.com/jackc/pgconn v1.13.0
	github.com/sirupsen/logrus v1.9.0
	gorm.io/driver/postgres v1.4.5
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.24.1
	gorm.io/plugin/dbresolver v1.3.0
)

require (
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	golang.org/x/crypto v0.2.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
)
