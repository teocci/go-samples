// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-19
package main

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	gopg "github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/go-pg/pg/v10/types"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgconn/stmtcache"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/teocci/go-samples/psql-samples/raw"
)

const jinanBenchConfig = "user=jinan password=jinan#db host=localhost port=5432 dbname=test_db sslmode=disable"

var (
	setupOnce     sync.Once
	pgxPool       *pgxpool.Pool
	pgxStdlib     *sql.DB
	pq            *sql.DB
	pg            *gopg.DB
	pgConn        *pgconn.PgConn
	randPersonIDs []int32
)

var (
	selectPersonNameSQL = `select first_name from Person where id=$1`

	selectPersonNameSQLQuestionMark = `select first_name from Person where id=?`

	selectPersonSQL = `
select id, first_name, last_name, sex, birth_date, weight, height, update_time
from Person
where id=$1`
	selectPersonSQLQuestionMark = `
select id, first_name, last_name, sex, birth_date, weight, height, update_time
from Person
where id=?`

	selectMultiplePeopleSQL = `
select id, first_name, last_name, sex, birth_date, weight, height, update_time
from Person
where id between $1 and $1 + 24`
	selectMultiplePeopleSQLQuestionMark = `
select id, first_name, last_name, sex, birth_date, weight, height, update_time
from Person
where id between ? and ? + 24`

	selectLargeTextSQL = `select repeat('*', $1)`
)

var (
	rawSelectPersonNameStmt     *raw.PreparedStatement
	rawSelectPersonStmt         *raw.PreparedStatement
	rawSelectMultiplePeopleStmt *raw.PreparedStatement
)

var rxBuf []byte

type Person struct {
	Id         int32
	FirstName  string
	LastName   string
	Sex        string
	BirthDate  time.Time
	Weight     int32
	Height     int32
	UpdateTime time.Time
}

type PersonBytes struct {
	Id         int32
	FirstName  []byte
	LastName   []byte
	Sex        []byte
	BirthDate  time.Time
	Weight     int32
	Height     int32
	UpdateTime time.Time
}

// Implements pg.ColumnScanner.
var _ orm.ColumnScanner = (*Person)(nil)

func (p *Person) ScanColumn(col types.ColumnInfo, rd types.Reader, n int) error {
	tmp, err := rd.ReadFullTemp()
	if err != nil {
		return err
	}

	var n64 int64

	switch col.Name {
	case "id":
		n64, err = strconv.ParseInt(string(tmp), 10, 64)
		p.Id = int32(n64)
	case "first_name":
		p.FirstName = string(tmp)
	case "last_name":
		p.LastName = string(tmp)
	case "sex":
		p.Sex = string(tmp)
	case "birth_date":
		p.BirthDate, err = types.ScanTime(rd, n)
	case "weight":
		n64, err = strconv.ParseInt(string(tmp), 10, 64)
		p.Weight = int32(n64)
	case "height":
		n64, err = strconv.ParseInt(string(tmp), 10, 64)
		p.Weight = int32(n64)
	case "update_time":
		p.UpdateTime, err = types.ScanTime(rd, n)
	default:
		panic(fmt.Sprintf("unsupported column: %s", col.Name))
	}

	return err
}

