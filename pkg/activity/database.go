package activity

import (
	"coding-challenge/pkg/db"
	"coding-challenge/pkg/model"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const SaveToDatabaseActivityTimeout = time.Second

type PostgreSqlConnection struct {
	Host   string
	Port   int
	User   string
	Pass   string
	DbName string
}

type PostgreSqlActivityHost struct {
	db db.BillDatabase
}

var _ ActivityHost = &PostgreSqlActivityHost{}

func NewPostgreSqlActivityHost(conn PostgreSqlConnection) (*PostgreSqlActivityHost, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		conn.Host, conn.Port, conn.User, conn.Pass, conn.DbName)
	sql, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	return &PostgreSqlActivityHost{db: db.NewSqlBillDatabase(sql)}, nil
}

func (a *PostgreSqlActivityHost) CreateBillIfNotExistActivity(bill model.BillInfo) (uint64, error) {
	return a.db.CreateBill(bill)
}

func (a *PostgreSqlActivityHost) AddBillLineItemIfNotExistActivity(lineItem model.BillLineItem, totalBefore model.TotalAmount) (uint64, error) {
	return a.db.AddLineItem(lineItem, totalBefore)
}

func (a *PostgreSqlActivityHost) CloseBillActivity(bill model.BillInfo) (uint64, error) {
	return a.db.CloseBill(bill.Id)
}