func setup(b *testing.B) {
	setupOnce.Do(func() {
		config, err := pgxpool.ParseConfig(jinanBenchConfig)
		if err != nil {
			b.Fatalf("extractConfig failed: %v", err)
		}

		config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Prepare(ctx, "selectPersonName", selectPersonNameSQL)
			if err != nil {
				return err
			}

			_, err = conn.Prepare(ctx, "selectPerson", selectPersonSQL)
			if err != nil {
				return err
			}

			_, err = conn.Prepare(ctx, "selectMultiplePeople", selectMultiplePeopleSQL)
			if err != nil {
				return err
			}

			_, err = conn.Prepare(ctx, "selectLargeText", selectLargeTextSQL)
			if err != nil {
				return err
			}

			return nil
		}

		err = loadTestData(config.ConnConfig)
		if err != nil {
			b.Fatalf("loadTestData failed: %v", err)
		}

		pgxPool, err = openPgxNative(config)
		if err != nil {
			b.Fatalf("openPgxNative failed: %v", err)
		}

		pgxStdlib, err = openPgxStdlib(config)
		if err != nil {
			b.Fatalf("openPgxNative failed: %v", err)
		}

		pq, err = openPq(config.ConnConfig)
		if err != nil {
			b.Fatalf("openPq failed: %v", err)
		}

		pg, err = openPg(*config.ConnConfig)
		if err != nil {
			b.Fatalf("openPg failed: %v", err)
		}

		pgConn, err = pgconn.Connect(context.Background(), jinanBenchConfig)
		if err != nil {
			b.Fatalf("pgconn.Connect() failed: %v", err)
		}
		_, err = pgConn.Prepare(context.Background(), "selectPerson", selectPersonSQL, nil)
		if err != nil {
			b.Fatalf("pgConn.Prepare() failed: %v", err)
		}
		_, err = pgConn.Prepare(context.Background(), "selectMultiplePeople", selectMultiplePeopleSQL, nil)
		if err != nil {
			b.Fatalf("pgConn.Prepare() failed: %v", err)
		}

		rxBuf = make([]byte, 16384)

		// Get random Person ids in random order outside of timing
		rows, _ := pgxPool.Query(context.Background(), "select id from Person order by random()")
		for rows.Next() {
			var id int32
			rows.Scan(&id)
			randPersonIDs = append(randPersonIDs, id)
		}

		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", err)
		}
	})
}

func BenchmarkPgxNativeSelectSingleShortString(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		var firstName string
		err := pgxPool.QueryRow(context.Background(), "selectPersonName", id).Scan(&firstName)
		if err != nil {
			b.Fatalf("pgxPool.QueryRow Scan failed: %v", err)
		}
		if len(firstName) == 0 {
			b.Fatal("FirstName was empty")
		}
	}
}

func BenchmarkPgxStdlibSelectSingleShortString(b *testing.B) {
	setup(b)
	stmt, err := pgxStdlib.Prepare(selectPersonNameSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectSingleShortString(b, stmt)
}

func BenchmarkPgSelectSingleShortString(b *testing.B) {
	setup(b)

	stmt, err := pg.Prepare(selectPersonNameSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var firstName string
		id := randPersonIDs[i%len(randPersonIDs)]
		_, err := stmt.QueryOne(gopg.Scan(&firstName), id)
		if err != nil {
			b.Fatalf("stmt.QueryOne failed: %v", err)
		}
		if len(firstName) == 0 {
			b.Fatal("FirstName was empty")
		}
	}
}

func BenchmarkPqSelectSingleShortString(b *testing.B) {
	setup(b)
	stmt, err := pq.Prepare(selectPersonNameSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectSingleShortString(b, stmt)
}

func benchmarkSelectSingleShortString(b *testing.B, stmt *sql.Stmt) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		row := stmt.QueryRow(id)
		var firstName string
		err := row.Scan(&firstName)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}
		if len(firstName) == 0 {
			b.Fatal("FirstName was empty")
		}
	}
}

func BenchmarkPgxNativeSelectSingleShortBytes(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		var firstName []byte
		err := pgxPool.QueryRow(context.Background(), "selectPersonName", id).Scan(&firstName)
		if err != nil {
			b.Fatalf("pgxPool.QueryRow Scan failed: %v", err)
		}
		if len(firstName) == 0 {
			b.Fatal("FirstName was empty")
		}
	}
}

func BenchmarkPgxStdlibSelectSingleShortBytes(b *testing.B) {
	setup(b)
	stmt, err := pgxStdlib.Prepare(selectPersonNameSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectSingleShortBytes(b, stmt)
}

func BenchmarkPqSelectSingleShortBytes(b *testing.B) {
	setup(b)
	stmt, err := pq.Prepare(selectPersonNameSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectSingleShortBytes(b, stmt)
}

func benchmarkSelectSingleShortBytes(b *testing.B, stmt *sql.Stmt) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		row := stmt.QueryRow(id)
		var firstName []byte
		err := row.Scan(&firstName)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}
		if len(firstName) == 0 {
			b.Fatal("FirstName was empty")
		}
	}
}

func checkPersonWasFilled(b *testing.B, p Person) {
	if p.Id == 0 {
		b.Fatal("id was 0")
	}
	if len(p.FirstName) == 0 {
		b.Fatal("FirstName was empty")
	}
	if len(p.LastName) == 0 {
		b.Fatal("LastName was empty")
	}
	if len(p.Sex) == 0 {
		b.Fatal("Sex was empty")
	}
	var zeroTime time.Time
	if p.BirthDate == zeroTime {
		b.Fatal("BirthDate was zero time")
	}
	if p.Weight == 0 {
		b.Fatal("Weight was 0")
	}
	if p.Height == 0 {
		b.Fatal("Height was 0")
	}
	if p.UpdateTime == zeroTime {
		b.Fatal("UpdateTime was zero time")
	}
}

func BenchmarkPgxNativeSelectSingleRow(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := pgxPool.Query(context.Background(), "selectPerson", id)
		for rows.Next() {
			rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPgxNativeSelectSingleRowNotPreparedWithStatementCacheModePrepare(b *testing.B) {
	setup(b)

	config, err := pgxpool.ParseConfig(jinanBenchConfig)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}
	config.ConnConfig.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
		return stmtcache.New(conn, stmtcache.ModePrepare, 512)
	}

	db, err := openPgxNative(config)
	if err != nil {
		b.Fatalf("openPgxNative failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := db.Query(context.Background(), selectPersonSQL, id)
		for rows.Next() {
			rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPgxNativeSelectSingleRowNotPreparedWithStatementCacheModeDescribe(b *testing.B) {
	setup(b)

	config, err := pgxpool.ParseConfig(jinanBenchConfig)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}
	config.ConnConfig.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
		return stmtcache.New(conn, stmtcache.ModeDescribe, 512)
	}

	db, err := openPgxNative(config)
	if err != nil {
		b.Fatalf("openPgxNative failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := db.Query(context.Background(), selectPersonSQL, id)
		for rows.Next() {
			rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPgxNativeSelectSingleRowNotPreparedWithStatementCacheDisabled(b *testing.B) {
	setup(b)

	config, err := pgxpool.ParseConfig(jinanBenchConfig)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}
	config.ConnConfig.BuildStatementCache = nil

	db, err := openPgxNative(config)
	if err != nil {
		b.Fatalf("openPgxNative failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := db.Query(context.Background(), selectPersonSQL, id)
		for rows.Next() {
			rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPgconnSelectSingleRowTextProtocolNoParsing(b *testing.B) {
	setup(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		buf := []byte{0, 0, 0, 0}
		binary.BigEndian.PutUint32(buf, uint32(id))

		rr := pgConn.ExecPrepared(context.Background(), "selectPerson", [][]byte{buf}, []int16{1}, nil)
		_, err := rr.Close()
		if err != nil {
			b.Fatalf("pgConn.ExecPrepared failed: %v", err)
		}
	}
}

func BenchmarkPgconnSelectSingleRowBinaryProtocolNoParsing(b *testing.B) {
	setup(b)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		buf := []byte{0, 0, 0, 0}
		binary.BigEndian.PutUint32(buf, uint32(id))

		rr := pgConn.ExecPrepared(context.Background(), "selectPerson", [][]byte{buf}, []int16{1}, []int16{1})
		_, err := rr.Close()
		if err != nil {
			b.Fatalf("pgConn.ExecPrepared failed: %v", err)
		}
	}
}

func BenchmarkPgxStdlibSelectSingleRow(b *testing.B) {
	setup(b)
	stmt, err := pgxStdlib.Prepare(selectPersonSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectSingleRow(b, stmt)
}

func BenchmarkPgxStdlibSelectSingleRowNotPreparedStatementCacheModePrepare(b *testing.B) {
	setup(b)

	config, err := pgxpool.ParseConfig(jinanBenchConfig)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}
	config.ConnConfig.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
		return stmtcache.New(conn, stmtcache.ModePrepare, 512)
	}

	pgxStdlib, err = openPgxStdlib(config)
	if err != nil {
		b.Fatalf("openPgxStdlib failed: %v", err)
	}

	benchmarkSelectSingleRowNotPrepared(b, pgxStdlib, selectPersonSQL)
}

func BenchmarkPgxStdlibSelectSingleRowNotPreparedStatementCacheModeDescribe(b *testing.B) {
	setup(b)

	config, err := pgxpool.ParseConfig(jinanBenchConfig)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}
	config.ConnConfig.BuildStatementCache = func(conn *pgconn.PgConn) stmtcache.Cache {
		return stmtcache.New(conn, stmtcache.ModeDescribe, 512)
	}

	pgxStdlib, err = openPgxStdlib(config)
	if err != nil {
		b.Fatalf("openPgxStdlib failed: %v", err)
	}

	benchmarkSelectSingleRowNotPrepared(b, pgxStdlib, selectPersonSQL)
}

func BenchmarkPgxStdlibSelectSingleRowNotPreparedStatementCacheModeDisabled(b *testing.B) {
	setup(b)

	config, err := pgxpool.ParseConfig(jinanBenchConfig)
	if err != nil {
		b.Fatalf("ParseConfig failed: %v", err)
	}
	config.ConnConfig.BuildStatementCache = nil

	pgxStdlib, err = openPgxStdlib(config)
	if err != nil {
		b.Fatalf("openPgxStdlib failed: %v", err)
	}

	benchmarkSelectSingleRowNotPrepared(b, pgxStdlib, selectPersonSQL)
}

func BenchmarkPgSelectSingleRow(b *testing.B) {
	setup(b)

	stmt, err := pg.Prepare(selectPersonSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		id := randPersonIDs[i%len(randPersonIDs)]
		_, err := stmt.QueryOne(&p, id)
		if err != nil {
			b.Fatalf("stmt.QueryOne failed: %v", err)
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPgSelectSingleRowNotPrepared(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var p Person
		id := randPersonIDs[i%len(randPersonIDs)]
		_, err := pg.QueryOne(&p, selectPersonSQLQuestionMark, id)
		if err != nil {
			b.Fatalf("pg.QueryOne failed: %v", err)
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPqSelectSingleRow(b *testing.B) {
	setup(b)
	stmt, err := pq.Prepare(selectPersonSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectSingleRow(b, stmt)
}

func BenchmarkPqSelectSingleRowNotPrepared(b *testing.B) {
	setup(b)
	benchmarkSelectSingleRowNotPrepared(b, pq, selectPersonSQL)
}

func benchmarkSelectSingleRow(b *testing.B, stmt *sql.Stmt) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		row := stmt.QueryRow(id)
		var p Person
		err := row.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}

		checkPersonWasFilled(b, p)
	}
}

func benchmarkSelectSingleRowNotPrepared(b *testing.B, db *sql.DB, sql string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		row := db.QueryRow(sql, id)
		var p Person
		err := row.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}

		checkPersonWasFilled(b, p)
	}
}

func BenchmarkPgxNativeSelectMultipleRows(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := pgxPool.Query(context.Background(), "selectMultiplePeople", id)
		var p Person
		for rows.Next() {
			err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
			if err != nil {
				b.Fatalf("rows.Scan failed: %v", err)
			}
			checkPersonWasFilled(b, p)
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}
	}
}

func BenchmarkPgxNativeSelectMultipleRowsIntoGenericBinary(b *testing.B) {
	setup(b)

	type personRaw struct {
		Id         pgtype.GenericBinary
		FirstName  pgtype.GenericBinary
		LastName   pgtype.GenericBinary
		Sex        pgtype.GenericBinary
		BirthDate  pgtype.GenericBinary
		Weight     pgtype.GenericBinary
		Height     pgtype.GenericBinary
		UpdateTime pgtype.GenericBinary
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := pgxPool.Query(context.Background(), "selectMultiplePeople", id)
		var p personRaw
		for rows.Next() {
			err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
			if err != nil {
				b.Fatalf("rows.Scan failed: %v", err)
			}
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}
	}
}

func BenchmarkPgConnSelectMultipleRowsWithWithDecodeBinary(b *testing.B) {
	setup(b)

	type personRaw struct {
		Id         pgtype.Int4
		FirstName  pgtype.GenericBinary
		LastName   pgtype.GenericBinary
		Sex        pgtype.GenericBinary
		BirthDate  pgtype.Date
		Weight     pgtype.Int4
		Height     pgtype.Int4
		UpdateTime pgtype.Timestamptz
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		buf := []byte{0, 0, 0, 0}
		binary.BigEndian.PutUint32(buf, uint32(id))

		rr := pgConn.ExecPrepared(context.Background(), "selectMultiplePeople", [][]byte{buf}, []int16{1}, []int16{1})

		var p personRaw
		for rr.NextRow() {
			var err error
			vi := 0

			err = p.Id.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.FirstName.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.LastName.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Sex.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.BirthDate.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Weight.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Height.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.UpdateTime.DecodeBinary(nil, rr.Values()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}

		}

		_, err := rr.Close()
		if err != nil {
			b.Fatalf("pgConn.ExecPrepared failed: %v", err)
		}
	}
}

func BenchmarkPgxNativeSelectMultipleRowsWithoutScan(b *testing.B) {
	setup(b)

	type personRaw struct {
		Id         pgtype.Int4
		FirstName  pgtype.GenericBinary
		LastName   pgtype.GenericBinary
		Sex        pgtype.GenericBinary
		BirthDate  pgtype.Date
		Weight     pgtype.Int4
		Height     pgtype.Int4
		UpdateTime pgtype.Timestamptz
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := pgxPool.Query(context.Background(), "selectMultiplePeople", id)
		var p personRaw
		for rows.Next() {
			var err error
			vi := 0

			err = p.Id.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.FirstName.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.LastName.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Sex.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.BirthDate.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Weight.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Height.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.UpdateTime.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}
	}
}

func BenchmarkPgxNativeSelectMultipleRowsIntoGenericBinaryWithoutScan(b *testing.B) {
	setup(b)

	type personRaw struct {
		Id         pgtype.GenericBinary
		FirstName  pgtype.GenericBinary
		LastName   pgtype.GenericBinary
		Sex        pgtype.GenericBinary
		BirthDate  pgtype.GenericBinary
		Weight     pgtype.GenericBinary
		Height     pgtype.GenericBinary
		UpdateTime pgtype.GenericBinary
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := pgxPool.Query(context.Background(), "selectMultiplePeople", id)
		var p personRaw
		for rows.Next() {
			var err error
			vi := 0

			err = p.Id.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.FirstName.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.LastName.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Sex.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.BirthDate.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Weight.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.Height.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
			vi += 1

			err = p.UpdateTime.DecodeBinary(nil, rows.RawValues()[vi])
			if err != nil {
				b.Fatalf("DecodeBinary failed: %v", err)
			}
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}
	}
}

func BenchmarkPgxStdlibSelectMultipleRows(b *testing.B) {
	setup(b)

	stmt, err := pgxStdlib.Prepare(selectMultiplePeopleSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectMultipleRows(b, stmt)
}

// This benchmark is different from the other multiple rows in that it collects
// all rows whereas the others process and discard. So it is not apples-to-
// apples for *SelectMultipleRows*.
func BenchmarkPgSelectMultipleRowsCollect(b *testing.B) {
	setup(b)

	stmt, err := pg.Prepare(selectMultiplePeopleSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var people []Person
		id := randPersonIDs[i%len(randPersonIDs)]
		_, err := stmt.Query(&people, id)
		if err != nil {
			b.Fatalf("stmt.Query failed: %v", err)
		}

		for i := range people {
			checkPersonWasFilled(b, people[i])
		}
	}
}

func BenchmarkPgSelectMultipleRowsAndDiscard(b *testing.B) {
	setup(b)

	stmt, err := pg.Prepare(selectMultiplePeopleSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		_, err := stmt.Query(gopg.Discard, id)
		if err != nil {
			b.Fatalf("stmt.Query failed: %v", err)
		}
	}
}

func BenchmarkPqSelectMultipleRows(b *testing.B) {
	setup(b)

	stmt, err := pq.Prepare(selectMultiplePeopleSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectMultipleRows(b, stmt)
}

func benchmarkSelectMultipleRows(b *testing.B, stmt *sql.Stmt) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		rows, err := stmt.Query(id)
		if err != nil {
			b.Fatalf("db.Query failed: %v", err)
		}

		var p Person
		for rows.Next() {
			err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
			if err != nil {
				b.Fatalf("rows.Scan failed: %v", err)
			}
			checkPersonWasFilled(b, p)
		}

		if rows.Err() != nil {
			b.Fatalf("rows.Err() returned an error: %v", err)
		}
	}
}

func checkPersonBytesWasFilled(b *testing.B, p PersonBytes) {
	if p.Id == 0 {
		b.Fatal("id was 0")
	}
	if len(p.FirstName) == 0 {
		b.Fatal("FirstName was empty")
	}
	if len(p.LastName) == 0 {
		b.Fatal("LastName was empty")
	}
	if len(p.Sex) == 0 {
		b.Fatal("Sex was empty")
	}
	var zeroTime time.Time
	if p.BirthDate == zeroTime {
		b.Fatal("BirthDate was zero time")
	}
	if p.Weight == 0 {
		b.Fatal("Weight was 0")
	}
	if p.Height == 0 {
		b.Fatal("Height was 0")
	}
	if p.UpdateTime == zeroTime {
		b.Fatal("UpdateTime was zero time")
	}
}

func BenchmarkPgxNativeSelectMultipleRowsBytes(b *testing.B) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]

		rows, _ := pgxPool.Query(context.Background(), "selectMultiplePeople", id)
		var p PersonBytes
		for rows.Next() {
			err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
			if err != nil {
				b.Fatalf("rows.Scan failed: %v", err)
			}
			checkPersonBytesWasFilled(b, p)
		}
		if rows.Err() != nil {
			b.Fatalf("pgxPool.Query failed: %v", rows.Err())
		}
	}
}

func BenchmarkPgxStdlibSelectMultipleRowsBytes(b *testing.B) {
	setup(b)

	stmt, err := pgxStdlib.Prepare(selectMultiplePeopleSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectMultipleRowsBytes(b, stmt)
}

func BenchmarkPqSelectMultipleRowsBytes(b *testing.B) {
	setup(b)

	stmt, err := pq.Prepare(selectMultiplePeopleSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectMultipleRowsBytes(b, stmt)
}

func benchmarkSelectMultipleRowsBytes(b *testing.B, stmt *sql.Stmt) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := randPersonIDs[i%len(randPersonIDs)]
		rows, err := stmt.Query(id)
		if err != nil {
			b.Fatalf("db.Query failed: %v", err)
		}

		var p PersonBytes
		for rows.Next() {
			err := rows.Scan(&p.Id, &p.FirstName, &p.LastName, &p.Sex, &p.BirthDate, &p.Weight, &p.Height, &p.UpdateTime)
			if err != nil {
				b.Fatalf("rows.Scan failed: %v", err)
			}
			checkPersonBytesWasFilled(b, p)
		}

		if rows.Err() != nil {
			b.Fatalf("rows.Err() returned an error: %v", err)
		}
	}
}

func BenchmarkPgxNativeSelectBatch3Query(b *testing.B) {
	setup(b)

	b.ResetTimer()
	batch := &pgx.Batch{}
	results := make([]string, 3)
	for j := range results {
		batch.Queue("selectLargeText", j)
	}

	for i := 0; i < b.N; i++ {
		br := pgxPool.SendBatch(context.Background(), batch)

		for j := range results {
			if err := br.QueryRow().Scan(&results[j]); err != nil {
				b.Fatal(err)
			}
		}

		if err := br.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPgxNativeSelectNoBatch3Query(b *testing.B) {
	setup(b)

	b.ResetTimer()
	results := make([]string, 3)
	for i := 0; i < b.N; i++ {
		for j := range results {
			if err := pgxPool.QueryRow(context.Background(), "selectLargeText", j).Scan(&results[j]); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkPgxStdlibSelectNoBatch3Query(b *testing.B) {
	setup(b)
	stmt, err := pgxStdlib.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectNoBatch3Query(b, stmt)
}

func BenchmarkPqSelectNoBatch3Query(b *testing.B) {
	setup(b)
	stmt, err := pq.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectNoBatch3Query(b, stmt)
}

func benchmarkSelectNoBatch3Query(b *testing.B, stmt *sql.Stmt) {
	b.ResetTimer()
	results := make([]string, 3)
	for i := 0; i < b.N; i++ {
		for j := range results {
			if err := stmt.QueryRow(j).Scan(&results[j]); err != nil {
				b.Fatal(err)
			}
		}
	}
}


func BenchmarkSelectLargeTextString(b *testing.B) {
	fixture := []struct {
		desc    string
		records int
	}{
		{
			desc:    "1 KB",
			records: 1024,
		},
		{
			desc:    "4 KB",
			records: 4*1024,
		},
		{
			desc:    "8 KB",
			records: 8*1024,
		},
		{
			desc:    "64 KB",
			records: 64*1024,
		},
		{
			desc:    "512 MB",
			records: 512*1024,
		},
		{
			desc:    "4 MB",
			records: 4*1024*1024,
		},
	}

	benchs := []struct {
		desc string
		fn   func(b *testing.B, size int)
	}{
		{
			desc: "PQ",
			fn: func(b *testing.B, size int) {
				benchmarkPqSelectLargeTextString(b, size)
			},
		},
		{
			desc: "GO-PG",
			fn: func(b *testing.B, size int) {
				benchmarkPgSelectLargeTextString(b, size)
			},
		},
		{
			desc: "PGX Stdlib",
			fn: func(b *testing.B, size int) {
				benchmarkPgxStdlibSelectLargeTextString(b, size)
			},
		},
		{
			desc: "PGX Native",
			fn: func(b *testing.B, size int) {
				benchmarkPgxNativeSelectLargeTextString(b, size)
			},
		},
	}

	for _, bench := range benchs {
		b.Run(bench.desc, func(b *testing.B) {
			for _, f := range fixture {
				b.Run(f.desc, func(b *testing.B) {
					//for i := 0; i < b.N; i++ {
					bench.fn(b, f.records)
					//}
				})
			}
		})
	}
}

func benchmarkPqSelectLargeTextString(b *testing.B, size int) {
	setup(b)
	stmt, err := pq.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectLargeTextString(b, stmt, size)
}

func benchmarkSelectLargeTextString(b *testing.B, stmt *sql.Stmt, size int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s string
		err := stmt.QueryRow(size).Scan(&s)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}
		if len(s) != size {
			b.Fatalf("expected length %v, got %v", size, len(s))
		}
	}
}

func benchmarkPgxStdlibSelectLargeTextString(b *testing.B, size int) {
	setup(b)
	stmt, err := pgxStdlib.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectLargeTextString(b, stmt, size)
}
func benchmarkPgxNativeSelectLargeTextString(b *testing.B, size int) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s string
		err := pgxPool.QueryRow(context.Background(), "selectLargeText", size).Scan(&s)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}
		if len(s) != size {
			b.Fatalf("expected length %v, got %v", size, len(s))
		}
	}
}


func BenchmarkSelectLargeTextBytes(b *testing.B) {
	fixture := []struct {
		desc    string
		records int
	}{
		{
			desc:    "1 KB",
			records: 1024,
		},
		{
			desc:    "4 KB",
			records: 4*1024,
		},
		{
			desc:    "8 KB",
			records: 8*1024,
		},
		{
			desc:    "64 KB",
			records: 64*1024,
		},
		{
			desc:    "512 MB",
			records: 512*1024,
		},
		{
			desc:    "4 MB",
			records: 4*1024*1024,
		},
	}

	benchs := []struct {
		desc string
		fn   func(b *testing.B, size int)
	}{
		{
			desc: "PQ",
			fn: func(b *testing.B, size int) {
				benchmarkPqSelectLargeTextBytes(b, size)
			},
		},
		{
			desc: "PGX Stdlib",
			fn: func(b *testing.B, size int) {
				benchmarkPgxStdlibSelectLargeTextBytes(b, size)
			},
		},
		{
			desc: "PGX Native",
			fn: func(b *testing.B, size int) {
				benchmarkPgxNativeSelectLargeTextBytes(b, size)
			},
		},
		{
			desc: "GO-PG",
			fn: func(b *testing.B, size int) {
				benchmarkPgSelectLargeTextString(b, size)
			},
		},
	}

	for _, bench := range benchs {
		b.Run(bench.desc, func(b *testing.B) {
			for _, f := range fixture {
				b.Run(f.desc, func(b *testing.B) {
					//for i := 0; i < b.N; i++ {
						bench.fn(b, f.records)
					//}
				})
			}
		})
	}
}

func benchmarkPqSelectLargeTextBytes(b *testing.B, size int) {
	setup(b)
	stmt, err := pq.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectLargeTextBytes(b, stmt, size)
}

func benchmarkPgxStdlibSelectLargeTextBytes(b *testing.B, size int) {
	setup(b)
	stmt, err := pgxStdlib.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	benchmarkSelectLargeTextBytes(b, stmt, size)
}

func benchmarkSelectLargeTextBytes(b *testing.B, stmt *sql.Stmt, size int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s []byte
		err := stmt.QueryRow(size).Scan(&s)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}
		if len(s) != size {
			b.Fatalf("expected length %v, got %v", size, len(s))
		}
	}
}

func benchmarkPgxNativeSelectLargeTextBytes(b *testing.B, size int) {
	setup(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s []byte
		err := pgxPool.QueryRow(context.Background(), "selectLargeText", size).Scan(&s)
		if err != nil {
			b.Fatalf("row.Scan failed: %v", err)
		}
		if len(s) != size {
			b.Fatalf("expected length %v, got %v", size, len(s))
		}
	}
}

func benchmarkPgSelectLargeTextString(b *testing.B, size int) {
	setup(b)

	stmt, err := pg.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s string
		_, err := stmt.QueryOne(gopg.Scan(&s), size)
		if err != nil {
			b.Fatalf("stmt.QueryOne failed: %v", err)
		}
		if len(s) != size {
			b.Fatalf("expected length %v, got %v", size, len(s))
		}
	}
}

func benchmarkPgSelectLargeTextStringStruct(b *testing.B, size int) {
	setup(b)

	stmt, err := pg.Prepare(selectLargeTextSQL)
	if err != nil {
		b.Fatalf("Prepare failed: %v", err)
	}
	defer stmt.Close()

	type stringLoader struct {
		Cad string
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var dst stringLoader
		_, err := stmt.QueryOne(&dst, size)
		if err != nil {
			b.Fatalf("stmt.QueryOne failed: %v", err)
		}
		if len(dst.Cad) != size {
			b.Fatalf("expected length %v, got %v", size, len(dst.Cad))
		}
	}
}
